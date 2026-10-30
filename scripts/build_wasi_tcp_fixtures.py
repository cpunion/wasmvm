#!/usr/bin/env python3
"""Reproduce full TCP consumers from pinned WASI WIT and authored Core code."""
import argparse,hashlib,json,re,shutil,subprocess,tarfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
SHA='2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992'
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--archive',type=Path,default=ROOT/'artifacts/components/wasi-inputs/wasi-wit-0.2.12.tar.gz')
p.add_argument('--work',type=Path,default=ROOT/'artifacts/components/tcp-producer')
p.add_argument('--output',type=Path,default=ROOT/'pkg/wasi/preview2/testdata')
a=p.parse_args();out=a.output;work=a.work
assert hashlib.sha256(a.archive.read_bytes()).hexdigest()==SHA
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
(work/'wit/deps').mkdir(parents=True,exist_ok=True);out.mkdir(parents=True,exist_ok=True)
with tarfile.open(a.archive) as archive:
 for name in archive.getnames():
  if name.startswith('wasi-wit-0.2.12/cli/deps/') and name.endswith('.wit'):(work/'wit/deps'/Path(name).name).write_bytes(archive.extractfile(name).read())
world=ROOT/'pkg/wasi/preview2/testdata/tcp.wit';shutil.copyfile(world,work/'wit/world.wit')
original=(work/'wit/deps/sockets.wit').read_text()
files=[]
for common in [False,True]:
 stem='tcp-common' if common else 'tcp'
 (work/'wit/deps/sockets.wit').write_text(original.replace('  network-error-code: func(err: borrow<error>) -> option<error-code>;','') if common else original)
 subprocess.run(['wasm-tools','component','embed',str(work/'wit'),'--world','tcp','--dummy-names','legacy','-t','-o',str(work/'dummy.wat')],check=True)
 s=(work/'dummy.wat').read_text()
 imports=re.findall(r'\(import "([^"]+)" "([^"]+)" \(func \(;(\d+);\)',s)
 assert len(imports)==(59 if common else 60)
 aliases={}
 for module,name,index in imports:
  alias=('tcp.'+name.removeprefix('[method]tcp-socket.') if '[method]tcp-socket.' in name else module.split('/')[1].split('@')[0]+'.'+name).replace('[resource-drop]','drop-').replace('[method]','')
  aliases[alias]=int(index)
  s=s.replace(f'(func (;{index};) (type',f'(func ${alias} (;{index};) (type',1)
 exports={n:int(i) for n,i in re.findall(r'\(export "([^"]+)" \(func (\d+)\)\)',s)}
 s=s.replace('(memory (;0;) 0)','(memory (;0;) 32 64)\n (global $heap (mut i32) (i32.const 8192))\n'+''.join(f' (global ${name} (mut i32) (i32.const 0))\n' for name in ['network','server','client','child','ci','co','si','so','opened'])+' (data (i32.const 1024) "'+('0123456789abcdef'*256)+'")')
 def body(name,locals,code):
  global s
  index=exports[name];pattern=r'(\(func \(;'+str(index)+r';\)[^\n]*)\n\s+unreachable\n  \)'
  s,n=re.subn(pattern,lambda m:m[1]+' '+locals+'\n'+code+'\n )',s,count=1);assert n==1,name
 # Start operations flatten the entire address variant (12 slots), retaining
 # explicit family/port. IPv4 occupies five payload slots, IPv6 eleven.
 # Explicit payload assembly avoids variant joins being inferred from a host ABI.
 def address(port):
  v4=[port,'i32.const 127','i32.const 0','i32.const 0','i32.const 1']+['i32.const 0']*6
  v6=[port,'i32.const 0']+['i32.const 0']*7+['i32.const 1','i32.const 0']
  return 'local.get $family '+' '.join('local.get $family if (result i32) '+v6[i]+' else '+v4[i]+' end' if v4[i]!=v6[i] else v4[i] for i in range(11))
 options=''
 for name,arg,kind in [('keep-alive-enabled',1,'i32'),('keep-alive-idle-time',1000000000,'i64'),('keep-alive-interval',1000000000,'i64'),('keep-alive-count',3,'i32'),('hop-limit',32,'i32'),('receive-buffer-size',65536,'i64'),('send-buffer-size',65536,'i64')]:
  options+=f'global.get $client {kind}.const {arg} i32.const 64 call $tcp.set-{name} i32.const 64 call $ok\n global.get $client i32.const 64 call $tcp.{name} i32.const 64 call $ok\n'
 open_code='''(func $open (param $family i32) (local $poll i32) (local $port i32)
 global.get $opened if unreachable end
 call $instance-network.instance-network global.set $network
 local.get $family i32.const 64 call $tcp-create-socket.create-tcp-socket i32.const 64 call $ok i32.const 68 i32.load global.set $server
 local.get $family i32.const 64 call $tcp-create-socket.create-tcp-socket i32.const 64 call $ok i32.const 68 i32.load global.set $client
 global.get $client call $tcp.address-family local.get $family i32.ne if unreachable end
 global.get $client call $tcp.is-listening if unreachable end
'''+options+'''
 global.get $server i64.const 128 i32.const 64 call $tcp.set-listen-backlog-size i32.const 64 call $ok
 global.get $server global.get $network '''+address('i32.const 0')+''' i32.const 64 call $tcp.start-bind i32.const 64 call $ok
 global.get $server i32.const 64 call $tcp.finish-bind i32.const 64 call $ok
 global.get $server i32.const 64 call $tcp.start-listen i32.const 64 call $ok
 global.get $server i32.const 64 call $tcp.finish-listen i32.const 64 call $ok
 global.get $server call $tcp.is-listening i32.eqz if unreachable end
 global.get $server i32.const 64 call $tcp.local-address i32.const 64 call $ok
 i32.const 72 i32.load16_u local.tee $port i32.eqz if unreachable end
 global.get $client global.get $network '''+address('local.get $port')+''' i32.const 64 call $tcp.start-connect i32.const 64 call $ok
 block $connected loop $connect
 global.get $client i32.const 64 call $tcp.finish-connect
 i32.const 64 i32.load8_u i32.eqz br_if $connected
 i32.const 68 i32.load8_u i32.const 8 i32.ne if unreachable end
 global.get $client call $tcp.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $connect end end
 i32.const 68 i32.load global.set $ci i32.const 72 i32.load global.set $co
 block $accepted loop $accept
 global.get $server i32.const 64 call $tcp.accept
 i32.const 64 i32.load8_u i32.eqz br_if $accepted
 i32.const 68 i32.load8_u i32.const 8 i32.ne if unreachable end
 global.get $server call $tcp.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $accept end end
 i32.const 68 i32.load global.set $child i32.const 72 i32.load global.set $si i32.const 76 i32.load global.set $so
 global.get $client i32.const 64 call $tcp.remote-address i32.const 64 call $ok
 global.get $child i32.const 64 call $tcp.local-address i32.const 64 call $ok
 i32.const 1 global.set $opened)
'''
 read_code='''(func $read-all (param $stream i32) (local $at i32) (local $n i32) (local $ptr i32)
 block $end loop $next local.get $at i32.const 4096 i32.ge_u br_if $end
 local.get $stream i32.const 4096 local.get $at i32.sub i64.extend_i32_u i32.const 64 call $streams.input-stream.blocking-read i32.const 64 call $ok
 i32.const 68 i32.load local.set $ptr i32.const 72 i32.load local.tee $n i32.eqz if unreachable end
 local.get $ptr i32.load8_u i32.const 1024 local.get $at i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end
 local.get $ptr local.get $n i32.add i32.const 1 i32.sub i32.load8_u i32.const 1024 local.get $at local.get $n i32.add i32.const 1 i32.sub i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end
 local.get $at local.get $n i32.add local.set $at br $next end end)
 (func $exchange (param $count i32) (result i32) (local $at i32)
 global.get $opened i32.eqz if unreachable end
 local.get $count i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get $count i32.ge_u br_if $end
 global.get $co i32.const 1024 i32.const 4096 i32.const 64 call $streams.output-stream.blocking-write-and-flush i32.const 64 call $ok
 global.get $si call $read-all
 global.get $so i32.const 1024 i32.const 4096 i32.const 64 call $streams.output-stream.blocking-write-and-flush i32.const 64 call $ok
 global.get $ci call $read-all
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at)
 (func $close
 global.get $opened i32.eqz if unreachable end
 global.get $client i32.const 2 i32.const 64 call $tcp.shutdown i32.const 64 call $ok
 global.get $child i32.const 2 i32.const 64 call $tcp.shutdown i32.const 64 call $ok
 global.get $ci call $streams.drop-input-stream global.get $co call $streams.drop-output-stream
 global.get $si call $streams.drop-input-stream global.get $so call $streams.drop-output-stream
 global.get $client call $tcp.drop-tcp-socket global.get $child call $tcp.drop-tcp-socket global.get $server call $tcp.drop-tcp-socket
 global.get $network call $network.drop-network i32.const 0 global.set $opened)
 (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
'''
 body('create-probe','', 'local.get 0 i32.const 64 call $tcp-create-socket.create-tcp-socket i32.const 64 i32.load8_u if (result i32) i32.const 68 i32.load8_u i32.const 1 i32.add else i32.const 68 i32.load call $tcp.drop-tcp-socket i32.const 0 end')
 body('setup','', 'local.get 0 call $open i32.const 1')
 body('teardown','', 'call $close i32.const 1')
 body('exercise','', 'local.get 0 call $open i32.const 3 call $exchange i32.const 3 i32.ne if unreachable end call $close i32.const 42')
 body('transfer','', 'local.get 0 call $exchange')
 body('cabi_realloc','(local $ptr i32)','global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr')
 for name in ['create-probe','setup','teardown','exercise','transfer']:
  idx=exports['cabi_post_'+name];s=s.replace(f'(func (;{idx};) (type 1) (param i32))',f'(func (;{idx};) (type 1) (param i32) i32.const 8192 global.set $heap)')
 s=s.replace('  (@custom "component-type"',open_code+read_code+'\n  (@custom "component-type"')
 (out/f'{stem}-core.wat').write_text('\n'.join(line.rstrip() for line in s.splitlines())+'\n')
 subprocess.run(['wasm-tools','parse',str(out/f'{stem}-core.wat'),'-o',str(work/'core.wasm')],check=True)
 subprocess.run(['wasm-tools','component','new',str(work/'core.wasm'),'-o',str(out/f'{stem}.wasm')],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(out/f'{stem}.wasm')],check=True)
 subprocess.run(['wasm-tools','print',str(out/f'{stem}.wasm'),'-o',str(out/f'{stem}.wat')],check=True)
 files.append(stem+'.wasm')
 if not common:
  api=(out/'tcp.wat').read_text().rstrip()
  for module,label in [('wasi:sockets/tcp@0.2.12','tcp'),('wasi:sockets/tcp-create-socket@0.2.12','create'),('wasi:sockets/instance-network@0.2.12','instance-network'),('wasi:io/streams@0.2.12','streams'),('wasi:io/poll@0.2.12','poll'),('wasi:io/error@0.2.12','errors'),('wasi:sockets/network@0.2.12','network')]:
   index=re.search(r'\(import "'+re.escape(module)+r'" \(instance (?:\$[^ ]+ )?\(;(\d+);\)',api).group(1);api=api[:-1]+f' (export "{label}" (instance {index}))\n)'
  (out/'tcp-api.wat').write_text(api+'\n');subprocess.run(['wasm-tools','parse',str(out/'tcp-api.wat'),'-o',str(out/'tcp-api.wasm')],check=True);subprocess.run(['wasm-tools','validate','--features','all',str(out/'tcp-api.wasm')],check=True);files.append('tcp-api.wasm')
cases={
 'address-field':('flow-info: u32,','traffic-info: u32,'),
 'address-width':('port: u16,','port: u32,'),
 'stream-order':('finish-connect: func() -> result<tuple<input-stream, output-stream>', 'finish-connect: func() -> result<tuple<output-stream, input-stream>'),
 'keep-count':('keep-alive-count: func() -> result<u32,','keep-alive-count: func() -> result<u64,'),
 'shutdown-label':('    both,','    all,'),
 'family-result':('address-family: func() -> ip-address-family;','address-family: func() -> u32;'),
 'subscribe-result':('subscribe: func() -> pollable;','subscribe: func() -> u32;'),
 'create-result':('create-tcp-socket: func(address-family: ip-address-family) -> result<tcp-socket,','create-tcp-socket: func(address-family: ip-address-family) -> result<u32,'),
}
for name,(before,after) in cases.items():
 assert before in original,name
 (work/'wit/deps/sockets.wit').write_text(original.replace(before,after))
 binary=out/f'tcp-contract-{name}.wasm'
 subprocess.run(['wasm-tools','component','embed',str(work/'wit'),'--world','tcp','--dummy-names','legacy','-o',str(work/'contract.wasm')],check=True)
 subprocess.run(['wasm-tools','component','new',str(work/'contract.wasm'),'-o',str(binary)],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(binary)],check=True);files.append(binary.name)
(work/'wit/deps/sockets.wit').write_text(original)
(out/'tcp-manifest.json').write_text(json.dumps({'producer':'wasm-tools 1.258.0','wasiVersion':'0.2.12','wasiArchiveSha256':SHA,'worldSha256':hashlib.sha256(world.read_bytes()).hexdigest(),'binarySha256':{f:hashlib.sha256((out/f).read_bytes()).hexdigest() for f in files}},indent=2)+'\n')
print('Reproduced and validated TCP components')
