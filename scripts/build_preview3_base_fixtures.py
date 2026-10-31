#!/usr/bin/env python3
"""Build version-specific WASI 0.3.2 base/clock execution and rejection fixtures."""
import hashlib
import json
from pathlib import Path
import subprocess
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'pkg/wasi/preview3/testdata'
OUT.mkdir(parents=True,exist_ok=True)
assert subprocess.check_output(['wasm-tools','--version'],text=True).strip()=='wasm-tools 1.258.0'
WIT=ROOT/'artifacts/components/wasi-inputs/preview3/wasi-wit-0.3.2.tar.gz'
SHA='2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311'
assert hashlib.sha256(WIT.read_bytes()).hexdigest()==SHA
fixtures={}
inputs={}
for name in ['environment','random','exit','wrong-arguments','random-wrong-seed','exit-wrong-code']:
 p=ROOT/'pkg/wasi/preview2/testdata'/f'{name}.wat'
 inputs[str(p.relative_to(ROOT))]=hashlib.sha256(p.read_bytes()).hexdigest()
 s=p.read_text().replace('@0.2.12','@0.3.2').replace('"initial-cwd"','"get-initial-cwd"').replace('"insecure-seed"','"get-insecure-seed"')
 fixtures[name]=s
fixtures['system-clock']='''(component
 (type $instant (record (field "seconds" s64) (field "nanoseconds" u32))) (export $instant-public "instant-type" (type $instant))
 (import "wasi:clocks/system-clock@0.3.2" (instance $clock
  (type $i (record (field "seconds" s64) (field "nanoseconds" u32))) (export "instant" (type $public (eq $i)))
  (export "now" (func (result $public))) (export "get-resolution" (func (result u64)))))
 (alias export $clock "now" (func $now)) (alias export $clock "get-resolution" (func $resolution))
 (core module $memory (memory (export "memory") 1)) (core instance $memory (instantiate $memory))
 (alias core export $memory "memory" (core memory $mem))
 (core func $now (canon lower (func $now) (memory $mem))) (core func $resolution (canon lower (func $resolution)))
 (core instance $env (export "now" (func $now)) (export "resolution" (func $resolution)))
 (core module $m (import "env" "now" (func $now (param i32))) (import "env" "resolution" (func $resolution (result i64)))
  (func (export "instant") (result i32) i32.const 32 call $now i32.const 32)
  (func (export "resolution") (result i64) call $resolution))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (func (export "instant") (result $instant-public) (canon lift (core func $m "instant") (memory $mem)))
 (func (export "resolution") (result u64) (canon lift (core func $m "resolution"))))'''
fixtures['monotonic-clock']='''(component
 (import "wasi:clocks/monotonic-clock@0.3.2" (instance $clock
  (export "now" (func (result u64))) (export "get-resolution" (func (result u64)))
  (export "wait-for" (func async (param "how-long" u64))) (export "wait-until" (func async (param "when" u64)))))
 (alias export $clock "now" (func $now)) (alias export $clock "get-resolution" (func $resolution))
 (alias export $clock "wait-for" (func $wait-for)) (alias export $clock "wait-until" (func $wait-until))
 (core func $now (canon lower (func $now))) (core func $resolution (canon lower (func $resolution)))
 (core func $wait-for (canon lower (func $wait-for))) (core func $wait-until (canon lower (func $wait-until)))
 (core instance $env (export "now" (func $now)) (export "resolution" (func $resolution)) (export "wait-for" (func $wait-for)) (export "wait-until" (func $wait-until)))
 (core module $m
  (import "env" "now" (func $now (result i64))) (import "env" "resolution" (func $resolution (result i64)))
  (import "env" "wait-for" (func $wait-for (param i64))) (import "env" "wait-until" (func $wait-until (param i64)))
  (func (export "now") (result i64) call $now) (func (export "resolution") (result i64) call $resolution)
  (func (export "wait-for") (param i64) (result i64) local.get 0 call $wait-for call $now)
  (func (export "wait-until") (param i64) (result i64) local.get 0 call $wait-until call $now))
 (core instance $m (instantiate $m (with "env" (instance $env))))
 (func (export "now") (result u64) (canon lift (core func $m "now")))
 (func (export "resolution") (result u64) (canon lift (core func $m "resolution")))
 (func (export "wait-for") async (param "duration" u64) (result u64) (canon lift (core func $m "wait-for")))
 (func (export "wait-until") async (param "mark" u64) (result u64) (canon lift (core func $m "wait-until"))))'''
fixtures['system-clock-wrong-sign']=fixtures['system-clock'].replace('s64','u64')
fixtures['system-clock-wrong-resolution']=fixtures['system-clock'].replace('(export "get-resolution" (func (result u64)))','(export "get-resolution" (func (result s64)))')
fixtures['monotonic-clock-wrong-async']=fixtures['monotonic-clock'].replace('(func async (param "when" u64))','(func (param "when" u64))')
fixtures['monotonic-clock-wrong-now']='(component (import "wasi:clocks/monotonic-clock@0.3.2" (instance (export "now" (func (param "unexpected" u32) (result u64))))))'
fixtures['monotonic-clock-wrong-wait']='(component (import "wasi:clocks/monotonic-clock@0.3.2" (instance (export "wait-for" (func async (param "duration" u32))))))'
for mode in ['success','failure','trap','sync','parameter','payload','exit']:
 kind='async';params='';coreparams='';result='(result)';body='i32.const 0';imports='';coreimports='';env=''
 if mode=='failure':body='i32.const 1'
 if mode=='trap':body='unreachable'
 if mode=='sync':kind=''
 if mode=='parameter':params='(param "unexpected" u32)';coreparams='(param i32)'
 if mode=='payload':result='(result (error u8))';body='i32.const 32';coreimports='(memory (export "memory") 1)';env='(alias core export $m "memory" (core memory $mem))'
 if mode=='exit':
  imports='(import "wasi:cli/exit@0.3.2" (instance $exit (export "exit-with-code" (func (param "status-code" u8))))) (alias export $exit "exit-with-code" (func $exit)) (core func $exit (canon lower (func $exit))) (core instance $env (export "exit" (func $exit)))'
  coreimports='(import "env" "exit" (func $exit (param i32)))';body='i32.const 255 call $exit unreachable'
 binding='(with "env" (instance $env))' if mode=='exit' else ''
 memory='(memory $mem)' if mode=='payload' else ''
 fixtures['command-'+mode]=f'(component {imports} (core module $m {coreimports} (func (export "run") {coreparams} (result i32) {body})) (core instance $m (instantiate $m {binding})) {env} (func $run {kind} {params} (result {result}) (canon lift (core func $m "run") {memory})) (instance $run (export "run" (func $run))) (export "wasi:cli/run@0.3.2" (instance $run)))'
pins={}
for name,s in fixtures.items():
 p=OUT/(name+'.wat');p.write_text(s.rstrip()+'\n')
 subprocess.run(['wasm-tools','parse',str(p),'-o',str(p.with_suffix('.wasm'))],check=True)
 subprocess.run(['wasm-tools','validate','--features','all',str(p.with_suffix('.wasm'))],check=True)
 pins[name]={ext:hashlib.sha256((OUT/(name+'.'+ext)).read_bytes()).hexdigest() for ext in ['wat','wasm']}
(OUT/'base.json').write_text(json.dumps({'wasiVersion':'0.3.2','witArchiveSHA256':SHA,'producer':'wasm-tools 1.258.0','generatorSHA256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),'inputs':inputs,'sha256':pins},indent=2)+'\n')
print('Built',len(pins),'validated WASI 0.3.2 fixtures')
