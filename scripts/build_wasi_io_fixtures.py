#!/usr/bin/env python3
"""Reproduce authored WASI IO components with the pinned WIT/producer."""
import argparse
import hashlib
import json
import re
import shutil
import subprocess
import tarfile
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]
ARCHIVE_SHA256="2b0e47bc02953267ad163c6c1188b65baed1292351998cb032f3dcd28d005992"
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument("--archive",type=Path,default=ROOT/"artifacts/components/wasi-inputs/wasi-wit-0.2.12.tar.gz")
parser.add_argument("--output",type=Path,default=ROOT/"pkg/wasi/preview2/testdata")
parser.add_argument("--work",type=Path,default=ROOT/"artifacts/components/io-producer")
args=parser.parse_args()
assert hashlib.sha256(args.archive.read_bytes()).hexdigest()==ARCHIVE_SHA256,"WASI archive checksum differs"
assert subprocess.check_output(["wasm-tools","--version"],text=True).strip()=="wasm-tools 1.258.0","producer version differs"
out=args.work;output=args.output
(out/"wit/deps").mkdir(parents=True,exist_ok=True);output.mkdir(parents=True,exist_ok=True)
with tarfile.open(args.archive) as archive:
 for name in archive.getnames():
  if name.startswith("wasi-wit-0.2.12/cli/deps/") and name.endswith(".wit"):
   (out/"wit/deps"/Path(name).name).write_bytes(archive.extractfile(name).read())
 (out/"wit/deps/cli.wit").write_bytes(archive.extractfile("wasi-wit-0.2.12/cli/cli.wit").read())
world=ROOT/"pkg/wasi/preview2/testdata/io-streams.wit"
(out/"wit/world.wit").write_bytes(world.read_bytes())
subprocess.run(["wasm-tools","component","embed",str(out/"wit"),"--world","io","--dummy-names","legacy","-t","-o",str(out/"io-dummy.wat")],check=True)
text=(out/'io-dummy.wat').read_text()
imports=re.findall(r'\(import "([^"]+)" "([^"]+)" \(func \(;(\d+);\)',text)
assert [(m,n) for m,n,i in imports]==[('wasi:io/error@0.2.12', '[method]error.to-debug-string'), ('wasi:io/error@0.2.12', '[resource-drop]error'), ('wasi:io/poll@0.2.12', '[method]pollable.ready'), ('wasi:io/poll@0.2.12', '[method]pollable.block'), ('wasi:io/poll@0.2.12', 'poll'), ('wasi:io/poll@0.2.12', '[resource-drop]pollable'), ('wasi:io/streams@0.2.12', '[method]input-stream.read'), ('wasi:io/streams@0.2.12', '[method]input-stream.blocking-read'), ('wasi:io/streams@0.2.12', '[method]input-stream.skip'), ('wasi:io/streams@0.2.12', '[method]input-stream.blocking-skip'), ('wasi:io/streams@0.2.12', '[method]input-stream.subscribe'), ('wasi:io/streams@0.2.12', '[method]output-stream.check-write'), ('wasi:io/streams@0.2.12', '[method]output-stream.write'), ('wasi:io/streams@0.2.12', '[method]output-stream.blocking-write-and-flush'), ('wasi:io/streams@0.2.12', '[method]output-stream.flush'), ('wasi:io/streams@0.2.12', '[method]output-stream.blocking-flush'), ('wasi:io/streams@0.2.12', '[method]output-stream.subscribe'), ('wasi:io/streams@0.2.12', '[method]output-stream.write-zeroes'), ('wasi:io/streams@0.2.12', '[method]output-stream.blocking-write-zeroes-and-flush'), ('wasi:io/streams@0.2.12', '[method]output-stream.splice'), ('wasi:io/streams@0.2.12', '[method]output-stream.blocking-splice'), ('wasi:io/streams@0.2.12', '[resource-drop]input-stream'), ('wasi:io/streams@0.2.12', '[resource-drop]output-stream'), ('wasi:cli/stdin@0.2.12', 'get-stdin'), ('wasi:cli/stdout@0.2.12', 'get-stdout'), ('wasi:cli/stderr@0.2.12', 'get-stderr'), ('wasi:cli/terminal-input@0.2.12', '[resource-drop]terminal-input'), ('wasi:cli/terminal-output@0.2.12', '[resource-drop]terminal-output'), ('wasi:cli/terminal-stdin@0.2.12', 'get-terminal-stdin'), ('wasi:cli/terminal-stdout@0.2.12', 'get-terminal-stdout'), ('wasi:cli/terminal-stderr@0.2.12', 'get-terminal-stderr')],"pinned Core import order changed"
exports={name:int(index) for name,index in re.findall(r'\(export "([^"]+)" \(func (\d+)\)\)',text)}
text=text.replace('(memory (;0;) 0)','(memory (;0;) 1)\n (global $heap (mut i32) (i32.const 8192))\n (data (i32.const 128) "XYZ!")\n (data (i32.const 1024) "'+('0123456789abcdef'*256)+'")')
exercise='''(func (;26;) (type 7) (result i32)
 (local $in i32) (local $out i32) (local $err i32) (local $poll i32)
 call 23 local.set $in call 24 local.set $out call 25 local.set $err
 local.get $in call 10 local.tee $poll call 3
 local.get $poll call 2 i32.eqz if unreachable end local.get $poll call 5
 local.get $in i64.const 0 i32.const 0 call 6 i32.const 0 call $ok
 i32.const 8 i32.load if unreachable end
 local.get $in i64.const 1 i32.const 0 call 8 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 1 i64.ne if unreachable end
 local.get $in i64.const 2 i32.const 0 call 7 i32.const 0 call $ok
 i32.const 8 i32.load i32.const 2 i32.ne if unreachable end
 i32.const 4 i32.load i32.load16_u i32.const 25442 i32.ne if unreachable end
 local.get $in i64.const 1 i32.const 0 call 9 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 1 i64.ne if unreachable end
 local.get $out i32.const 0 call 11 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 3 i64.lt_u if unreachable end
 local.get $out i32.const 128 i32.const 3 i32.const 0 call 12 i32.const 0 call $ok
 local.get $out i32.const 0 call 14 i32.const 0 call $ok
 local.get $out i32.const 0 call 15 i32.const 0 call $ok
 local.get $out call 16 local.tee $poll call 3
 local.get $poll call 2 i32.eqz if unreachable end local.get $poll call 5
 local.get $out i32.const 0 call 11 i32.const 0 call $ok
 local.get $out i64.const 2 i32.const 0 call 17 i32.const 0 call $ok
 local.get $out i64.const 3 i32.const 0 call 18 i32.const 0 call $ok
 local.get $out i32.const 131 i32.const 1 i32.const 0 call 13 i32.const 0 call $ok
 local.get $out local.get $in i64.const 2 i32.const 0 call 19 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 2 i64.ne if unreachable end
 local.get $out local.get $in i64.const 1 i32.const 0 call 20 i32.const 0 call $ok
 i32.const 8 i64.load i64.const 1 i64.ne if unreachable end
 local.get $out i32.const 0 call 15 i32.const 0 call $ok
 local.get $in i64.const 10 i32.const 0 call 6 i32.const 0 call $ok
 i32.const 8 i32.load i32.const 1 i32.ne if unreachable end
 i32.const 4 i32.load i32.load8_u i32.const 104 i32.ne if unreachable end
 local.get $in i64.const 1 i32.const 0 call 7
 i32.const 0 i32.load8_u i32.const 1 i32.ne if unreachable end
 i32.const 4 i32.load8_u i32.const 1 i32.ne if unreachable end
 local.get $in call 21 local.get $out call 22 local.get $err call 22
 i32.const 42
 )'''
exercise=exercise.replace('(;26;)',f'(;{exports["exercise"]};)')
text=re.sub(r'\(func \(;'+str(exports['exercise'])+r';\).*?\n  \)',exercise,text,count=1,flags=re.S)
post=exports['cabi_post_exercise']
text=text.replace(f'(func (;{post};) (type 1) (param i32))',f'(func (;{post};) (type 1) (param i32) i32.const 8192 global.set $heap)')
write='''(func (;28;) (type 2) (param $count i32) (result i32)
 (local $out i32) (local $at i32)
 call 24 local.set $out
 block $end loop $next
 local.get $at local.get $count i32.ge_u br_if $end
 local.get $out i32.const 1024 i32.const 4096 i32.const 0 call 13
 i32.const 0 call $ok
 local.get $at i32.const 1 i32.add local.set $at br $next
 end end
 local.get $out call 22 local.get $at
 )'''
write=write.replace('(;28;)',f'(;{exports["write-batch"]};)')
text=re.sub(r'\(func \(;'+str(exports['write-batch'])+r';\).*?\n  \)',write,text,count=1,flags=re.S)
post=exports['cabi_post_write-batch']
text=text.replace(f'(func (;{post};) (type 1) (param i32))',f'(func (;{post};) (type 1) (param i32) i32.const 8192 global.set $heap)')
alloc='''(func (;30;) (type 8) (param i32 i32 i32 i32) (result i32)
 (local $ptr i32)
 global.get $heap local.get 2 i32.const 1 i32.sub i32.add
 i32.const 0 local.get 2 i32.sub i32.and local.tee $ptr
 local.get 3 i32.add global.set $heap local.get $ptr
 )'''
alloc=alloc.replace('(;30;)',f'(;{exports["cabi_realloc"]};)')
text=re.sub(r'\(func \(;'+str(exports['cabi_realloc'])+r';\).*?\n  \)',alloc,text,count=1,flags=re.S)
failure='(func (;ERROR;) (type 7) (result i32)\n (local $out i32) (local $error i32)\n call 24 local.set $out\n local.get $out i32.const 131 i32.const 1 i32.const 0 call 13\n i32.const 0 i32.load8_u i32.const 1 i32.ne if unreachable end\n i32.const 4 i32.load8_u if unreachable end\n i32.const 8 i32.load local.set $error\n local.get $error i32.const 16 call 0\n i32.const 20 i32.load i32.eqz if unreachable end\n local.get $error call 1\n local.get $out i32.const 0 call 11\n i32.const 0 i32.load8_u i32.const 1 i32.ne if unreachable end\n i32.const 4 i32.load8_u i32.const 1 i32.ne if unreachable end\n local.get $out call 22 i32.const 1\n )'
failure=failure.replace('i32.const 4 i32.load8_u i32.const 1 i32.ne if unreachable end','i32.const 8 i32.load8_u i32.const 1 i32.ne if unreachable end')
failure=failure.replace('(;ERROR;)' ,f'(;{exports["error-probe"]};)')
text=re.sub(r'\(func \(;'+str(exports['error-probe'])+r';\).*?\n  \)',failure,text,count=1,flags=re.S)
post=exports['cabi_post_error-probe']
text=text.replace(f'(func (;{post};) (type 1) (param i32))',f'(func (;{post};) (type 1) (param i32) i32.const 8192 global.set $heap)')
extras={
 'terminal-probe':'i32.const 64 call 28 i32.const 64 i32.load8_u if i32.const 68 i32.load call 26 local.get $out i32.const 1 i32.add local.set $out end i32.const 64 call 29 i32.const 64 i32.load8_u if i32.const 68 i32.load call 27 local.get $out i32.const 1 i32.add local.set $out end i32.const 64 call 30 i32.const 64 i32.load8_u if i32.const 68 i32.load call 27 local.get $out i32.const 1 i32.add local.set $out end',
 'write-unchecked':'call 24 local.set $out local.get $out i32.const 128 i32.const 3 i32.const 0 call 12',
 'write-overpermit':'call 24 local.set $out local.get $out i32.const 0 call 11 i32.const 0 call $ok local.get $out i64.const -1 i32.const 0 call 17',
 'blocking-too-large':'call 24 local.set $out local.get $out i64.const 4097 i32.const 0 call 18',
 'parent-drop':'call 24 local.set $out local.get $out call 16 drop local.get $out call 22',
}
for name,body in extras.items():
 idx=exports[name]
 result='local.get $out' if name=='terminal-probe' else 'i32.const 0'
 definition=f'(func (;{idx};) (type 7) (result i32) (local $out i32) {body} {result})'
 text=re.sub(r'\(func \(;'+str(idx)+r';\).*?\n  \)',definition,text,count=1,flags=re.S)
text=text.replace('  (@custom "component-type"','  (func $ok (param i32) local.get 0 i32.load8_u if unreachable end)\n  (@custom "component-type"')
core=output/'io-streams-core.wat';core.write_text(text)
subprocess.run(['wasm-tools','parse',str(core),'-o',str(out/'io-streams-core.wasm')],check=True)
subprocess.run(['wasm-tools','component','new',str(out/'io-streams-core.wasm'),'-o',str(output/'io-streams.wasm')],check=True)
subprocess.run(['wasm-tools','validate','--features','all',str(output/'io-streams.wasm')],check=True)
subprocess.run(['wasm-tools','print',str(output/'io-streams.wasm'),'-o',str(output/'io-streams.wat')],check=True)

# Contract fixtures preserve independent component validity while changing one
# provider signature. Their bodies stay unreachable because binding must fail.
cases={
 "read-width":("read: func(len: u64)","read: func(len: u32)"),
 "write-list":("write: func(contents: list<u8>)","write: func(contents: list<u16>)"),
 "subscribe-result":("subscribe: func() -> pollable","subscribe: func() -> u32"),
 "error-case":("last-operation-failed(error)","last-operation-failed(u32)"),
 "error-label":("last-operation-failed(error)","failure(error)"),
 "skip-result":("skip: func(len: u64) -> result<u64,","skip: func(len: u64) -> result<u32,"),
 "extra-argument":("check-write: func()","check-write: func(unexpected: u32)"),
 "debug-result":("to-debug-string: func() -> string","to-debug-string: func() -> u32"),
}
contracts=out/"contract-wit"
if contracts.exists():shutil.rmtree(contracts)
shutil.copytree(out/"wit",contracts)
original=(contracts/"deps/io.wit").read_text()
for name,(before,after) in cases.items():
 assert before in original
 (contracts/"deps/io.wit").write_text(original.replace(before,after))
 binary=output/f"io-contract-{name}.wasm"
 core=out/f"io-contract-{name}-core.wasm"
 subprocess.run(["wasm-tools","component","embed",str(contracts),"--world","io","--dummy-names","legacy","-o",str(core)],check=True)
 subprocess.run(["wasm-tools","component","new",str(core),"-o",str(binary)],check=True)
 subprocess.run(["wasm-tools","validate","--features","all",str(binary)],check=True)
(contracts/"deps/io.wit").write_text(original)
subprocess.run(["wasm-tools","parse",str(ROOT/"pkg/wasi/preview2/testdata/io-errors.wat"),"-o",str(output/"io-errors.wasm")],check=True)
subprocess.run(["wasm-tools","validate","--features","all",str(output/"io-errors.wasm")],check=True)
files=[output/"io-streams.wasm",output/"io-errors.wasm"]+[output/f"io-contract-{name}.wasm" for name in cases]
manifest={"producer":"wasm-tools 1.258.0","wasiVersion":"0.2.12","wasiArchiveSha256":ARCHIVE_SHA256,"worldSha256":hashlib.sha256(world.read_bytes()).hexdigest(),"binarySha256":{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in files},"contractChanges":cases}
(output/"io-manifest.json").write_text(json.dumps(manifest,indent=2)+"\n")
print(f"Reproduced and validated {len(files)} IO components")
