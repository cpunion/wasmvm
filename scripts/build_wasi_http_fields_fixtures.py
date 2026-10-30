#!/usr/bin/env python3
"""Reproduce HTTP field probes against the complete pinned stable types interface."""
import argparse,hashlib,json,re,shutil,subprocess,tarfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
SHA='2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992'
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--archive',type=Path,default=ROOT/'artifacts/components/wasi-inputs/wasi-wit-0.2.12.tar.gz')
p.add_argument('--work',type=Path,default=ROOT/'artifacts/components/http-fields-producer')
p.add_argument('--output',type=Path,default=ROOT/'pkg/wasi/preview2/testdata')
a=p.parse_args();work,out=a.work,a.output
assert hashlib.sha256(a.archive.read_bytes()).hexdigest()==SHA
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
(work/'wit/deps').mkdir(parents=True,exist_ok=True);out.mkdir(parents=True,exist_ok=True)
with tarfile.open(a.archive) as archive:
 for name in archive.getnames():
  if name.startswith('wasi-wit-0.2.12/http/') and name.endswith('.wit'):(work/'wit/deps'/Path(name).name).write_bytes(archive.extractfile(name).read())
world=ROOT/'pkg/wasi/preview2/testdata/http-fields.wit';shutil.copyfile(world,work/'wit/world.wit')
original=(work/'wit/deps/http.wit').read_text()
subprocess.run(['wasm-tools','component','embed',str(work/'wit'),'--world','http-fields','--dummy-names','legacy','-t','-o',str(work/'dummy.wat')],check=True)
s=(work/'dummy.wat').read_text()
imports=re.findall(r'\(import "([^"]+)" "([^"]+)" \(func \(;(\d+);\)',s)
assert len([x for x in imports if x[0]=='wasi:http/types@0.2.12' and not x[1].startswith('[resource-drop]')])==51
for module,name,index in imports:
 label=name.replace('[method]','').replace('[constructor]','new-').replace('[static]','').replace('[resource-drop]','drop-')
 s=s.replace(f'(func (;{index};) (type',f'(func ${label} (;{index};) (type',1)
exports={name:int(i) for name,i in re.findall(r'\(export "([^"]+)" \(func (\d+)\)\)',s)}
s=s.replace('(memory (;0;) 0)','''(memory (;0;) 1 4)
 (global $heap (mut i32) (i32.const 8192))
 (data (i32.const 1024) "x-key") (data (i32.const 1032) "one")
 (data (i32.const 2048) "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
 (data (i32.const 1040) "X-KEY") (data (i32.const 1056) "bad name")
''')
def body(name,locals,code):
 global s
 index=exports[name];pattern=r'(\(func \(;'+str(index)+r';\)[^\n]*)\n\s+unreachable\n  \)'
 s,n=re.subn(pattern,lambda m:m[1]+' '+locals+'\n'+code+'\n )',s,count=1);assert n==1,name
extra='''(func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
 (func $probe (result i32) (local $h i32) (local $copy i32) (local $other i32) (local $list i32)
 call $new-fields local.set $h
 local.get $h i32.const 1024 i32.const 5 i32.const 1032 i32.const 3 i32.const 64 call $fields.append i32.const 64 call $ok
 i32.const 256 i32.const 1032 i32.store i32.const 260 i32.const 3 i32.store
 i32.const 264 i32.const 0 i32.store i32.const 268 i32.const 0 i32.store
 local.get $h i32.const 1024 i32.const 5 i32.const 256 i32.const 2 i32.const 64 call $fields.set i32.const 64 call $ok
 local.get $h i32.const 1040 i32.const 5 call $fields.has i32.eqz if unreachable end
 local.get $h i32.const 1040 i32.const 5 i32.const 64 call $fields.get
 i32.const 68 i32.load i32.const 2 i32.ne if unreachable end
 i32.const 64 i32.load local.tee $list i32.const 4 i32.add i32.load i32.const 3 i32.ne if unreachable end
 local.get $list i32.const 12 i32.add i32.load if unreachable end
 local.get $h i32.const 64 call $fields.entries
 i32.const 68 i32.load i32.const 2 i32.ne if unreachable end
 local.get $h call $fields.clone local.set $copy
 local.get $copy i32.const 1024 i32.const 5 i32.const 64 call $fields.delete i32.const 64 call $ok
 local.get $copy i32.const 1040 i32.const 5 call $fields.has if unreachable end
 i32.const 288 i32.const 1024 i32.store i32.const 292 i32.const 5 i32.store
 i32.const 296 i32.const 1032 i32.store i32.const 300 i32.const 3 i32.store
 i32.const 288 i32.const 1 i32.const 64 call $fields.from-list i32.const 64 call $ok
 i32.const 68 i32.load local.set $other
 local.get $other i32.const 1056 i32.const 8 i32.const 1032 i32.const 3 i32.const 64 call $fields.append
 i32.const 64 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 65 i32.load8_u if unreachable end
 local.get $h call $drop-fields local.get $copy call $drop-fields local.get $other call $drop-fields
 i32.const 42)
'''
body('probe','','call $probe')
body('fields-batch','(local $at i32)','''local.get 0 i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get 0 i32.ge_u br_if $end
 call $probe i32.const 42 i32.ne if unreachable end
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at''')
body('body-batch','(local $at i32) (local $request i32) (local $body i32) (local $stream i32)','''local.get 0 i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get 0 i32.ge_u br_if $end
 call $new-fields call $new-outgoing-request local.set $request
 local.get $request i32.const 64 call $outgoing-request.body i32.const 64 call $ok i32.const 68 i32.load local.set $body
 local.get $body i32.const 64 call $outgoing-body.write i32.const 64 call $ok i32.const 68 i32.load local.set $stream
 local.get $stream i32.const 2048 i32.const 4096 i32.const 64 call $output-stream.blocking-write-and-flush i32.const 64 call $ok
 local.get $stream call $drop-output-stream
 local.get $body i32.const 0 i32.const 0 i32.const 64 call $outgoing-body.finish i32.const 64 call $ok
 local.get $request call $drop-outgoing-request
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at''')
body('cabi_realloc','(local $ptr i32)','global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr')
for name in ['probe','fields-batch','body-batch']:
 index=exports['cabi_post_'+name]
 s=re.sub(r'\(func \(;'+str(index)+r';\) \(type \d+\) \(param i32\)\)',lambda m:m[0][:-1]+' i32.const 8192 global.set $heap)',s,count=1)
s=s.replace('  (@custom "component-type"',extra+'\n  (@custom "component-type"')
(out/'http-fields-core.wat').write_text('\n'.join(line.rstrip() for line in s.splitlines())+'\n')
subprocess.run(['wasm-tools','parse',str(out/'http-fields-core.wat'),'-o',str(work/'core.wasm')],check=True)
subprocess.run(['wasm-tools','component','new',str(work/'core.wasm'),'-o',str(out/'http-fields.wasm')],check=True)
subprocess.run(['wasm-tools','validate','--features','all',str(out/'http-fields.wasm')],check=True)
subprocess.run(['wasm-tools','print',str(out/'http-fields.wasm'),'-o',str(out/'http-fields.wat')],check=True)
files=['http-fields.wasm']
api=(out/'http-fields.wat').read_text().rstrip()
index=re.search(r'\(import "wasi:http/types@0.2.12" \(instance (?:\$[^ ]+ )?\(;(\d+);\)',api).group(1)
api=api[:-1]+f' (export "types" (instance {index}))\n)'
(out/'http-fields-api.wat').write_text(api+'\n')
subprocess.run(['wasm-tools','parse',str(out/'http-fields-api.wat'),'-o',str(out/'http-fields-api.wasm')],check=True)
subprocess.run(['wasm-tools','validate','--features','all',str(out/'http-fields-api.wasm')],check=True);files.append('http-fields-api.wasm')
cases={
 'tuple-order':('tuple<field-name, field-value>','tuple<field-value, field-name>'),
 'value-width':('type field-value = list<u8>;','type field-value = list<u16>;'),
 'header-error':('    invalid-syntax,','    invalid-header,'),
 'clone-result':('clone: func() -> fields;','clone: func() -> u32;'),
 'lookup-result':('get: func(name: field-name) -> list<field-value>;','get: func(name: field-name) -> field-value;'),
 'request-option':('connect-timeout: func() -> option<duration>;','connect-timeout: func() -> option<u32>;'),
 'request-path':('path-with-query: func() -> option<string>;','path-with-query: func() -> string;'),
 'http-error':('internal-error(option<string>),','internal-error(string),'),
}
for name,(before,after) in cases.items():
 assert before in original,name
 (work/'wit/deps/http.wit').write_text(original.replace(before,after))
 binary=out/f'http-contract-{name}.wasm'
 subprocess.run(['wasm-tools','component','embed',str(work/'wit'),'--world','http-fields','--dummy-names','legacy','-o',str(work/'contract.wasm')],check=True)
 subprocess.run(['wasm-tools','component','new',str(work/'contract.wasm'),'-o',str(binary)],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(binary)],check=True);files.append(binary.name)
(work/'wit/deps/http.wit').write_text(original)
(out/'http-fields-manifest.json').write_text(json.dumps({'producer':'wasm-tools 1.258.0','wasiVersion':'0.2.12','wasiArchiveSha256':SHA,'worldSha256':hashlib.sha256(world.read_bytes()).hexdigest(),'stableTypesOperations':51,'unstableFeatures':[],'scope':'HTTP field workload importing the complete stable types interface; not full HTTP world conformance','binarySha256':{name:hashlib.sha256((out/name).read_bytes()).hexdigest() for name in files}},indent=2)+'\n')
print('Reproduced and validated HTTP field consumers')
