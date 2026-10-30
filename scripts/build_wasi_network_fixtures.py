#!/usr/bin/env python3
"""Reproduce network/name lookup components with pinned WASI WIT and producer."""
import argparse
import hashlib
import json
import re
import shutil
import subprocess
import sys
import tarfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
SHA='2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992'
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--archive',type=Path,default=ROOT/'artifacts/components/wasi-inputs/wasi-wit-0.2.12.tar.gz')
p.add_argument('--work',type=Path,default=ROOT/'artifacts/components/network-producer')
p.add_argument('--output',type=Path,default=ROOT/'pkg/wasi/preview2/testdata')
p.add_argument('--reference-subset',action='store_true',help='omit unused network-error-code requirement for a narrow consumer')
a=p.parse_args();work=a.work;out=a.output
out.mkdir(parents=True,exist_ok=True)
assert hashlib.sha256(a.archive.read_bytes()).hexdigest()==SHA
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
(work/'wit/deps').mkdir(parents=True,exist_ok=True)
with tarfile.open(a.archive) as archive:
 for name in archive.getnames():
  if name.startswith('wasi-wit-0.2.12/cli/deps/') and name.endswith('.wit'):(work/'wit/deps'/Path(name).name).write_bytes(archive.extractfile(name).read())
world=ROOT/'pkg/wasi/preview2/testdata/network.wit'
shutil.copyfile(world,work/'wit/world.wit')
if a.reference_subset:
 sockets=work/'wit/deps/sockets.wit';content=sockets.read_text();before='  network-error-code: func(err: borrow<error>) -> option<error-code>;';assert before in content;sockets.write_text(content.replace(before,''))
subprocess.run(['wasm-tools','component','embed',str(work/'wit'),'--world','network','--dummy-names','legacy','-t','-o',str(work/'dummy.wat')],check=True)
s=(work/'dummy.wat').read_text()
imports=re.findall(r'\(import "([^"]+)" "([^"]+)" \(func \(;(\d+);\)',s)
assert len(imports)==(12 if a.reference_subset else 13) and imports[-1][1]=='[resource-drop]resolve-address-stream'
exports={n:int(i) for n,i in re.findall(r'\(export "([^"]+)" \(func (\d+)\)\)',s)}
s=s.replace('(memory (;0;) 0)','(memory (;0;) 1)\n (global $heap (mut i32) (i32.const 8192))\n (data (i32.const 1024) "127.0.0.1")\n (data (i32.const 1040) "::1")')
def body(name,signature,code):
 global s
 if a.reference_subset:code=re.sub(r'call (\d+)',lambda m:'call '+str(int(m[1])-int(int(m[1])>6)),code)
 idx=exports[name];s,n=re.subn(r'\(func \(;'+str(idx)+r';\).*?\n  \)',f'(func (;{idx};) {signature}\n{code}\n )',s,count=1,flags=re.S);assert n==1
# Literal lookup includes the complete stream/pollable lifetime and a second
# resolve-next-address call proving end-of-stream. No network packet is sent.
lifetime='''
 local.get $network i32.const 1024 i32.const 9 i32.const 64 call 11 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream call 10 local.tee $poll call 3
 local.get $poll call 2 i32.eqz if unreachable end
 local.get $stream i32.const 64 call 9 i32.const 64 call $ok
 i32.const 66 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load8_u if unreachable end
 i32.const 70 i32.load i32.const 16777343 i32.ne if unreachable end
 local.get $stream i32.const 64 call 9 i32.const 64 call $ok
 i32.const 66 i32.load8_u if unreachable end
 local.get $poll call 5 local.get $stream call 12
'''
body('exercise','(type 4) (result i32) (local $network i32) (local $stream i32) (local $poll i32)','call 8 local.set $network'+lifetime+'''
 local.get $network i32.const 1040 i32.const 3 i32.const 64 call 11 i32.const 64 call $ok
 i32.const 68 i32.load local.set $stream
 local.get $stream call 10 local.tee $poll call 3
 local.get $stream i32.const 64 call 9 i32.const 64 call $ok
 i32.const 66 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 84 i32.load16_u i32.const 1 i32.ne if unreachable end
 local.get $poll call 5 local.get $stream call 12 local.get $network call 7 i32.const 42
''')
body('lookup-batch','(type 2) (param $count i32) (result i32) (local $network i32) (local $stream i32) (local $poll i32) (local $at i32)','''call 8 local.set $network
 block $end loop $next local.get $at local.get $count i32.ge_u br_if $end
'''+lifetime+'''local.get $at i32.const 1 i32.add local.set $at br $next end end
 local.get $network call 7 local.get $at''')
body('cabi_realloc','(type 6) (param i32 i32 i32 i32) (result i32) (local $ptr i32)','global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr')
for name in ['exercise','lookup-batch']:
 idx=exports['cabi_post_'+name];s=s.replace(f'(func (;{idx};) (type 1) (param i32))',f'(func (;{idx};) (type 1) (param i32) i32.const 8192 global.set $heap)')
s=s.replace('  (@custom "component-type"','  (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)\n  (@custom "component-type"')
(out/'network-core.wat').write_text(s)
subprocess.run(['wasm-tools','parse',str(out/'network-core.wat'),'-o',str(work/'core.wasm')],check=True)
subprocess.run(['wasm-tools','component','new',str(work/'core.wasm'),'-o',str(out/'network.wasm')],check=True)
subprocess.run(['wasm-tools','validate','--features','all',str(out/'network.wasm')],check=True)
subprocess.run(['wasm-tools','print',str(out/'network.wasm'),'-o',str(out/'network.wat')],check=True)
api=(out/'network.wat').read_text().rstrip()
for name,label in [('wasi:sockets/network@0.2.12','network'),('wasi:sockets/instance-network@0.2.12','instance-network'),('wasi:sockets/ip-name-lookup@0.2.12','lookup'),('wasi:io/error@0.2.12','errors'),('wasi:io/poll@0.2.12','poll')]:
 idx=re.search(r'\(import "'+re.escape(name)+r'" \(instance (?:\$[^ ]+ )?\(;(\d+);\)',api).group(1)
 api=api[:-1]+f' (export "{label}" (instance {idx}))\n)'
(out/'network-api.wat').write_text(api+'\n')
subprocess.run(['wasm-tools','parse',str(out/'network-api.wat'),'-o',str(out/'network-api.wasm')],check=True)
subprocess.run(['wasm-tools','validate','--features','all',str(out/'network-api.wasm')],check=True)
cases={
 'error-label':('    unknown,','    unexpected,'),
 'ip-label':('    ipv4(ipv4-address),','    v4(ipv4-address),'),
 'ip-width':('tuple<u16, u16, u16, u16, u16, u16, u16, u16>', 'tuple<u32, u16, u16, u16, u16, u16, u16, u16>'),
 'lookup-name':('name: string) -> result<resolve-address-stream,', 'name: u32) -> result<resolve-address-stream,'),
 'subscribe-result':('subscribe: func() -> pollable;', 'subscribe: func() -> u32;'),
 'network-result':('instance-network: func() -> network;', 'instance-network: func() -> u32;'),
}
contracts=work/'contract-wit'
if contracts.exists():shutil.rmtree(contracts)
shutil.copytree(work/'wit',contracts)
original=(contracts/'deps/sockets.wit').read_text()
for name,(before,after) in ({} if a.reference_subset else cases).items():
 assert before in original,(name,before)
 (contracts/'deps/sockets.wit').write_text(original.replace(before,after))
 binary=out/f'network-contract-{name}.wasm'
 subprocess.run(['wasm-tools','component','embed',str(contracts),'--world','network','--dummy-names','legacy','-o',str(work/'contract-core.wasm')],check=True)
 subprocess.run(['wasm-tools','component','new',str(work/'contract-core.wasm'),'-o',str(binary)],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(binary)],check=True)
(contracts/'deps/sockets.wit').write_text(original)
files=['network.wasm','network-api.wasm']+[f'network-contract-{name}.wasm' for name in ({} if a.reference_subset else cases)]
if not a.reference_subset:
 common=work/'reference-output'
 subprocess.run([sys.executable,__file__,'--archive',str(a.archive),'--work',str(work/'reference-subset'),'--output',str(common),'--reference-subset'],check=True)
 for suffix in ['wasm','wat','core.wat']:
  source=common/('network-core.wat' if suffix=='core.wat' else 'network.'+suffix)
  target=out/('network-common-core.wat' if suffix=='core.wat' else 'network-common.'+suffix)
  shutil.copyfile(source,target)
 files.append('network-common.wasm')
(out/'network-manifest.json').write_text(json.dumps({'producer':'wasm-tools 1.258.0','wasiVersion':'0.2.12','wasiArchiveSha256':SHA,'worldSha256':hashlib.sha256(world.read_bytes()).hexdigest(),'binarySha256':{name:hashlib.sha256((out/name).read_bytes()).hexdigest() for name in files}},indent=2)+'\n')
print('Reproduced and validated network components')
