 Justified gaps — defensive panics (unreachable with current Go reflection):                                                                          
  ┌──────────────────────────┬─────────────────────────────────────┬─────────────────────────────────────────────────────────────┬────────────────────┐
  │           File           │                Lines                │                            Guard                            │   Justification    │
  ├──────────────────────────┼─────────────────────────────────────┼─────────────────────────────────────────────────────────────┼────────────────────┤
  │ compare.go:502-504       │ convertReflectValue panic           │ Reflect Convert() always yields target type                 │ Volatile API →     │
  │                          │                                     │                                                             │ panic              │
  ├──────────────────────────┼─────────────────────────────────────┼─────────────────────────────────────────────────────────────┼────────────────────┤
  │ object.go:57-68          │ CanConvert/CanInterface panics      │ Already converted to panic in this session                  │ Volatile API →     │
  │                          │                                     │                                                             │ panic              │
  ├──────────────────────────┼─────────────────────────────────────┼─────────────────────────────────────────────────────────────┼────────────────────┤
  │ object.go:82-93          │ numeric CanConvert/CanInterface     │ Already converted to panic in this session                  │ Volatile API →     │
  │                          │ panics                              │                                                             │ panic              │
  ├──────────────────────────┼─────────────────────────────────────┼─────────────────────────────────────────────────────────────┼────────────────────┤
  │ equal.go:390-392         │ copyExportedFields CanInterface     │ Array/slice elements from reflect.ValueOf() are always      │ Volatile API →     │
  │                          │ panic                               │ interfaceable                                               │ panic              │
  ├──────────────────────────┼─────────────────────────────────────┼─────────────────────────────────────────────────────────────┼────────────────────┤
  │ order.go:307-309,319-320 │ isStrictlyOrdered CanInterface      │ Same reasoning — slice elements are always interfaceable    │ Volatile API →     │
  │                          │ panics                              │                                                             │ panic              │
  └──────────────────────────┴─────────────────────────────────────┴─────────────────────────────────────────────────────────────┴────────────────────┘
  Justified gaps — platform/environment dependent:                                                                                                     
  ┌─────────────────────────┬────────────────────────────────────────────────┬────────────────────────────────┬────────────────────────────┐           
  │          File           │                     Lines                      │             Guard              │       Justification        │           
  ├─────────────────────────┼────────────────────────────────────────────────┼────────────────────────────────┼────────────────────────────┤           
  │ file.go:155-157,196-198 │ isSymlink error path in FileEmpty/FileNotEmpty │ Reachable on Windows, deferred │ Platform-dependent → defer │           
  ├─────────────────────────┼────────────────────────────────────────────────┼────────────────────────────────┼────────────────────────────┤           
  │ file.go:229-232         │ isSymlink os.Readlink error                    │ Reachable on Windows, deferred │ Platform-dependent → defer │           
  └─────────────────────────┴────────────────────────────────────────────────┴────────────────────────────────┴────────────────────────────┘           
  Justified gaps — only exercised via cross-package tests (assert/require):                                                                            
  File: compare.go:517-574                                                                                                                             
  Lines: compareOrderedWithAny                                                                                                                         
  Function: Called only via generic Greater/Less etc from assert/require                                                                               
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: object.go:38-96                                                                                                                                
  Lines: ObjectsAreEqualValues                                                                                                                         
  Function: Called from EqualValues in assert/require                                                                                                  
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: order.go:293-335                                                                                                                               
  Lines: isStrictlyOrdered (body)                                                                                                                      
  Function: Called via assert/require ordering assertions                                                                                              
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: equal.go:349-401                                                                                                                               
  Lines: copyExportedFields                                                                                                                            
  Function: Called from EqualExportedValues in assert/require                                                                                          
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: testing.go:83-91                                                                                                                               
  Lines: isTest                                                                                                                                        
  Function: Called from callerInfo which is only called via cross-package                                                                              
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: testing.go:113-173                                                                                                                             
  Lines: callerInfo                                                                                                                                    
  Function: Only reached via Fail → errorWithCallerInfo from assert/require                                                                            
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: testing.go:101-110                                                                                                                             
  Lines: errorWithCallerInfo namer/message blocks                                                                                                      
  Function: Same — cross-package only                                                                                                                  
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: condition.go:305-478                                                                                                                           
  Lines: pollAtTickFunc/executeCondition                                                                                                               
  Function: Concurrent polling internals — covered by Eventually/Never tests in assert/require                                                         
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: condition.go:503                                                                                                                               
  Lines: CollectT.Helper                                                                                                                               
  Function: Empty function, 0% is a no-op body                                                                                                         
  Reason:                                                                                                                                              
  ────────────────────────────────────────                                                                                                             
  File: compare.go:496-508                                                                                                                             
  Lines: convertReflectValue (non-panic path)                                                                                                          
  Function: Generic helper, called from assert/require                                                                                                 
  Reason:                                                                                                                                              
  Justified gaps — not yet implemented:                                                                                                                
  ┌─────────────────┬─────────────────┬─────────────────────────────────────────────┬────────┐                                                         
  │      File       │      Lines      │                  Function                   │ Reason │                                                         
  ├─────────────────┼─────────────────┼─────────────────────────────────────────────┼────────┤                                                         
  │ safety.go:27-39 │ NoGoRoutineLeak │ Marked with TODO, not yet fully implemented │        │                                                         
  └─────────────────┴─────────────────┴─────────────────────────────────────────────┴────────┘                                                         
  Summary: All gaps are justified:                                                                                                                     
  - Defensive panics on volatile reflect APIs (by design, untestable)                                                                                  
  - Platform-dependent code deferred (Windows symlink semantics)                                                                                       
  - Cross-package only functions that show 0% in the single-package profile but are covered by assert/require tests (the go tool cover -func on        
  cover.out confirmed these at 83-97%)                                                                                                                 
  - One stub (NoGoRoutineLeak) not yet implemented       
