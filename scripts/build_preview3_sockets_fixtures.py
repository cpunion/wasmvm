#!/usr/bin/env python3
"""Build exact-WIT socket APIs and an async Core datagram producer."""
import hashlib,json,runpy,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
ns=runpy.run_path(str(ROOT/'scripts/build_preview3_sockets_schema.py'))
graph=ns['graph'];q=json.dumps;OUT=ROOT/'pkg/wasi/preview3/testdata'

def instance(interface):
 names={index:name for name,index in interface['types'].items()};lines=[];emitted=set()
 def ref(t):
  if isinstance(t,int):emit(t);return '$t'+str(t)
  return t
 def body(index):
  k=graph['types'][index]['kind']
  if k=='resource':raise ValueError('resource binding required')
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
 def emit(i):
  if i in emitted:return
  if graph['types'][i]['kind']=='resource':lines.append('(export '+q(names[i])+' (type $t'+str(i)+' (sub resource)))')
  else:
   text=body(i)
   if i in names:lines.extend(['(type $def'+str(i)+' '+text+')','(export '+q(names[i])+' (type $t'+str(i)+' (eq $def'+str(i)+')))'])
   else:lines.append('(type $t'+str(i)+' '+text+')')
  emitted.add(i)
 for i in names:emit(i)
 for name,f in interface['functions'].items():
  signature='func'+(' async' if 'async' in json.dumps(f['kind']) else '')
  signature+=''.join(' (param '+q(p['name'])+' '+ref(p['type'])+')' for p in f['params'])
  if f['result'] is not None:signature+=' (result '+ref(f['result'])+')'
  lines.append('(export '+q(name)+' ('+signature+'))')
 return '\n'.join(lines)
head='(component\n'
for interface in graph['interfaces']:
 if interface['package']!=len(graph['packages'])-1:continue
 alias='types' if interface['name']=='types' else 'lookup'
 head+='(import "wasi:sockets/'+interface['name']+'@0.3.2" (instance $'+alias+'\n'+instance(interface)+'))\n'
api=head+'(export "types" (instance $types)) (export "lookup" (instance $lookup)))'
guest=head+'''
 (alias export $types "udp-socket" (type $socket))
 (alias export $types "[static]udp-socket.create" (func $create))
 (alias export $types "[method]udp-socket.bind" (func $bind))
 (alias export $types "[method]udp-socket.send" (func $send))
 (core module $mem (memory (export "memory") 1) (data (i32.const 32) "abc")
  (global $heap (mut i32) (i32.const 1024))
  (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
   global.get $heap local.tee $p local.get 3 i32.add global.set $heap local.get $p))
 (core instance $mem (instantiate $mem)) (alias core export $mem "memory" (core memory $memory))
 (alias core export $mem "realloc" (core func $realloc))
 (core func $create (canon lower (func $create) (memory $memory) (realloc $realloc)))
 (core func $bind (canon lower (func $bind) (memory $memory) (realloc $realloc)))
 (core func $send (canon lower (func $send) (memory $memory) (realloc $realloc)))
 (core func $drop (canon resource.drop $socket))
 (core instance $env (export "memory" (memory $memory)) (export "create" (func $create)) (export "bind" (func $bind)) (export "send" (func $send)) (export "drop" (func $drop)))
 (core module $m
  (import "env" "memory" (memory 1)) (import "env" "create" (func $create (param i32 i32)))
  (import "env" "bind" (func $bind (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)))
  (import "env" "send" (func $send (param i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32 i32)))
  (import "env" "drop" (func $drop (param i32)))
  (func (export "run") (param $port i32) (result i32) (local $s i32) (local $code i32)
   i32.const 0 i32.const 128 call $create i32.const 128 i32.load8_u if unreachable end
   i32.const 132 i32.load local.set $s
   local.get $s i32.const 0 i32.const 0 i32.const 127 i32.const 0 i32.const 0 i32.const 1 i32.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 160 call $bind
   i32.const 160 i32.load8_u if unreachable end
   local.get $s i32.const 32 i32.const 3 i32.const 1 i32.const 0 local.get $port i32.const 127 i32.const 0 i32.const 0 i32.const 1 i32.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 192 call $send
   i32.const 192 i32.load8_u local.set $code
   local.get $s call $drop local.get $code))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (func (export "run") async (param "port" u16) (result u32) (canon lift (core func $m "run"))))
'''
fixtures={'sockets-api':api,'sockets-guest':guest,'sockets-wrong-error':api.replace('"connection-broken"','"connection-broken-other"'),'sockets-wrong-async':api.replace('(func async (param "self" $t15) (param "remote-address" $t10)','(func (param "self" $t15) (param "remote-address" $t10)',1)}
assert fixtures['sockets-wrong-async']!=api
pins={}
for name,text in fixtures.items():
 p=OUT/(name+'.wat');p.write_text(text+'\n')
 subprocess.run(['wasm-tools','parse',str(p),'-o',str(p.with_suffix('.wasm'))],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(p.with_suffix('.wasm'))],check=True)
 pins[name]={ext:hashlib.sha256(p.with_suffix('.'+ext).read_bytes()).hexdigest() for ext in ['wat','wasm']}
(OUT/'sockets.json').write_text(json.dumps({'wasiVersion':'0.3.2','witArchiveSHA256':ns['SHA'],'producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'sha256':pins},indent=2)+'\n')
print('Built',len(pins),'validated socket fixtures')
