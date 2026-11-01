#!/usr/bin/env python3
"""Build exact-WIT HTTP resource APIs and malformed binding contracts."""
import hashlib,json,runpy,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
ns=runpy.run_path(str(ROOT/'scripts/build_preview3_http_schema.py'))
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
  if kind=='type':
   if isinstance(v,int) and (base_resource(v) is not None or v in resources):return '(eq '+ref(v)+')'
   return body(v) if isinstance(v,int) else v
  if kind=='handle':handle,t=next(iter(v.items()));return '('+handle+' '+ref(t)+')'
  if kind in ['list','option','future','stream']:return '('+kind+(' '+ref(v) if v is not None else '')+')'
  if kind=='record':return '(record '+ ' '.join('(field '+q(f['name'])+' '+ref(f['type'])+')' for f in v['fields'])+')'
  if kind=='variant':return '(variant '+ ' '.join('(case '+q(f['name'])+(' '+ref(f['type']) if f['type'] is not None else '')+')' for f in v['cases'])+')'
  if kind in ['enum','flags']:return '('+kind+' '+' '.join(q(f['name']) for f in v['cases' if kind=='enum' else 'flags'])+')'
  if kind=='tuple':return '(tuple '+' '.join(ref(t) for t in v['types'])+')'
  if kind=='result':return '(result'+(' '+ref(v['ok']) if v['ok'] is not None else '')+(' (error '+ref(v['err'])+')' if v['err'] is not None else '')+')'
  raise ValueError(kind)
 def base_resource(i):
  k=graph['types'][i]['kind']
  if k=='resource':return i
  if isinstance(k,dict) and isinstance(k.get('type'),int):return base_resource(k['type'])
  return None
 def emit(i):
  if i in emitted:return
  if i in resources:lines.append('(alias outer 1 $r'+str(i)+' (type $t'+str(i)+'))')
  elif graph['types'][i]['kind']=='resource':lines.append('(export '+q(names[i])+' (type $t'+str(i)+' (sub resource)))')
  else:
   text=body(i)
   if base_resource(i) is not None or isinstance(graph['types'][i]['kind'],dict) and graph['types'][i]['kind'].get('type') in resources:
    v=graph['types'][i]['kind']['type'];target=ref(v)
    if i in names:lines.append('(export '+q(names[i])+' (type $t'+str(i)+' (eq '+target+')))')
    else:lines.append('(alias outer 1 $r'+str(base_resource(i))+' (type $t'+str(i)+'))')
   elif i in names:lines.extend(['(type $def'+str(i)+' '+text+')','(export '+q(names[i])+' (type $t'+str(i)+' (eq $def'+str(i)+')))'])
   else:lines.append('(type $t'+str(i)+' '+text+')')
  emitted.add(i)
 for i in names:emit(i)
 for name,f in interface['functions'].items():
  signature='func'+(' async' if 'async' in json.dumps(f['kind']) else '')
  signature+=''.join(' (param '+q(p['name'])+' '+ref(p['type'])+')' for p in f['params'])
  if f['result'] is not None:signature+=' (result '+ref(f['result'])+')'
  lines.append('(export '+q(name)+' ('+signature+'))')
 return '\n'.join(lines)
resources=set()
head='(component\n'
for interface in graph['interfaces']:
 if interface['package']!=len(graph['packages'])-1:continue
 alias=interface['name']
 head+='(import "wasi:http/'+interface['name']+'@0.3.2" (instance $'+alias+'\n'+instance(interface)+'))\n'
 for name,i in interface['types'].items():
  head+='(alias export $'+alias+' '+q(name)+' (type $r'+str(i)+'))\n';resources.add(i)

api=head+'(export "types" (instance $types)) (export "handler" (instance $handler)) (export "client" (instance $client)))'
fixtures={'http-api':api,'http-wrong-header-error':api.replace('"size-exceeded"','"size-other"'),'http-wrong-async':api.replace('(export "send" (func async','(export "send" (func',1)}

fixtures['http-handler-api']=head+"""
(alias export $handler "handle" (func $handle))
(instance $serve (export "handle" (func $handle)))
(export "wasi:http/handler@0.3.2" (instance $serve)))
"""
client_head=head[:head.index('(import "wasi:http/handler@')]+head[head.index('(import "wasi:http/client@'):]
fixtures['http-guest']=client_head+r"""
(alias export $types "[constructor]fields" (func $fields))
(alias export $types "[static]request.new" (func $new))
(alias export $types "[method]request.set-authority" (func $authority))
(alias export $types "[method]request.set-path-with-query" (func $path))
(alias export $types "[method]request.set-scheme" (func $scheme))
(alias export $types "[method]response.get-status-code" (func $status))
(alias export $client "send" (func $send))
(type $trailers-result (result (option (own $r39)) (error $r34)))
(type $trailers (future $trailers-result))
(type $tx (future (result (error $r34))))
(core module $memory (memory (export "memory") 1) (data (i32.const 32) "/")
 (global $heap (mut i32) (i32.const 1024))
 (func (export "realloc") (param i32 i32 i32 i32) (result i32) (local $p i32)
  global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $p
  local.get 3 i32.add global.set $heap local.get $p))
(core instance $memory (instantiate $memory))
(alias core export $memory "memory" (core memory $mem))
(alias core export $memory "realloc" (core func $realloc))
(core func $fields (canon lower (func $fields)))
(core func $new (canon lower (func $new) (memory $mem) (realloc $realloc)))
(core func $authority (canon lower (func $authority) (memory $mem) (realloc $realloc)))
(core func $path (canon lower (func $path) (memory $mem) (realloc $realloc)))
(core func $scheme (canon lower (func $scheme) (memory $mem) (realloc $realloc)))
(core func $send (canon lower (func $send) (memory $mem) (realloc $realloc)))
(core func $status (canon lower (func $status)))
(core func $drop-response (canon resource.drop $r45))
(core func $new-trailers (canon future.new $trailers))
(core func $write-trailers (canon future.write $trailers (memory $mem) async))
(core func $cancel-write (canon future.cancel-write $trailers))
(core func $drop-writable (canon future.drop-writable $trailers))
(core func $read-tx (canon future.read $tx (memory $mem) (realloc $realloc)))
(core func $drop-tx (canon future.drop-readable $tx))
(core instance $env (export "memory" (memory $mem))
 (export "fields" (func $fields)) (export "new" (func $new)) (export "authority" (func $authority))
 (export "path" (func $path)) (export "scheme" (func $scheme))
 (export "send" (func $send)) (export "status" (func $status)) (export "drop-response" (func $drop-response))
 (export "new-trailers" (func $new-trailers)) (export "write-trailers" (func $write-trailers))
 (export "cancel-write" (func $cancel-write)) (export "drop-writable" (func $drop-writable))
 (export "read-tx" (func $read-tx)) (export "drop-tx" (func $drop-tx)))
(core module $m
 (import "env" "memory" (memory 1))
 (import "env" "fields" (func $fields (result i32)))
 (import "env" "new" (func $new (param i32 i32 i32 i32 i32 i32 i32)))
 (import "env" "authority" (func $authority (param i32 i32 i32 i32) (result i32)))
 (import "env" "path" (func $path (param i32 i32 i32 i32) (result i32)))
 (import "env" "scheme" (func $scheme (param i32 i32 i32 i32 i32) (result i32)))
 (import "env" "send" (func $send (param i32 i32)))
 (import "env" "status" (func $status (param i32) (result i32)))
 (import "env" "drop-response" (func $drop-response (param i32)))
 (import "env" "new-trailers" (func $new-trailers (result i64)))
 (import "env" "write-trailers" (func $write-trailers (param i32 i32) (result i32)))
 (import "env" "cancel-write" (func $cancel-write (param i32) (result i32)))
 (import "env" "drop-writable" (func $drop-writable (param i32)))
 (import "env" "read-tx" (func $read-tx (param i32 i32) (result i32)))
 (import "env" "drop-tx" (func $drop-tx (param i32)))
 (func (export "run") (param $ptr i32) (param $len i32) (result i32)
  (local $ends i64) (local $request i32) (local $writer i32) (local $response i32) (local $status i32)
  call $new-trailers local.tee $ends i64.const 32 i64.shr_u i32.wrap_i64 local.set $writer
  local.get $writer i32.const 128 call $write-trailers i32.const -1 i32.ne if unreachable end
  call $fields i32.const 0 i32.const 0 local.get $ends i32.wrap_i64 i32.const 0 i32.const 0 i32.const 64 call $new
  i32.const 64 i32.load local.set $request
  local.get $request i32.const 1 local.get $ptr local.get $len call $authority if unreachable end
  local.get $request i32.const 1 i32.const 32 i32.const 1 call $path if unreachable end
  local.get $request i32.const 1 i32.const 0 i32.const 0 i32.const 0 call $scheme if unreachable end
  local.get $request i32.const 192 call $send
  i32.const 192 i32.load8_u if
   i32.const 200 i32.load8_u i32.const 65536 i32.or local.set $status
  else
   i32.const 200 i32.load local.tee $response call $status local.set $status
   local.get $response call $drop-response
  end
  i32.const 68 i32.load i32.const 256 call $read-tx drop
  i32.const 68 i32.load call $drop-tx
  local.get $writer call $cancel-write i32.const 1 i32.gt_u if unreachable end
  local.get $writer call $drop-writable
  local.get $status))
(core instance $m (instantiate $m (with "env" (instance $env))))
(func (export "run") async (param "authority" string) (result u32)
 (canon lift (core func $m "run") (memory $mem) (realloc $realloc))))
"""

memory=fixtures['http-guest'].split('(core module $memory',1)[1].split('(core func $fields',1)[0]
types_head=head[:head.index('(import "wasi:http/handler@')]
fixtures['http-server']=types_head+r"""
(alias export $types "[constructor]fields" (func $fields))
(alias export $types "[static]response.new" (func $new))
(type $trailers (future (result (option (own $r39)) (error $r34))))
(type $tx (future (result (error $r34))))
(type $bytes (stream u8))
(type $reply (result (own $r45) (error $r34)))
"""+'(core module $memory'+memory+r"""
(core func $fields (canon lower (func $fields)))
(core func $new (canon lower (func $new) (memory $mem) (realloc $realloc)))
(core func $drop-request (canon resource.drop $r42))
(core func $new-trailers (canon future.new $trailers))
(core func $write-trailers (canon future.write $trailers (memory $mem)))
(core func $drop-trailers (canon future.drop-writable $trailers))
(core func $new-bytes (canon stream.new $bytes))
(core func $write-bytes (canon stream.write $bytes (memory $mem)))
(core func $drop-bytes (canon stream.drop-writable $bytes))
(core func $return (canon task.return (result $reply) (memory $mem)))
(core func $read-tx (canon future.read $tx (memory $mem) (realloc $realloc)))
(core func $drop-tx (canon future.drop-readable $tx))
(core instance $env (export "memory" (memory $mem))
 (export "fields" (func $fields)) (export "new" (func $new)) (export "drop-request" (func $drop-request))
 (export "new-trailers" (func $new-trailers)) (export "write-trailers" (func $write-trailers)) (export "drop-trailers" (func $drop-trailers))
 (export "new-bytes" (func $new-bytes)) (export "write-bytes" (func $write-bytes)) (export "drop-bytes" (func $drop-bytes))
 (export "return" (func $return)) (export "read-tx" (func $read-tx)) (export "drop-tx" (func $drop-tx)))
(core module $m
 (import "env" "memory" (memory 1))
 (import "env" "fields" (func $fields (result i32)))
 (import "env" "new" (func $new (param i32 i32 i32 i32 i32)))
 (import "env" "drop-request" (func $drop-request (param i32)))
 (import "env" "new-trailers" (func $new-trailers (result i64)))
 (import "env" "write-trailers" (func $write-trailers (param i32 i32) (result i32)))
 (import "env" "drop-trailers" (func $drop-trailers (param i32)))
 (import "env" "new-bytes" (func $new-bytes (result i64)))
 (import "env" "write-bytes" (func $write-bytes (param i32 i32 i32) (result i32)))
 (import "env" "drop-bytes" (func $drop-bytes (param i32)))
 (import "env" "return" (func $return (param i32 i32 i32 i64 i32 i32 i32 i32)))
 (import "env" "read-tx" (func $read-tx (param i32 i32) (result i32)))
 (import "env" "drop-tx" (func $drop-tx (param i32)))
 (data (i32.const 32) "hello")
 (func (export "handle") (param $request i32)
  (local $trailers i64) (local $bytes i64) (local $written i32) (local $status i32)
  local.get $request call $drop-request
  call $new-trailers local.set $trailers call $new-bytes local.set $bytes
  call $fields i32.const 1 local.get $bytes i32.wrap_i64 local.get $trailers i32.wrap_i64 i32.const 64 call $new
  i32.const 0 i32.const 64 i32.load i32.const 0 i64.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 0 call $return
  loop $write
   local.get $bytes i64.const 32 i64.shr_u i32.wrap_i64
   i32.const 32 local.get $written i32.add i32.const 5 local.get $written i32.sub call $write-bytes local.tee $status
   i32.const 15 i32.and if unreachable end
   local.get $written local.get $status i32.const 4 i32.shr_u i32.add local.tee $written i32.const 5 i32.lt_u br_if $write
  end
  local.get $bytes i64.const 32 i64.shr_u i32.wrap_i64 call $drop-bytes
  local.get $trailers i64.const 32 i64.shr_u i32.wrap_i64 i32.const 128 call $write-trailers drop
  local.get $trailers i64.const 32 i64.shr_u i32.wrap_i64 call $drop-trailers
  i32.const 68 i32.load i32.const 256 call $read-tx drop i32.const 68 i32.load call $drop-tx))
(core instance $m (instantiate $m (with "env" (instance $env))))
(func $handle async (param "request" (own $r42)) (result $reply)
 (canon lift (core func $m "handle") (memory $mem) async))
(instance $serve (export "handle" (func $handle)))
(export "wasi:http/handler@0.3.2" (instance $serve)))
"""

pins={}
for name,text in fixtures.items():
 p=OUT/(name+'.wat');p.write_text(text.rstrip()+'\n')
 subprocess.run(['wasm-tools','parse',str(p),'-o',str(p.with_suffix('.wasm'))],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(p.with_suffix('.wasm'))],check=True)
 pins[name]={ext:hashlib.sha256(p.with_suffix('.'+ext).read_bytes()).hexdigest() for ext in ['wat','wasm']}
(OUT/'http.json').write_text(json.dumps({'wasiVersion':'0.3.2','witArchiveSHA256':ns['SHA'],'producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'sha256':pins},indent=2)+'\n')
print('Built',len(pins),'validated HTTP fixtures')
