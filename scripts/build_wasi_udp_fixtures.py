#!/usr/bin/env python3
"""Reproduce complete UDP consumers against pinned WASI 0.2.12 WIT."""
import argparse, hashlib, json, re, shutil, subprocess, tarfile
from pathlib import Path
ROOT = Path(__file__).resolve().parents[1]
SHA = '2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992'
p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--archive', type=Path, default=ROOT/'artifacts/components/wasi-inputs/wasi-wit-0.2.12.tar.gz')
p.add_argument('--work', type=Path, default=ROOT/'artifacts/components/udp-producer')
p.add_argument('--output', type=Path, default=ROOT/'pkg/wasi/preview2/testdata')
a = p.parse_args(); work, out = a.work, a.output
assert hashlib.sha256(a.archive.read_bytes()).hexdigest() == SHA
assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'
(work/'wit/deps').mkdir(parents=True, exist_ok=True); out.mkdir(parents=True, exist_ok=True)
with tarfile.open(a.archive) as archive:
 for name in archive.getnames():
  if name.startswith('wasi-wit-0.2.12/cli/deps/') and name.endswith('.wit'):
   (work/'wit/deps'/Path(name).name).write_bytes(archive.extractfile(name).read())
world = ROOT/'pkg/wasi/preview2/testdata/udp.wit'; shutil.copyfile(world, work/'wit/world.wit')
original = (work/'wit/deps/sockets.wit').read_text(); files = []
for common in [False, True]:
 stem = 'udp-common' if common else 'udp'
 (work/'wit/deps/sockets.wit').write_text(original.replace('  network-error-code: func(err: borrow<error>) -> option<error-code>;', '') if common else original)
 subprocess.run(['wasm-tools', 'component', 'embed', str(work/'wit'), '--world', 'udp', '--dummy-names', 'legacy', '-t', '-o', str(work/'dummy.wat')], check=True)
 s = (work/'dummy.wat').read_text()
 imports = re.findall(r'\(import "([^"]+)" "([^"]+)" \(func \(;(\d+);\)', s)
 assert len(imports) == (30 if common else 31)
 for module, name, index in imports:
  label = module.split('/')[1].split('@')[0]+'.'+name.replace('[method]', '').replace('[resource-drop]', 'drop-')
  s = s.replace(f'(func (;{index};) (type', f'(func ${label} (;{index};) (type', 1)
 exports = {name:int(index) for name,index in re.findall(r'\(export "([^"]+)" \(func (\d+)\)\)',s)}
 s = s.replace('(memory (;0;) 0)', '(memory (;0;) 32 64)\n (global $heap (mut i32) (i32.const 8192))\n'+''.join(f' (global ${name} (mut i32) (i32.const 0))\n' for name in ['network','client','server','ci','co','si','so','opened','connected'])+' (data (i32.const 1024) "'+('0123456789abcdef'*256)+'")')
 def body(name, locals, code):
  global s
  index = exports[name]
  pattern = r'(\(func \(;'+str(index)+r';\)[^\n]*)\n\s+unreachable\n  \)'
  s,n = re.subn(pattern,lambda m:m[1]+' '+locals+'\n'+code+'\n )',s,count=1); assert n==1,name
 def address(port):
  v4 = [port,'i32.const 127','i32.const 0','i32.const 0','i32.const 1']+['i32.const 0']*6
  v6 = [port,'i32.const 0']+['i32.const 0']*7+['i32.const 1','i32.const 0']
  return 'local.get $family '+' '.join('local.get $family if (result i32) '+v6[i]+' else '+v4[i]+' end' if v4[i]!=v6[i] else v4[i] for i in range(11))
 options = ''
 for name,value,kind in [('unicast-hop-limit',32,'i32'),('receive-buffer-size',65536,'i64'),('send-buffer-size',65536,'i64')]:
  options += f'global.get $client {kind}.const {value} i32.const 64 call $udp.udp-socket.set-{name} i32.const 64 call $ok\n global.get $client i32.const 64 call $udp.udp-socket.{name} i32.const 64 call $ok\n'
 extra = '''(func $ok (param i32) local.get 0 i32.load8_u if unreachable end)
 (func $open (param $family i32) (param $connected i32) (local $cp i32) (local $sp i32) (local $poll i32)
 global.get $opened if unreachable end
 local.get $connected global.set $connected
 call $instance-network.instance-network global.set $network
 local.get $family i32.const 64 call $udp-create-socket.create-udp-socket i32.const 64 call $ok i32.const 68 i32.load global.set $client
 local.get $family i32.const 64 call $udp-create-socket.create-udp-socket i32.const 64 call $ok i32.const 68 i32.load global.set $server
 global.get $client call $udp.udp-socket.address-family local.get $family i32.ne if unreachable end
 '''+options+'''
 global.get $client global.get $network '''+address('i32.const 0')+''' i32.const 64 call $udp.udp-socket.start-bind i32.const 64 call $ok
 global.get $client i32.const 64 call $udp.udp-socket.finish-bind i32.const 64 call $ok
 global.get $server global.get $network '''+address('i32.const 0')+''' i32.const 64 call $udp.udp-socket.start-bind i32.const 64 call $ok
 global.get $server i32.const 64 call $udp.udp-socket.finish-bind i32.const 64 call $ok
 global.get $client call $udp.udp-socket.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable
 global.get $client i32.const 64 call $udp.udp-socket.local-address i32.const 64 call $ok
 i32.const 256 i32.const 68 i32.const 32 memory.copy i32.const 72 i32.load16_u local.tee $cp i32.eqz if unreachable end
 global.get $server i32.const 64 call $udp.udp-socket.local-address i32.const 64 call $ok
 i32.const 288 i32.const 68 i32.const 32 memory.copy i32.const 72 i32.load16_u local.tee $sp i32.eqz if unreachable end
 global.get $client local.get $connected '''+address('local.get $sp')+''' i32.const 64 call $udp.udp-socket.stream i32.const 64 call $ok
 i32.const 68 i32.load global.set $ci i32.const 72 i32.load global.set $co
 global.get $server local.get $connected '''+address('local.get $cp')+''' i32.const 64 call $udp.udp-socket.stream i32.const 64 call $ok
 i32.const 68 i32.load global.set $si i32.const 72 i32.load global.set $so
 local.get $connected if
 global.get $client i32.const 64 call $udp.udp-socket.remote-address i32.const 64 call $ok
 i32.const 72 i32.load16_u local.get $sp i32.ne if unreachable end end
 global.get $co call $udp.outgoing-datagram-stream.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable
 global.get $si i64.const 0 i32.const 64 call $udp.incoming-datagram-stream.receive i32.const 64 call $ok
 i32.const 72 i32.load if unreachable end
 i32.const 1 global.set $opened)
 (func $send (param $stream i32) (param $peer i32) (param $length i32) (local $poll i32)
 block $permitted loop $again
 local.get $stream i32.const 64 call $udp.outgoing-datagram-stream.check-send i32.const 64 call $ok
 i32.const 72 i64.load i64.const 0 i64.gt_u br_if $permitted
 local.get $stream call $udp.outgoing-datagram-stream.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $again end end
 i32.const 512 i32.const 1024 i32.store i32.const 516 local.get $length i32.store
 i32.const 520 i32.const 1 i32.store8
 i32.const 524 local.get $peer i32.const 32 memory.copy
 local.get $stream i32.const 512 i32.const 1 i32.const 64 call $udp.outgoing-datagram-stream.send i32.const 64 call $ok
 i32.const 72 i64.load i64.const 1 i64.ne if unreachable end)
 (func $receive (param $stream i32) (param $length i32) (local $poll i32) (local $item i32) (local $data i32)
 block $received loop $again
 local.get $stream i64.const 1 i32.const 64 call $udp.incoming-datagram-stream.receive i32.const 64 call $ok
 i32.const 72 i32.load br_if $received
 local.get $stream call $udp.incoming-datagram-stream.subscribe local.tee $poll call $poll.pollable.block local.get $poll call $poll.drop-pollable br $again end end
 i32.const 72 i32.load i32.const 1 i32.ne if unreachable end
 i32.const 68 i32.load local.tee $item i32.const 4 i32.add i32.load local.get $length i32.ne if unreachable end
 local.get $length if
 local.get $item i32.load local.tee $data i32.load8_u i32.const 48 i32.ne if unreachable end
 local.get $data local.get $length i32.add i32.const 1 i32.sub i32.load8_u i32.const 1024 local.get $length i32.const 1 i32.sub i32.const 15 i32.and i32.add i32.load8_u i32.ne if unreachable end end)
 (func $exchange (param $count i32) (result i32) (local $at i32)
 global.get $opened i32.eqz if unreachable end
 local.get $count i32.const 128 i32.gt_u if unreachable end
 block $end loop $next local.get $at local.get $count i32.ge_u br_if $end
 global.get $co i32.const 288 i32.const 4096 call $send global.get $si i32.const 4096 call $receive
 global.get $so i32.const 256 i32.const 4096 call $send global.get $ci i32.const 4096 call $receive
 local.get $at i32.const 1 i32.add local.set $at br $next end end local.get $at)
 (func $close
 global.get $opened i32.eqz if unreachable end
 global.get $ci call $udp.drop-incoming-datagram-stream global.get $co call $udp.drop-outgoing-datagram-stream
 global.get $si call $udp.drop-incoming-datagram-stream global.get $so call $udp.drop-outgoing-datagram-stream
 global.get $client call $udp.drop-udp-socket global.get $server call $udp.drop-udp-socket
 global.get $network call $network.drop-network i32.const 0 global.set $opened)
'''
 body('setup','','local.get 0 local.get 1 call $open i32.const 1')
 body('teardown','','call $close i32.const 1')
 body('transfer','','local.get 0 call $exchange')
 body('exercise','','local.get 0 i32.const 0 call $open global.get $co i32.const 288 i32.const 0 call $send global.get $si i32.const 0 call $receive i32.const 3 call $exchange i32.const 3 i32.ne if unreachable end call $close local.get 0 i32.const 1 call $open i32.const 3 call $exchange i32.const 3 i32.ne if unreachable end call $close i32.const 42')
 body('cabi_realloc','(local $ptr i32)','global.get $heap local.get 2 i32.const 1 i32.sub i32.add i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr local.get 3 i32.add global.set $heap local.get $ptr')
 for name in ['setup','teardown','exercise','transfer']:
  index=exports['cabi_post_'+name];s=s.replace(f'(func (;{index};) (type 1) (param i32))', f'(func (;{index};) (type 1) (param i32) i32.const 8192 global.set $heap)')
 s=s.replace('  (@custom "component-type"',extra+'\n  (@custom "component-type"')
 (out/f'{stem}-core.wat').write_text('\n'.join(line.rstrip() for line in s.splitlines())+'\n')
 subprocess.run(['wasm-tools','parse',str(out/f'{stem}-core.wat'),'-o',str(work/'core.wasm')],check=True)
 subprocess.run(['wasm-tools','component','new',str(work/'core.wasm'),'-o',str(out/f'{stem}.wasm')],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(out/f'{stem}.wasm')],check=True)
 subprocess.run(['wasm-tools','print',str(out/f'{stem}.wasm'),'-o',str(out/f'{stem}.wat')],check=True)
 files.append(stem+'.wasm')
 if not common:
  api=(out/'udp.wat').read_text().rstrip()
  for module,label in [('wasi:sockets/udp@0.2.12','udp'),('wasi:sockets/udp-create-socket@0.2.12','create'),('wasi:sockets/instance-network@0.2.12','instance-network'),('wasi:io/poll@0.2.12','poll'),('wasi:io/error@0.2.12','errors'),('wasi:sockets/network@0.2.12','network')]:
   index=re.search(r'\(import "'+re.escape(module)+r'" \(instance (?:\$[^ ]+ )?\(;(\d+);\)',api).group(1)
   api=api[:-1]+f' (export "{label}" (instance {index}))\n)'
  (out/'udp-api.wat').write_text(api+'\n')
  subprocess.run(['wasm-tools','parse',str(out/'udp-api.wat'),'-o',str(out/'udp-api.wasm')],check=True)
  subprocess.run(['wasm-tools','validate','--features','all',str(out/'udp-api.wasm')],check=True)
  files.append('udp-api.wasm')
cases={
 'address-field':('flow-info: u32,','traffic-info: u32,'),
 'stream-order':('tuple<incoming-datagram-stream, outgoing-datagram-stream>','tuple<outgoing-datagram-stream, incoming-datagram-stream>'),
 'receive-width':('receive: func(max-results: u64)','receive: func(max-results: u32)'),
 'send-width':('send: func(datagrams: list<outgoing-datagram>) -> result<u64,','send: func(datagrams: list<outgoing-datagram>) -> result<u32,'),
 'outgoing-address':('remote-address: option<ip-socket-address>,','peer-address: option<ip-socket-address>,'),
 'incoming-width':('data: list<u8>,','data: list<u16>,'),
 'subscribe-result':('subscribe: func() -> pollable;','subscribe: func() -> u32;'),
 'create-result':('create-udp-socket: func(address-family: ip-address-family) -> result<udp-socket,','create-udp-socket: func(address-family: ip-address-family) -> result<u32,'),
}
for name,(before,after) in cases.items():
 assert before in original,name
 (work/'wit/deps/sockets.wit').write_text(original.replace(before,after))
 binary=out/f'udp-contract-{name}.wasm'
 subprocess.run(['wasm-tools','component','embed',str(work/'wit'),'--world','udp','--dummy-names','legacy','-o',str(work/'contract.wasm')],check=True)
 subprocess.run(['wasm-tools','component','new',str(work/'contract.wasm'),'-o',str(binary)],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(binary)],check=True);files.append(binary.name)
(work/'wit/deps/sockets.wit').write_text(original)
(out/'udp-manifest.json').write_text(json.dumps({'producer':'wasm-tools 1.258.0','wasiVersion':'0.2.12','wasiArchiveSha256':SHA,'worldSha256':hashlib.sha256(world.read_bytes()).hexdigest(),'binarySha256':{name:hashlib.sha256((out/name).read_bytes()).hexdigest() for name in files}},indent=2)+'\n')
print('Reproduced and validated UDP components')
