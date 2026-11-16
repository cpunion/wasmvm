from pathlib import Path
import ctypes,json,os,random,re,statistics,subprocess,threading

OUT=Path(__file__).resolve().parent
ROOT=Path('/private/tmp/wasmvm-p3-gc-20261116')
header=Path('/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/sys/resource.h').read_text()
body=header.split('struct rusage_info_v4 {',1)[1].split('};',1)[0]
fields=re.findall(r'uint64_t\s+(ri_\w+);',body)
class Info(ctypes.Structure):_fields_=[('ri_uuid',ctypes.c_ubyte*16)]+[(n,ctypes.c_uint64)for n in fields]
lib=ctypes.CDLL('/usr/lib/libproc.dylib',use_errno=True)
lib.proc_pid_rusage.argtypes=[ctypes.c_int,ctypes.c_int,ctypes.c_void_p];lib.proc_pid_rusage.restype=ctypes.c_int
names=['ri_user_time','ri_system_time','ri_instructions','ri_cycles','ri_pkg_idle_wkups','ri_interrupt_wkups','ri_phys_footprint','ri_lifetime_max_phys_footprint','ri_runnable_time']
def snapshot(pid):
 info=Info();result=lib.proc_pid_rusage(pid,4,ctypes.byref(info))
 if result:raise OSError(ctypes.get_errno(),'proc_pid_rusage')
 return {n:int(getattr(info,n)) for n in names}
assert snapshot(os.getpid())['ri_user_time']>0
class Timebase(ctypes.Structure):
 _fields_=[('numer',ctypes.c_uint32),('denom',ctypes.c_uint32)]
system=ctypes.CDLL('/usr/lib/libSystem.B.dylib')
system.mach_timebase_info.argtypes=[ctypes.POINTER(Timebase)]
system.mach_timebase_info.restype=ctypes.c_int
clockbase=Timebase()
assert system.mach_timebase_info(ctypes.byref(clockbase))==0
tickNanoseconds=clockbase.numer/clockbase.denom



import hashlib,time,signal,gzip
variants={'base':'/private/tmp/llgo-p3-gc-base-20261116','head':'/private/tmp/llgo-p3-gc-scan-20261116'}
plan=dict(blocks=20,samples=20,batchRequests=512,seed=202611150601,payloadBytes=65536,variants=variants,source='Frozen wasmvm393bf75c; main113daaad5 versus scan-work plus occupied-heap GOGC pacing candidate; LLVM22.1.8, Go1.27.3, -O3/fullLTO, GOGC100, normal GC.',scope='Fixed20blocks/80processes per payload with identical-binary controls and 20 interleaved batches; focused GC diagnostic, not portfolio acceptance. All phases and adverse controls retained.')
plan['binarySHA256']={k:hashlib.sha256(Path(v).read_bytes()).hexdigest()for k,v in variants.items()};plan['machTimebase']={'numer':clockbase.numer,'denom':clockbase.denom};assert not(OUT/'rows.jsonl.gz').exists();(OUT/'plan.json').write_text(json.dumps(plan,indent=2)+'\n');rng=random.Random(plan['seed']);rows=[]
def freeze(p):
 os.kill(p.pid,signal.SIGSTOP);pid,status=os.waitpid(p.pid,os.WUNTRACED);assert pid==p.pid and os.WIFSTOPPED(status)
def resume(p):os.kill(p.pid,signal.SIGCONT)
for block in range(plan['blocks']):
 load=[]
 for line in subprocess.check_output(['ps','-axo','pid,pcpu,comm'],text=True).splitlines()[1:]:
  fields=line.split(None,2)
  if len(fields)==3 and float(fields[1])>5:load.append(dict(cpuPercent=float(fields[1]),pid=int(fields[0]),command=fields[2]))
 (OUT/f'host-load-{block}.json').write_text(json.dumps(sorted(load,key=lambda p:-p['cpuPercent']),indent=2)+'\n')
 order=[(name,alias)for name in variants for alias in ('','-aa')];rng.shuffle(order);children=[]
 try:
  for name,alias in order:
   reference=name=='wasmtime';cmd=[str(variants[name]),'--engine','wasmtime'if reference else'wasmvm','--backend','compiler'if reference else'native','--wasm',str(ROOT/'pkg/wasi/preview3/testdata/world-rust-http.wasm'),'--payload-bytes',str(plan['payloadBytes']),'--expected',str(1757737011 if plan['payloadBytes']==16 else 2629362236),'--workers','4','--clients','16','--requests','32','--samples','21','--warmups','2']
   if not reference:cmd+=['--response-mode','buffered','--request-mode','snapshot']
   env=dict(os.environ,GOMAXPROCS='4',GOGC='100',WASMVM_INTERLEAVE='1')
   for key in ['GC_DONT_GC','WASMVM_DIAGNOSTIC_JIT_MAP','WASMVM_WARM_PROFILE','WASMVM_WARM_ALLOC_PROFILE','WASMVM_COMPILE_PHASE']:env.pop(key,None)
   err=(OUT/f'{block}-{name}{alias}.stderr').open('w');p=subprocess.Popen(cmd,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=err,text=True,env=env);child=dict(id=name+alias,process=p,errorFile=err,command=cmd,deltas=[],rounds=[]);children.append(child);child['watchdog']=threading.Timer(120,p.kill);child['watchdog'].start()
   assert json.loads(p.stdout.readline())=={'ready':True};p.stdin.write('run\n');p.stdin.flush();assert json.loads(p.stdout.readline())=={'sampleReady':0};time.sleep(.1);freeze(p)
  for sample_index in range(plan['samples']):
   order=list(children);rng.shuffle(order)
   for c in order:
    p=c['process'];before=snapshot(p.pid);start=time.perf_counter_ns();resume(p);p.stdin.write('sample\n');p.stdin.flush();assert json.loads(p.stdout.readline())=={'sampleReady':sample_index+1}
    if sample_index==plan['samples']-1:time.sleep(.1)
    freeze(p);after=snapshot(p.pid);c['deltas'].append({n:after[n]-before[n]for n in names});c['rounds'].append(dict(round=sample_index,dispatchNs=start,stopNs=time.perf_counter_ns(),before=before,after=after))
  for c in children:
   p=c['process'];resume(p);p.stdin.write('sample\n');p.stdin.close();record=json.loads(p.stdout.readline());rest=p.stdout.read();rc=p.wait(timeout=30);assert record['hostCompiler']==('llgo' if c['id'].removesuffix('-aa') in ('base','head') else 'gc');assert rc==0 and not rest and record['warmResults']==[1757737011 if plan['payloadBytes']==16 else 2629362236]*21;counts=plan['samples']*512;total={n:sum(d[n]for d in c['deltas'])for n in names};batches=record['service']['Batches'][3:23];assert len(batches)==20
   row=dict(block=block,id=c['id'],command=c['command'],rounds=c['rounds'],windowRequests=counts,totals=total,cpuUs=(total['ri_user_time']+total['ri_system_time'])*tickNanoseconds/counts/1000,instructions=total['ri_instructions']/counts,throughput=512e9/statistics.median(record['warmCallNs'][:20]),allocsPerRequest=sum(b.get('Mallocs',0)for b in batches)/counts,bytesPerRequest=sum(b['TotalAllocBytes']for b in batches)/counts,record=record)
   with gzip.open(OUT/'rows.jsonl.gz','at',compresslevel=1)as f:f.write(json.dumps(row)+'\n')
   rows.append(row);print('checked',block,c['id'],flush=True)
 finally:
  for c in children:
   p=c['process'];c['watchdog'].cancel()
   if p.poll()is None:p.kill();p.wait()
   c['errorFile'].close()
summary={name+alias:{key:statistics.median(r[key]for r in rows if r['id']==name+alias)for key in ['cpuUs','instructions','throughput','allocsPerRequest','bytesPerRequest']}for name in variants for alias in ('','-aa')}
(OUT/'summary.json').write_text(json.dumps(dict(plan=plan,summary=summary),indent=2)+'\n');print(json.dumps(summary),flush=True)
