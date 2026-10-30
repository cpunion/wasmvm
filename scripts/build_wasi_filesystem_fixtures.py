#!/usr/bin/env python3
"""Reproduce filesystem components from pinned WASI WIT and authored Core code."""
import argparse
import hashlib
import json
import re
import shutil
import subprocess
import tarfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
SHA="2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992"
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--archive',type=Path,default=ROOT/'artifacts/components/wasi-inputs/wasi-wit-0.2.12.tar.gz')
p.add_argument('--work',type=Path,default=ROOT/'artifacts/components/filesystem-producer')
p.add_argument('--output',type=Path,default=ROOT/'pkg/wasi/preview2/testdata')
a=p.parse_args();work=a.work;out=a.output
assert hashlib.sha256(a.archive.read_bytes()).hexdigest()==SHA
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
(work/'wit/deps').mkdir(parents=True,exist_ok=True)
with tarfile.open(a.archive) as archive:
 for name in archive.getnames():
  if name.startswith('wasi-wit-0.2.12/cli/deps/') and name.endswith('.wit'):
   (work/'wit/deps'/Path(name).name).write_bytes(archive.extractfile(name).read())
shutil.copyfile(out/'filesystem.wit',work/'wit/world.wit')
subprocess.run(['wasm-tools','component','embed',str(work/'wit'),'--world','filesystem','--dummy-names','legacy','-t','-o',str(work/'dummy.wat')],check=True)
s=(work/'dummy.wat').read_text()
imports={(m,n):int(i) for m,n,i in re.findall(r'\(import "([^"]+)" "([^"]+)" \(func \(;(\d+);\)',s)}
assert len(imports)==57
exports={n:int(i) for n,i in re.findall(r'\(export "([^"]+)" \(func (\d+)\)\)',s)}
s=s.replace('(memory (;0;) 0)','(memory (;0;) 1)\n (global $heap (mut i32) (i32.const 8192))\n (data (i32.const 1024) "alpha")\n (data (i32.const 1040) "beta")\n (data (i32.const 1056) "gamma")\n (data (i32.const 1072) "sym")\n (data (i32.const 1088) "sub")\n (data (i32.const 1104) "abcdEF!")\n (data (i32.const 1120) "bench.bin")')
def body(name,signature,code):
 global s
 idx=exports[name]
 s,n=re.subn(r'\(func \(;'+str(idx)+r';\).*?\n  \)',f'(func (;{idx};) {signature}\n{code}\n )',s,count=1,flags=re.S)
 assert n==1
get='i32.const 0 call 56 i32.const 4 i32.load i32.const 1 i32.ne if unreachable end i32.const 0 i32.load i32.load local.set $root'
# Every assertion runs inside the guest, through real canonical ABI lowering.
body('exercise','(type 17) (result i32) (local $root i32) (local $file i32) (local $stream i32) (local $poll i32) (local $entries i32) (local $count i32)',get+'''
 local.get $root i32.const 64 call 30 i32.const 64 call $ok
 i32.const 65 i32.load8_u i32.const 1 i32.and i32.eqz if unreachable end
 local.get $root i32.const 64 call 31 i32.const 64 call $ok
 i32.const 65 i32.load8_u i32.const 3 i32.ne if unreachable end
 local.get $root i32.const 1088 i32.const 3 i32.const 64 call 38 i32.const 64 call $ok
 local.get $root i32.const 0 i32.const 1024 i32.const 5 i32.const 1 i32.const 3 i32.const 64 call 43 i32.const 64 call $ok
 i32.const 68 i32.load local.set $file
 local.get $file i32.const 1104 i32.const 4 i64.const 0 i32.const 64 call 35 i32.const 64 call $ok
 i32.const 72 i64.load i64.const 4 i64.ne if unreachable end
 local.get $file i64.const 4 i64.const 0 i32.const 64 call 34 i32.const 64 call $ok
 i32.const 72 i32.load i32.const 4 i32.ne if unreachable end
 i32.const 68 i32.load i32.load i32.const 1684234849 i32.ne if unreachable end
 local.get $file i64.const 8 i32.const 64 call 32 i32.const 64 call $ok
 local.get $file i32.const 64 call 39 i32.const 64 call $ok
 i32.const 88 i64.load i64.const 8 i64.ne if unreachable end
 local.get $file i32.const 64 call 31 i32.const 64 call $ok
 i32.const 65 i32.load8_u i32.const 6 i32.ne if unreachable end
 local.get $file i32.const 64 call 30 i32.const 64 call $ok
 i32.const 65 i32.load8_u i32.const 3 i32.and i32.const 3 i32.ne if unreachable end
 local.get $file i64.const 0 i64.const 8 i32.const 0 i32.const 64 call 28 i32.const 64 call $ok
 local.get $file i32.const 64 call 29 i32.const 64 call $ok
 local.get $file i32.const 64 call 37 i32.const 64 call $ok
 local.get $file i32.const 1 i64.const 0 i32.const 0 i32.const 1 i64.const 0 i32.const 0 i32.const 64 call 33 i32.const 64 call $ok
 local.get $file i64.const 0 i32.const 64 call 25 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream call 10 local.tee $poll call 3 local.get $poll call 5
 local.get $stream i64.const 4 i32.const 64 call 7 i32.const 64 call $ok
 i32.const 72 i32.load i32.const 4 i32.ne if unreachable end
 local.get $stream call 21
 local.get $file i64.const 4 i32.const 64 call 26 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream i32.const 1108 i32.const 2 i32.const 64 call 13 i32.const 64 call $ok
 local.get $stream call 22
 local.get $file i32.const 64 call 27 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream i32.const 1110 i32.const 1 i32.const 64 call 13 i32.const 64 call $ok
 local.get $stream call 22
 local.get $root i32.const 0 i32.const 1024 i32.const 5 i32.const 64 call 40 i32.const 64 call $ok
 i32.const 88 i64.load i64.const 9 i64.ne if unreachable end
 local.get $root i32.const 0 i32.const 1024 i32.const 5 i32.const 1 i64.const 0 i32.const 0 i32.const 1 i64.const 0 i32.const 0 i32.const 64 call 41 i32.const 64 call $ok
 local.get $root i32.const 0 i32.const 1024 i32.const 5 local.get $root i32.const 1040 i32.const 4 i32.const 64 call 42 i32.const 64 call $ok
 local.get $root i32.const 1040 i32.const 4 local.get $root i32.const 1056 i32.const 5 i32.const 64 call 46 i32.const 64 call $ok
 local.get $root i32.const 1024 i32.const 5 i32.const 1072 i32.const 3 i32.const 64 call 47 i32.const 64 call $ok
 local.get $root i32.const 1072 i32.const 3 i32.const 64 call 44 i32.const 64 call $ok
 i32.const 72 i32.load i32.const 5 i32.ne if unreachable end
 local.get $root i32.const 1 i32.const 1072 i32.const 3 i32.const 0 i32.const 1 i32.const 64 call 43 i32.const 64 call $ok
 local.get $file i32.const 68 i32.load call 49 i32.eqz if unreachable end
 i32.const 68 i32.load call 54
 local.get $file i32.const 64 call 50 i32.const 64 call $ok
 local.get $root i32.const 0 i32.const 1024 i32.const 5 i32.const 128 call 51 i32.const 128 call $ok
 i32.const 72 i64.load i32.const 136 i64.load i64.ne if unreachable end
 i32.const 80 i64.load i32.const 144 i64.load i64.ne if unreachable end
 local.get $root i32.const 64 call 36 i32.const 64 call $ok
 i32.const 68 i32.load local.set $entries
 block $end loop $next
 local.get $entries i32.const 64 call 52 i32.const 64 call $ok
 i32.const 68 i32.load8_u i32.eqz br_if $end
 local.get $count i32.const 1 i32.add local.set $count br $next end end
 local.get $count i32.const 4 i32.lt_u if unreachable end
 local.get $entries call 55
 local.get $root i32.const 1056 i32.const 5 i32.const 64 call 48 i32.const 64 call $ok
 local.get $root i32.const 1072 i32.const 3 i32.const 64 call 48 i32.const 64 call $ok
 local.get $file call 54
 local.get $root i32.const 1024 i32.const 5 i32.const 64 call 48 i32.const 64 call $ok
 local.get $root i32.const 1088 i32.const 3 i32.const 64 call 45 i32.const 64 call $ok
 local.get $root call 54 i32.const 42
''')
body('driver','(type 17) (result i32) (local $root i32)',get+' local.get $root call 54 i32.const 1')
body('stat-batch','(type 2) (param $count i32) (result i32) (local $root i32) (local $file i32) (local $at i32)',get+'''
 local.get $root i32.const 0 i32.const 1120 i32.const 9 i32.const 0 i32.const 1 i32.const 64 call 43 i32.const 64 call $ok
 i32.const 68 i32.load local.set $file
 block $end loop $next
 local.get $at local.get $count i32.ge_u br_if $end
 local.get $file i32.const 64 call 39 i32.const 64 call $ok
 i32.const 88 i64.load i64.const 16 i64.ne if unreachable end
 local.get $at i32.const 1 i32.add local.set $at br $next end end
 local.get $file call 54 local.get $root call 54 local.get $at
''')
# The host test revokes the filesystem owner after creating a read stream.
# The guest then verifies IO error provenance and drops the error and streams.
body('error-probe','(type 17) (result i32) (local $root i32) (local $file i32) (local $stream i32) (local $error i32)',get+'''
 local.get $root i32.const 0 i32.const 1120 i32.const 9 i32.const 0 i32.const 1 i32.const 64 call 43 i32.const 64 call $ok
 i32.const 68 i32.load local.set $file
 local.get $file i64.const 0 i32.const 64 call 25 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream i64.const 1 i32.const 64 call 6
 i32.const 64 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load8_u if unreachable end
 i32.const 72 i32.load local.set $error
 local.get $error i32.const 128 call 53
 i32.const 128 i32.load8_u i32.const 1 i32.ne if unreachable end
 local.get $error call 1 local.get $stream call 21
 local.get $file call 54 local.get $root call 54 i32.const 1
''')
body('cabi_realloc','(type 18) (param i32 i32 i32 i32) (result i32) (local $ptr i32)','''global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr''')
for name in ['exercise','driver','stat-batch','error-probe']:
 idx=exports['cabi_post_'+name]
 s=s.replace(f'(func (;{idx};) (type 1) (param i32))',f'(func (;{idx};) (type 1) (param i32) i32.const 8192 global.set $heap)')
s=s.replace('  (@custom "component-type"','  (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)\n  (@custom "component-type"')
core=out/'filesystem-core.wat';core.write_text(s)
subprocess.run(['wasm-tools','parse',str(core),'-o',str(work/'core.wasm')],check=True)
subprocess.run(['wasm-tools','component','new',str(work/'core.wasm'),'-o',str(out/'filesystem.wasm')],check=True)
subprocess.run(['wasm-tools','validate','--features','all',str(out/'filesystem.wasm')],check=True)
subprocess.run(['wasm-tools','print',str(out/'filesystem.wasm'),'-o',str(out/'filesystem.wat')],check=True)
# Reexport the same nominal imported instances for direct embedding contract
# tests. The main fixture above still executes every filesystem operation in Core.
api=(out/'filesystem.wat').read_text().rstrip()
for name,label in [('wasi:filesystem/types@0.2.12','types'),('wasi:filesystem/preopens@0.2.12','preopens'),('wasi:io/streams@0.2.12','streams'),('wasi:io/error@0.2.12','errors')]:
 idx=re.search(r'\(import "'+re.escape(name)+r'" \(instance (?:\$[^ ]+ )?\(;(\d+);\)',api).group(1)
 api=api[:-1]+f' (export "{label}" (instance {idx}))\n)'
(out/'filesystem-api.wat').write_text(api+'\n')
subprocess.run(['wasm-tools','parse',str(out/'filesystem-api.wat'),'-o',str(out/'filesystem-api.wasm')],check=True)
subprocess.run(['wasm-tools','validate','--features','all',str(out/'filesystem-api.wasm')],check=True)
# Each negative fixture is valid Component Model input with one mismatched
# provider contract. The body remains unreachable because linking must fail.
cases={
 'read-width':('read: func(length: filesize, offset: filesize)', 'read: func(length: u32, offset: filesize)'),
 'read-result':('tuple<list<u8>, bool>', 'tuple<list<u8>, u32>'),
 'stat-field':('size: filesize,', 'length: filesize,'),
 'error-label':('    access,', '    forbidden,'),
 'timestamp-label':('    no-change,', '    keep,'),
 'advice-label':('    no-reuse,', '    never-reuse,'),
 'directory-entry':('    name: string,', '    name: u32,'),
 'stream-resource':('read-via-stream: func(offset: filesize) -> result<input-stream,', 'read-via-stream: func(offset: filesize) -> result<output-stream,'),
 'preopen-order':('list<tuple<descriptor, string>>', 'list<tuple<string, descriptor>>'),
}
contracts=work/'contract-wit'
if contracts.exists():shutil.rmtree(contracts)
shutil.copytree(work/'wit',contracts)
original=(contracts/'deps/filesystem.wit').read_text()
for name,(before,after) in cases.items():
 assert before in original,(name,before)
 (contracts/'deps/filesystem.wit').write_text(original.replace(before,after))
 binary=out/f'filesystem-contract-{name}.wasm'
 subprocess.run(['wasm-tools','component','embed',str(contracts),'--world','filesystem','--dummy-names','legacy','-o',str(work/'contract-core.wasm')],check=True)
 subprocess.run(['wasm-tools','component','new',str(work/'contract-core.wasm'),'-o',str(binary)],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(binary)],check=True)
(contracts/'deps/filesystem.wit').write_text(original)
files=['filesystem.wasm','filesystem-api.wasm']+[f'filesystem-contract-{name}.wasm' for name in cases]
(out/'filesystem-manifest.json').write_text(json.dumps({'producer':'wasm-tools 1.258.0','wasiVersion':'0.2.12','wasiArchiveSha256':SHA,'worldSha256':hashlib.sha256((out/'filesystem.wit').read_bytes()).hexdigest(),'binarySha256':{name:hashlib.sha256((out/name).read_bytes()).hexdigest() for name in files}},indent=2)+'\n')
print('Reproduced and validated filesystem components')
