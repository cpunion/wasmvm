#!/usr/bin/env python3
"""Build exact-WIT filesystem API and real asynchronous Core guest fixtures."""
import hashlib,json,runpy,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
ns=runpy.run_path(str(ROOT/'scripts/build_preview3_filesystem_schema.py'))
graph=ns['graph'];q=json.dumps;OUT=ROOT/'pkg/wasi/preview3/testdata'
fs=next(i for i in graph['interfaces'] if i['name']=='types' and i['package']==len(graph['packages'])-1)
pre=next(i for i in graph['interfaces'] if i['name']=='preopens')
def ref(t):return '$t'+str(t) if isinstance(t,int) else t

def body(index):
 k=graph['types'][index]['kind']
 if k=='resource':raise ValueError('resource needs an import binding')
 kind,v=next(iter(k.items()))
 if kind=='type':return body(v) if isinstance(v,int) else v
 if kind=='handle':handle,t=next(iter(v.items()));return '('+handle+' '+ref(t)+')'
 if kind in ['list','option','future','stream']:return '('+kind+(' '+ref(v) if v is not None else '')+')'
 if kind=='record':return '(record '+ ' '.join('(field '+q(f['name'])+' '+ref(f['type'])+')' for f in v['fields'])+')'
 if kind=='variant':return '(variant '+ ' '.join('(case '+q(f['name'])+(' '+ref(f['type']) if f['type'] is not None else '')+')' for f in v['cases'])+')'
 if kind in ['enum','flags']:return '('+kind+' '+' '.join(q(f['name']) for f in v['cases' if kind=='enum' else 'flags'])+')'
 if kind=='tuple':return '(tuple '+' '.join(ref(t) for t in v['types'])+')'
 if kind=='result':return '(result'+(' '+ref(v['ok']) if v['ok'] is not None else '')+(' (error '+ref(v['err'])+')' if v['err'] is not None else '')+')'
 raise ValueError(kind)
def instance(interface,start,end):
 names={index:name for name,index in interface['types'].items()};lines=[]
 for i in range(start,end):
  if graph['types'][i]['kind']=='resource':lines+=['(export '+q(names[i])+' (type $t'+str(i)+' (sub resource)))'];continue
  if i==33:
   lines+=['(alias outer 1 $descriptor (type $outer))','(export "descriptor" (type $t33 (eq $outer)))'];continue
  if i in names:lines+=['(type $def'+str(i)+' '+body(i)+')','(export '+q(names[i])+' (type $t'+str(i)+' (eq $def'+str(i)+')))']
  else:lines+=['(type $t'+str(i)+' '+body(i)+')']
 for name,f in interface['functions'].items():
  async_='async' in json.dumps(f['kind']);signature='func'+(' async' if async_ else '')
  signature+=''.join(' (param '+q(p['name'])+' '+ref(p['type'])+')' for p in f['params'])
  if f['result'] is not None:signature+=' (result '+ref(f['result'])+')'
  lines+=['(export '+q(name)+' ('+signature+'))']
 return '\n'.join(lines)
head='''(component
 (type $fs (instance '''+instance(fs,0,33)+'''))
 (import "wasi:filesystem/types@0.3.2" (instance $fs (type $fs)))
 (alias export $fs "descriptor" (type $descriptor))
 (type $preopens (instance '''+instance(pre,33,37)+'''))
 (import "wasi:filesystem/preopens@0.3.2" (instance $preopens (type $preopens)))
 (export "fs" (instance $fs)) (export "preopens" (instance $preopens))
'''
guest='''
 (alias export $preopens "get-directories" (func $directories))
 (alias export $fs "[method]descriptor.open-at" (func $open))
 (alias export $fs "[method]descriptor.stat" (func $stat))
 (core module $memory (memory (export "memory") 1) (data (i32.const 32) "data.txt")
  (global $heap (mut i32) (i32.const 1024))
  (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
   global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $p
   local.get 3 i32.add global.set $heap local.get $p))
 (core instance $memory (instantiate $memory)) (alias core export $memory "memory" (core memory $mem)) (alias core export $memory "realloc" (core func $realloc))
 (core func $directories (canon lower (func $directories) (memory $mem) (realloc $realloc)))
 (core func $open (canon lower (func $open) (memory $mem) (realloc $realloc)))
 (core func $stat (canon lower (func $stat) (memory $mem) (realloc $realloc)))
 (core func $drop (canon resource.drop $descriptor))
 (core instance $env (export "memory" (memory $mem)) (export "directories" (func $directories)) (export "open" (func $open)) (export "stat" (func $stat)) (export "drop" (func $drop)))
 (core module $m
  (import "env" "memory" (memory 1)) (import "env" "directories" (func $directories (param i32)))
  (import "env" "open" (func $open (param i32 i32 i32 i32 i32 i32 i32))) (import "env" "stat" (func $stat (param i32 i32))) (import "env" "drop" (func $drop (param i32)))
  (func (export "run") (result i64) (local $root i32) (local $file i32) (local $size i64)
   i32.const 0 call $directories i32.const 4 i32.load i32.eqz if unreachable end
   i32.const 0 i32.load i32.load local.set $root
   local.get $root i32.const 0 i32.const 32 i32.const 8 i32.const 0 i32.const 1 i32.const 128 call $open
   i32.const 128 i32.load8_u if local.get $root call $drop i32.const 132 i32.load8_u i64.extend_i32_u i64.const 4294967296 i64.or return end i32.const 132 i32.load local.set $file
   local.get $file i32.const 256 call $stat i32.const 256 i32.load8_u if unreachable end
   i32.const 264 i32.load8_u i32.const 5 i32.ne if unreachable end
   i32.const 288 i64.load local.set $size
   local.get $file call $drop local.get $root call $drop local.get $size))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (func (export "run") async (result u64) (canon lift (core func $m "run"))))'''
fixtures={'filesystem-api':head+')','filesystem-guest':head.replace('(export "fs" (instance $fs)) (export "preopens" (instance $preopens))','')+guest}
fixtures['filesystem-wrong-flags']=fixtures['filesystem-api'].replace('"mutate-directory"','"mutate-other"')
fixtures['filesystem-wrong-time']=fixtures['filesystem-api'].replace('(field "seconds" s64)','(field "seconds" u64)')
source=ROOT/'pkg/component/testdata/host-source-stream-u8.wat'
source_text=source.read_text()
stdin=' (import "wasi:cli/stdin@0.3.2" (instance $stdin\n  (type $e (enum "io" "illegal-byte-sequence" "pipe")) (export "error-code" (type $err (eq $e)))\n  (type $s (stream u8)) (type $f (future (result (error $err)))) (type $r (tuple $s $f))\n  (export "read-via-stream" (func (result $r)))))\n (alias export $stdin "error-code" (type $err)) (export "error-type" (type $err))\n (alias export $stdin "read-via-stream" (func $input)) (export "input" (func $input))\n'
fixtures['stdin-cancel-progress']=source_text.replace('(component\n','(component\n'+stdin,1).replace('call $cancel i32.const 2','call $cancel i32.const 18')
pins={}
for name,text in fixtures.items():
 p=OUT/(name+'.wat');p.write_text(text+'\n')
 subprocess.run(['wasm-tools','parse',str(p),'-o',str(p.with_suffix('.wasm'))],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(p.with_suffix('.wasm'))],check=True)
 pins[name]={ext:hashlib.sha256(p.with_suffix('.'+ext).read_bytes()).hexdigest() for ext in ['wat','wasm']}
(OUT/'filesystem.json').write_text(json.dumps({'wasiVersion':'0.3.2','witArchiveSHA256':ns['SHA'],'producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'sourceSHA256':{str(source.relative_to(ROOT)):hashlib.sha256(source.read_bytes()).hexdigest()},'sha256':pins},indent=2)+'\n')
print('Built',len(pins),'validated filesystem/IO fixtures')
