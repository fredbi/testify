❯Good. Today I've been reflecting on the next round of features, which are more oriented "code safety" (perhaps a new          
domain for asserions).                                                                                                        

I'd like to onboard the very useful "no goroutine leak" check. This one is not hard to implement and several libraries        
already do it. I can easily do that. So (almost) a no-brainer.                                                                

Another useful one is the "no file descriptor leak" check.                                                                    
But that one is way more challenging. I found an existing implementation (https://github.com/thediveo/fdooze),                
but it's linux only and not very well implemented (well not by my coding standards at least).                                 
The principle is rather simple on unix (open /proc, find fd, then filter out network fd and other special stuff).             
So basically this the lsof command.                                                                                           
Implementing the same functionality for windows seems on the other hand super challenging.                                    
Are you aware of any existing implementation (possibly just prototypes) that could drive my inspiration?                      

Eventually, memory leaks. It is indeed possible to produce memory leaks in go by tricking the gc with things like             
sync.Pool.                                                                                                                    
But I don't think it is possible to check that without instrumenting the tested poolsa.                                       
PErhaps you've a different opinion on that part?                                                                              

Please advise. Just brainstorming, no coding now.                                                                             



●Excellent brainstorming topic! Let me share thoughts on each:                                                                 

1. Goroutine Leak Detection ✅ (Easy)                                                                                         

Gold standard: https://github.com/uber-go/goleak                                                                              

You could either:                                                                                                             
- Wrap goleak as a testify assertion                                                                                          
- Implement your own (it's ~200 LOC of runtime.Stack() parsing)                                                               

Key challenges:                                                                                                               
- Filtering expected goroutines (test runner, runtime, HTTP servers in tests)                                                 
- Timing: need to wait for goroutines to settle after test                                                                    
- False positives from background goroutines                                                                                  

API suggestion:                                                                                                               
assert.NoGoroutineLeak(t, func() {                                                                                            
// test code that shouldn't leak                                                                                          
}, IgnoreGoroutines("database/sql.(*DB).connectionOpener"))                                                                   

2. File Descriptor Leak Detection ⚠️ (Hard on Windows)                                                                        

Linux/macOS: Straightforward                                                                                                  
- Linux: /proc/self/fd                                                                                                        
- macOS: lsof -p $PID or libproc APIs                                                                                         

Windows: Very challenging. Here's what I found:                                                                               

Existing Windows Implementations:                                                                                             

1. https://github.com/elastic/go-windows (Most promising)                                                                     
- Has handle enumeration via Native API                                                                                       
- NtQuerySystemInformation with SystemHandleInformation                                                                       
- Production code from Elastic beats                                                                                          
- Look at ntdll.go and handle tracking                                                                                        

2. https://github.com/shirou/gopsutil                                                                                         
- Cross-platform process utilities                                                                                            
- Has Process.OpenFiles() that works on Windows                                                                               
- Uses WMI or native APIs depending on version                                                                                
- Not perfect but battle-tested                                                                                               

3. https://github.com/Microsoft/go-winio                                                                                      
- Windows I/O utilities from Microsoft                                                                                        
- Has some handle management code                                                                                             
- Focus on named pipes, but handle tracking exists                                                                            

4. Native API approach (complex):                                                                                             
// Pseudo-code - Windows Native API                                                                                           
// NtQuerySystemInformation(SystemHandleInformation, ...)                                                                     
// Filter by process ID                                                                                                       
// Filter by object type (File, Socket, etc.)                                                                                 
// Compare snapshots before/after test                                                                                        

Practical Recommendation:                                                                                                     

Phase 1: Unix-only                                                                                                            
- Ship Linux/macOS support first                                                                                              
- Document Windows limitation                                                                                                 
- Better than nothing!                                                                                                        

Phase 2: Windows handle count (coarse)                                                                                        
// Track total handle count, not individual handles                                                                           
before := getProcessHandleCount()                                                                                             
// run test                                                                                                                   
after := getProcessHandleCount()                                                                                              
assert.Equal(t, before, after, "handle leak detected")                                                                        
Uses GetProcessHandleCount() Windows API - simple but can't tell you which handles leaked.                                    

Phase 3: Full Windows support                                                                                                 
- Use gopsutil for cross-platform abstraction                                                                                 
- Or implement native API yourself                                                                                            
- Filter by handle type (file vs socket vs mutex)                                                                             

fdooze Issues You Mentioned:                                                                                                  

Looking at https://github.com/thediveo/fdooze:                                                                                
- Only Linux                                                                                                                  
- Parsing /proc/self/fd manually                                                                                              
- No filtering options                                                                                                        
- Not actively maintained                                                                                                     

You can do better by:                                                                                                         
- Using gopsutil for cross-platform                                                                                           
- Adding filter options (ignore network sockets, pipes, etc.)                                                                 
- Better error messages showing leaked FD details                                                                             

3. Memory Leak Detection 🤔 (Philosophical Challenge)                                                                         

You're right - this is hard. Here's why:                                                                                      

The Problem:                                                                                                                  

var globalCache = make(map[string]*HugeStruct)                                                                                

func TestSomething(t *testing.T) {                                                                                            
// Adds to globalCache                                                                                                    
// Is this a "leak" or "intended caching"?                                                                                
}                                                                                                                             

Go's GC works correctly - the issue is retained references, not leaks.                                                        

What's Possible:                                                                                                              

1. Heap Growth Detection (coarse):                                                                                            
func (a *Assertions) NoHeapGrowth(f func()) bool {                                                                            
runtime.GC()                                                                                                              
var before, after runtime.MemStats                                                                                        
runtime.ReadMemStats(&before)                                                                                             

f()                                                                                                                       

runtime.GC()                                                                                                              
runtime.ReadMemStats(&after)                                                                                              

growth := after.HeapAlloc - before.HeapAlloc                                                                              
return a.Less(growth, threshold, "heap grew by %d bytes", growth)                                                         
}                                                                                                                             

Problems:                                                                                                                     
- Doesn't distinguish leak from legitimate allocation                                                                         
- Flaky due to GC timing                                                                                                      
- Requires careful threshold tuning                                                                                           

2. Specific Pattern Detection:                                                                                                

More practical - test for common leak patterns:                                                                               
// Assert slice doesn't grow unbounded                                                                                        
assert.BoundedSliceGrowth(t, maxSize, func() {                                                                                
// code that manipulates slice                                                                                            
})                                                                                                                            

// Assert goroutine-local storage is cleaned up                                                                               
assert.NoGoroutineLocalLeak(t, func() {                                                                                       
// code using goroutine locals                                                                                            
})                                                                                                                            

3. Profiling Integration:                                                                                                     
// Not runtime assertion, but test helper                                                                                     
assert.ProfileHeap(t, "testdata/heap.prof", func() {                                                                          
// code to profile                                                                                                        
})                                                                                                                            
// Then manually inspect profiles                                                                                             

My Opinion on Memory Leaks:                                                                                                   

Don't try to auto-detect general memory leaks - it's a fool's errand because:                                                 
- Go's GC is not broken - "leaks" are design issues                                                                           
- Static analysis (go vet, staticcheck) is better at catching patterns                                                        
- Runtime detection has too many false positives                                                                              
- Heap profiling (pprof) is the right tool                                                                                    

Instead, consider:                                                                                                            
- Helper to capture heap profiles for manual inspection                                                                       
- Assertions for specific patterns (bounded collections, resource cleanup)                                                    
- Integration with existing tools (pprof, memstats)                                                                           

My Recommendations (Priority Order):                                                                                          

1. ✅ NoGoroutineLeak - Clear value, proven implementations exist                                                             
2. ✅ NoFileDescriptorLeak (Unix) - Very useful, achievable                                                                   
3. ⚠️ NoFileDescriptorLeak (Windows) - Phase 2, use gopsutil or handle count                                                  
4. ❌ NoMemoryLeak (general) - Skip it, misleading and unreliable                                                             
5. ✅ Bounded resource assertions - Specific, testable patterns                                                               

Code Safety Domain - API Sketch:                                                                                              

// Goroutines                                                                                                                 
assert.NoGoroutineLeak(t, testFunc, opts)                                                                                     
assert.GoroutineCount(t, expected)                                                                                            

// File Descriptors (Unix + partial Windows)                                                                                  
assert.NoFileDescriptorLeak(t, testFunc)                                                                                      
assert.FileDescriptorCount(t, expected)                                                                                       

// Bounded Resources (better than memory leak detection)                                                                      
assert.BoundedHeapGrowth(t, maxBytes, testFunc)                                                                               
assert.BoundedGoroutineCount(t, max, testFunc)                                                                                

// Resource cleanup helpers                                                                                                   
defer assert.Cleanup(t) // Combines all cleanup checks                                                                        

Want me to sketch out a prototype for the goroutine leak detector? That's the lowest-hanging fruit and has clear value.  
