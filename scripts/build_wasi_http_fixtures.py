#!/usr/bin/env python3
"""Reproduce client/proxy guests for the complete stable WASI 0.2.12 HTTP world."""
import argparse,hashlib,json,re,shutil,subprocess,tarfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];SHA='2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992'
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--archive',type=Path,default=ROOT/'artifacts/components/wasi-inputs/wasi-wit-0.2.12.tar.gz');p.add_argument('--work',type=Path,default=ROOT/'artifacts/components/http-producer');p.add_argument('--output',type=Path,default=ROOT/'pkg/wasi/preview2/testdata');a=p.parse_args();work,out=a.work,a.output
assert hashlib.sha256(a.archive.read_bytes()).hexdigest()==SHA
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
(work/'wit/deps').mkdir(parents=True,exist_ok=True);out.mkdir(parents=True,exist_ok=True)
with tarfile.open(a.archive) as archive:
 for name in archive.getnames():
  if name.startswith('wasi-wit-0.2.12/http/') and name.endswith('.wit'):(work/'wit/deps'/Path(name).name).write_bytes(archive.extractfile(name).read())
world=ROOT/'pkg/wasi/preview2/testdata/http.wit';shutil.copyfile(world,work/'wit/world.wit')
subprocess.run(['wasm-tools','component','embed',str(work/'wit'),'--world','http','--dummy-names','legacy','-t','-o',str(work/'dummy.wat')],check=True)
s=(work/'dummy.wat').read_text();imports=re.findall(r'\(import "([^"]+)" "([^"]+)" \(func \(;(\d+);\)',s)
assert len([x for x in imports if x[0].startswith('wasi:http/') and not x[1].startswith('[resource-drop]')])==52
for module,name,index in imports:
 label=module.split('/')[1].split('@')[0]+'.'+name.replace('[method]','').replace('[constructor]','new-').replace('[static]','').replace('[resource-drop]','drop-')
 s=s.replace(f'(func (;{index};) (type',f'(func ${label} (;{index};) (type',1)
exports={name:int(index) for name,index in re.findall(r'\(export "([^"]+)" \(func (\d+)\)\)',s)}
s=s.replace('(memory (;0;) 0)','''(memory (;0;) 64 128)
 (global $heap (mut i32) (i32.const 8192))
 (data (i32.const 1024) "x-wasmvm") (data (i32.const 1040) "value")
 (data (i32.const 1056) "/echo") (data (i32.const 1072) "x-request-trailer")
 (data (i32.const 1144) "trailer") (data (i32.const 1104) "done") (data (i32.const 1120) "x-reply-trailer")
 (data (i32.const 2048) "'''+('0123456789abcdef'*256)+'''")''')
def body(name,locals,code):
 global s
 index=exports[name];pattern=r'(\(func \(;'+str(index)+r';\)[^\n]*)\n\s+unreachable\n  \)'
 s,n=re.subn(pattern,lambda m:m[1]+' '+locals+'\n'+code+'\n )',s,count=1);assert n==1,name
extra='''
 (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
 (func $trailers (param $future i32) (param $name i32) (param $len i32) (local $poll i32) (local $headers i32)
 block $ready loop $next
 local.get $future i32.const 64 call $types.future-trailers.get i32.const 64 i32.load8_u br_if $ready
 local.get $future call $types.future-trailers.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $next end end
 i32.const 72 i32.load8_u if unreachable end i32.const 80 i32.load8_u if unreachable end
 i32.const 88 i32.load8_u if
 i32.const 92 i32.load local.set $headers
 local.get $headers local.get $name local.get $len i32.const 64 call $types.fields.get
 i32.const 68 i32.load i32.const 1 i32.ne if unreachable end
 i32.const 64 i32.load i32.const 4 i32.add i32.load i32.const 4 i32.ne if unreachable end
 local.get $headers call $types.drop-fields end
 local.get $future call $types.drop-future-trailers)
 (func $read-body (param $body i32) (result i32) (local $stream i32) (local $at i32) (local $n i32) (local $ptr i32)
 local.get $body i32.const 64 call $types.incoming-body.stream i32.const 64 call $ok i32.const 68 i32.load local.set $stream
 block $end loop $next
 local.get $stream i64.const 4096 i32.const 64 call $streams.input-stream.blocking-read
 i32.const 64 i32.load8_u if
 i32.const 68 i32.load8_u i32.const 1 i32.ne if unreachable end br $end end
 i32.const 68 i32.load local.set $ptr i32.const 72 i32.load local.tee $n i32.eqz if unreachable end
 local.get $ptr i32.load8_u i32.const 2048 local.get $at i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end
 local.get $ptr local.get $n i32.add i32.const 1 i32.sub i32.load8_u i32.const 2048 local.get $at local.get $n i32.add i32.const 1 i32.sub i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end
 local.get $at local.get $n i32.add local.tee $at i32.const 65536 i32.gt_u if unreachable end
 br $next end end
 local.get $stream call $streams.drop-input-stream local.get $at)
 (func $new-trailers (param $name i32) (param $len i32) (result i32) (local $fields i32)
 call $types.new-fields local.tee $fields local.get $name local.get $len i32.const 1104 i32.const 4 i32.const 64 call $types.fields.append i32.const 64 call $ok local.get $fields)
 (func $request (param $authority i32) (param $alen i32) (param $secure i32) (result i32)
 (local $headers i32) (local $request i32) (local $body i32) (local $stream i32) (local $options i32) (local $future i32) (local $poll i32) (local $response i32) (local $incoming i32)
 call $types.new-fields local.tee $headers i32.const 1024 i32.const 8 i32.const 1040 i32.const 5 i32.const 64 call $types.fields.append i32.const 64 call $ok
 local.get $headers i32.const 1144 i32.const 7 i32.const 1072 i32.const 17 i32.const 64 call $types.fields.append i32.const 64 call $ok
 local.get $headers call $types.new-outgoing-request local.set $request
 local.get $request i32.const 2 i32.const 0 i32.const 0 call $types.outgoing-request.set-method if unreachable end
 local.get $request i32.const 64 call $types.outgoing-request.method i32.const 64 i32.load8_u i32.const 2 i32.ne if unreachable end
 local.get $request i32.const 1 i32.const 1056 i32.const 5 call $types.outgoing-request.set-path-with-query if unreachable end
 local.get $request i32.const 1 local.get $secure i32.const 0 i32.const 0 call $types.outgoing-request.set-scheme if unreachable end
 local.get $request i32.const 1 local.get $authority local.get $alen call $types.outgoing-request.set-authority if unreachable end
 local.get $request i32.const 64 call $types.outgoing-request.path-with-query i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $request i32.const 64 call $types.outgoing-request.scheme i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $request i32.const 64 call $types.outgoing-request.authority i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $request call $types.outgoing-request.headers local.set $headers
 local.get $headers i32.const 1024 i32.const 8 i32.const 1040 i32.const 5 i32.const 64 call $types.fields.append
 i32.const 64 i32.load8_u i32.const 1 i32.ne if unreachable end i32.const 65 i32.load8_u i32.const 2 i32.ne if unreachable end
 local.get $headers call $types.drop-fields
 local.get $request i32.const 64 call $types.outgoing-request.body i32.const 64 call $ok i32.const 68 i32.load local.set $body
 local.get $body i32.const 64 call $types.outgoing-body.write i32.const 64 call $ok i32.const 68 i32.load local.set $stream
 local.get $stream i32.const 2048 i32.const 4096 i32.const 64 call $streams.output-stream.blocking-write-and-flush i32.const 64 call $ok
 local.get $stream call $streams.drop-output-stream
 local.get $body i32.const 1 i32.const 1072 i32.const 17 call $new-trailers i32.const 64 call $types.outgoing-body.finish i32.const 64 call $ok
 call $types.new-request-options local.set $options
 local.get $options i32.const 64 call $types.request-options.connect-timeout
 local.get $options i32.const 1 i64.const 1000000000 call $types.request-options.set-connect-timeout if unreachable end
 local.get $options i32.const 1 i64.const 1000000000 call $types.request-options.set-first-byte-timeout if unreachable end
 local.get $options i32.const 1 i64.const 1000000000 call $types.request-options.set-between-bytes-timeout if unreachable end
 local.get $options i32.const 64 call $types.request-options.first-byte-timeout i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $options i32.const 64 call $types.request-options.between-bytes-timeout i32.const 64 i32.load8_u i32.eqz if unreachable end
 local.get $request i32.const 1 local.get $options i32.const 64 call $outgoing-handler.handle i32.const 64 call $ok i32.const 72 i32.load local.set $future
 block $ready loop $next
 local.get $future i32.const 64 call $types.future-incoming-response.get i32.const 64 i32.load8_u br_if $ready
 local.get $future call $types.future-incoming-response.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $next end end
 i32.const 72 i32.load8_u if unreachable end i32.const 80 i32.load8_u if unreachable end i32.const 88 i32.load local.set $response
 local.get $future call $types.drop-future-incoming-response
 local.get $response call $types.incoming-response.status i32.const 200 i32.ne if unreachable end
 local.get $response call $types.incoming-response.headers local.tee $headers call $types.drop-fields
 local.get $response i32.const 64 call $types.incoming-response.consume i32.const 64 call $ok i32.const 68 i32.load local.set $incoming
 local.get $response call $types.drop-incoming-response
 local.get $incoming call $read-body i32.const 4096 i32.ne if unreachable end
 local.get $incoming call $types.incoming-body.finish i32.const 1120 i32.const 15 call $trailers
 i32.const 42)
'''
handler='''(local $headers i32) (local $body i32) (local $response i32) (local $output i32)
 local.get 0 i32.const 64 call $types.incoming-request.method
 local.get 0 i32.const 64 call $types.incoming-request.path-with-query
 local.get 0 i32.const 64 call $types.incoming-request.scheme
 local.get 0 i32.const 64 call $types.incoming-request.authority
 local.get 0 call $types.incoming-request.headers local.tee $headers call $types.drop-fields
 local.get 0 i32.const 64 call $types.incoming-request.consume i32.const 64 call $ok i32.const 68 i32.load local.set $body
 local.get 0 call $types.drop-incoming-request
 local.get $body call $read-body i32.const 4096 i32.ne if unreachable end
 local.get $body call $types.incoming-body.finish i32.const 1072 i32.const 17 call $trailers
 call $types.new-fields call $types.new-outgoing-response local.set $response
 local.get $response i32.const 200 call $types.outgoing-response.set-status-code if unreachable end
 local.get $response call $types.outgoing-response.status-code i32.const 200 i32.ne if unreachable end
 local.get $response call $types.outgoing-response.headers local.tee $headers call $types.drop-fields
 local.get $response i32.const 64 call $types.outgoing-response.body i32.const 64 call $ok i32.const 68 i32.load local.set $body
 local.get 1 i32.const 0 local.get $response i32.const 0 i64.const 0 i32.const 0 i32.const 0 i32.const 0 i32.const 0 call $types.response-outparam.set
 local.get $body i32.const 64 call $types.outgoing-body.write i32.const 64 call $ok i32.const 68 i32.load local.set $output
 local.get $output i32.const 2048 i32.const 4096 i32.const 64 call $streams.output-stream.blocking-write-and-flush i32.const 64 call $ok
 local.get $output call $streams.drop-output-stream
 local.get $body i32.const 1 i32.const 1120 i32.const 15 call $new-trailers i32.const 64 call $types.outgoing-body.finish i32.const 64 call $ok
'''
# The helper receives the original Core parameter indices before named locals.
body('probe','','local.get 0 local.get 1 local.get 2 call $request')
body('request-batch','(local $at i32)','''local.get 2 i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get 2 i32.ge_u br_if $end
 local.get 0 local.get 1 i32.const 0 call $request i32.const 42 i32.ne if unreachable end
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at''')
body('wasi:http/incoming-handler@0.2.12#handle','',handler)
body('cabi_realloc','(local $ptr i32)','global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr')
for name in ['probe','request-batch']:
 index=exports['cabi_post_'+name];s=re.sub(r'\(func \(;'+str(index)+r';\) \(type \d+\) \(param i32\)\)',lambda m:m[0][:-1]+' i32.const 8192 global.set $heap)',s,count=1)
index=exports['cabi_post_wasi:http/incoming-handler@0.2.12#handle'];s=re.sub(r'\(func \(;'+str(index)+r';\) \(type \d+\)\)',lambda m:m[0][:-1]+' i32.const 8192 global.set $heap)',s,count=1)
s=s.replace('  (@custom "component-type"',extra+'\n  (@custom "component-type"')
(out/'http-core.wat').write_text('\n'.join(line.rstrip() for line in s.splitlines())+'\n')
subprocess.run(['wasm-tools','parse',str(out/'http-core.wat'),'-o',str(work/'core.wasm')],check=True)
subprocess.run(['wasm-tools','component','new',str(work/'core.wasm'),'-o',str(out/'http.wasm')],check=True)
subprocess.run(['wasm-tools','validate','--features','all',str(out/'http.wasm')],check=True)
subprocess.run(['wasm-tools','print',str(out/'http.wasm'),'-o',str(out/'http.wat')],check=True)
api=(out/'http.wat').read_text().rstrip()
for module,label in [('wasi:http/types@0.2.12','types'),('wasi:http/outgoing-handler@0.2.12','outgoing'),('wasi:io/poll@0.2.12','poll'),('wasi:io/streams@0.2.12','streams'),('wasi:io/error@0.2.12','errors'),('wasi:cli/stdout@0.2.12','stdout')]:
 index=re.search(r'\(import "'+re.escape(module)+r'" \(instance (?:\$[^ ]+ )?\(;(\d+);\)',api).group(1);api=api[:-1]+f' (export "{label}" (instance {index}))\n)'
(out/'http-api.wat').write_text(api+'\n');subprocess.run(['wasm-tools','parse',str(out/'http-api.wat'),'-o',str(out/'http-api.wasm')],check=True);subprocess.run(['wasm-tools','validate','--features','all',str(out/'http-api.wasm')],check=True)
(out/'http-manifest.json').write_text(json.dumps({'producer':'wasm-tools 1.258.0','wasiVersion':'0.2.12','wasiArchiveSha256':SHA,'worldSha256':hashlib.sha256(world.read_bytes()).hexdigest(),'stableHTTPOperations':53,'unstableFeatures':[],'binarySha256':{n:hashlib.sha256((out/n).read_bytes()).hexdigest() for n in ['http.wasm','http-api.wasm']}},indent=2)+'\n')
print('Reproduced and validated HTTP client/proxy components')
