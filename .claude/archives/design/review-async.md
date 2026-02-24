Now I have a complete picture. Here's my review of the implementation:

  ---
  Review of Async Assertions with Flexible Condition Signatures

  1. Immediate Build Breakage: Comparison type removed

  The Comparison type was deleted from ifaces.go:35, but the generated files assert/assert_types.go:36 and require/require_types.go:36 still reference assertions.Comparison.
   This is the compiler error you're seeing. You need to either:
  - Keep Comparison (it's a separate concern from the condition changes), or
  - Re-run go generate ./... to update the generated files

  Since Condition() now takes func() bool directly instead of Comparison, removing Comparison is fine conceptually — but it's a breaking API change for users who reference
  assert.Comparison or require.Comparison in their code. Worth deciding whether to keep it as a deprecated alias.

  2. The select-before-call pattern is a no-op guard

  Throughout makeCondition and makeCollectibleCondition, every branch does:

  select {
  case <-ctx.Done():
      return false
  default:
      return typed()  // or typed(ctx)
  }

  This check is not protective — the context could be cancelled between the select and the actual function call. For func() bool variants (that don't take a context), this
  is the best you can do, so it's fine there. But for func(context.Context) bool and func(context.Context) error, the condition itself already receives ctx and should handle
   cancellation internally. The outer select adds a false sense of safety. It's not harmful, just redundant for the context-aware variants.

  3. The func() error signature loses error information

  When a func() error condition fails, the error is silently swallowed:

  case func() error:
      return func(ctx context.Context) bool {
          // ...
          err := typed()
          return err == nil
      }

  The actual error message is discarded. For Eventually, when the condition times out, the user only sees "condition never satisfied" with no indication of why the condition
   kept failing. Consider capturing the last error and surfacing it in the failure message, similar to how eventuallyWithT captures lastCollectedErrors.

  4. CollectibleConditioner — the func(*CollectT) bool variant is awkward

  The func(*CollectT) bool case adds a synthetic collector.Errorf("returned false") when false is returned. This is a confusing UX: the user uses CollectT to report detailed
   errors, but also returns a boolean? These are competing error-reporting mechanisms. If the user calls True(c, someCondition) (which reports via CollectT) and also returns
   false, they get a duplicate error. I'd drop func(*CollectT) bool from the constraint — the CollectT path already has its own failure signaling.

  5. Never with context-aware conditions — semantic mismatch

  For Never, the condition returning true means failure. But func() error returning nil (no error = success) is mapped to true, which means Never fails. This is correct but
  counterintuitive: Never(t, func() error { return nil }, ...) fails because "no error" is treated as "condition satisfied". The semantics work but the mental model for
  Never + func() error is confusing. Worth documenting this explicitly.

  6. conditionChan buffered channel sizing

  conditionChan has buffer size 1 (make(chan func(context.Context) bool, 1)). The initial condition is pushed before the goroutines start (line 402: p.conditionChan <-
  condition). If the first execution hasn't consumed it yet and a tick fires, the send in pollAtTickFunc will block on the nested select. This is fine for correctness (no
  parallel execution), but the comment on line 66-68 ("The condition function is never executed in parallel") should probably also note that ticks can be skipped if the
  condition takes longer than the tick interval.

  7. makeCondition / makeCollectibleCondition — type-switch on any loses compile-time safety

  Converting a generic C Conditioner to any and type-switching defeats the purpose of having the constraint. If the constraint is somehow violated (which Go prevents at
  compile time), the default branch returns nil, which would panic on first poll. This is a minor concern since the constraint makes default truly unreachable, but a
  panic("unreachable") would be clearer than a nil return.

  8. Code generation implications

  The big question: can the codegen handle generic type parameters on these functions? The scanner parses signatures with go/types, but Eventually[C Conditioner] is a new
  pattern for this codebase. Have you verified that go generate ./... produces correct output for the 8 variants? The generated assert.Eventually and require.Eventually need
   to propagate the type parameter correctly. This seems like the highest-risk area.

  9. Minor: func(context.Context) bool in Conditioner — missing from doc comment

  The doc for Eventually lists func(context.Context) (no return) as a supported signature, but Conditioner doesn't include func(context.Context) (void). The doc says:

  func(context.Context), func(context.Context) error

  But the constraint only has func(context.Context) bool and func(context.Context) error. The doc and the constraint disagree.

  Summary

  ┌──────────────────────────────────────────────────────────────────────┬───────────────────┐
  │                                Issue                                 │     Severity      │
  ├──────────────────────────────────────────────────────────────────────┼───────────────────┤
  │ Comparison type removed, generated code broken                       │ Build-breaking    │
  ├──────────────────────────────────────────────────────────────────────┼───────────────────┤
  │ func() error discards error details                                  │ Functional gap    │
  ├──────────────────────────────────────────────────────────────────────┼───────────────────┤
  │ Doc mentions func(context.Context) but constraint doesn't include it │ Doc/code mismatch │
  ├──────────────────────────────────────────────────────────────────────┼───────────────────┤
  │ Codegen handling of generic type params untested                     │ Risk              │
  ├──────────────────────────────────────────────────────────────────────┼───────────────────┤
  │ func(*CollectT) bool is a confusing variant                          │ Design smell      │
  ├──────────────────────────────────────────────────────────────────────┼───────────────────┤
  │ select before ctx-aware calls is redundant                           │ Minor             │
  ├──────────────────────────────────────────────────────────────────────┼───────────────────┤
  │ default: return nil should be panic                                  │ Minor             │
  ├──────────────────────────────────────────────────────────────────────┼───────────────────┤
  │ Never + func() error semantics need docs                             │ Minor             │
  └──────────────────────────────────────────────────────────────────────┴───────────────────┘

  The core architecture (constraint unions, makeCondition adapter, unified poller) is sound. The main things to address before merging are the build break, the lost error
  information, and verifying codegen handles the generics.

I share most of the points you've raised. I needed some confirmation/guidance.
  When I am asking lik that, it's already that I feel like it didn't do the right thing...

  1. Removed type. I should do this in another commit, it's indeed unrelated.
     Breaking change indeed. Last chance for use to rush in such to v2.4.0.
     I need to regenerate code.

     > OK. Regenerated code & doc

  2. Yes I am aware. Not sure yet what's the best approach here

     > OK. Removed the extraneous select case.

  3. Yes I am aware. It is not obvious to convey the error in this case. Unless I set the "pivot" function (result of the conversion)
     to be func(context.Context) error instead of func(context.Context) bool

     > OK. Pivoted to func(context.Context) error as the common denominator

  4. Yes. The thing is that I never really liked this CollecT thing. Now when trying to extend it, it is clear how awkward it is.
     I think I'll ditch the bool and error returned values and just add the option to add context. How about that?

     > OK ditched. Only 2 signatures for each kind.

  5. About "Never". Yeah right, this is misleading. I am adding some flag to support reversed logic

     > OK documented and implemented reverse logic for never when using error-returning function

  6. Chan. Yes the documentation is not super clear. "parallel" is not the right term: "serialized" is better. And yes, I should mention
     the issue with ticks.

     > OK reformulated docs. Also added caveat about flaky tests.

  7. Mmmh agree. The use of generics here is just to avoid expanding the API with many more variants. It's a poor man's substitue to overloading,
     which doesn't exist in go. Agree type safety is not there. Agree should panic (code assertion) rather than possibly leaking a nil.

    > Kept it. Reduced the number of supported overloading.

  8. Codegen. Should work since we do support generics already. Once the design is fixed in internal/assertions, I'll move to check the codegen.

     > OK. Regenerated code & doc

  9. godoc: yes need to tidy this up

    > Should be fine now.

  Overall good design challenge, Mr Opus. I am more and more inclined to think that this improvement (both you and mee - or was it Mr Sonnet? -
  were initially super enthousiastic about growing our support for async, but the UX is simply not great.

  About the next improvement for async (requires go1.25 minimum) I am less skeptical: executing all async assertions in the new testing/synctest
  bubble sandbox.

● Good instincts across the board. A few thoughts on the open items:

  Point 3 — pivoting to func(context.Context) error: This is appealing because it naturally threads the error through to the failure message. The poller can capture the last
   error and surface it: "condition never satisfied: last error: connection refused". The bool becomes a derived property (err == nil). Clean.

  Point 4 — simplifying CollectibleConditioner: Yes, just func(*CollectT) and func(context.Context, *CollectT) is the right call. The CollectT is the error-reporting
  mechanism — adding bool or error returns on top of it is two competing channels for the same information. Keep it simple: you collect errors, or you don't. Two variants,
  done.

  Point 5 — reversed logic flag for Never: Be careful this doesn't leak into the public API or the condition signatures. If it's internal to the poller (e.g., the poller
  interprets "condition returned error" as success for Never), that's fine. The key insight is that for Never with func() error, the user mental model should be: "this error
   should never stop happening" — which is unusual. Might be worth questioning whether Never even needs the func() error variant, or if func() bool is sufficient there.

  On the broader reflection: The enthusiasm trap is real — "we can support 10 condition signatures" doesn't mean we should. The combinatorial explosion (4 signatures x 3
  assertions x context/no-context) is exactly the kind of thing that looks elegant in the type system but confusing in practice. If func() bool covers 95% of use cases and
  func(context.Context) error covers the rest, that might be the sweet spot — just two variants in Conditioner instead of four.

  The synctest direction sounds much more promising — letting the runtime handle the scheduling instead of hand-rolling poller goroutines with channels is a fundamentally
  better abstraction level.


