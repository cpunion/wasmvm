#!/usr/bin/env python3
"""Build direction-aware cm64 borrow fixtures for the October 30 ABI correction."""
import hashlib
import json
from pathlib import Path
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT/'pkg/component/testdata'
SPEC = 'fed10b0b823b82553dd8363f16a640816f5eac65'
assert subprocess.check_output(['wasm-tools', '--version'], text=True).strip() == 'wasm-tools 1.258.0'


def provider(wide):
    a = 'i64' if wide else 'i32'
    memory = 'i64 1' if wide else '1'
    fields = ' '.join(['(borrow $R)'] * 17)
    return f'''(component
 (core module $memory
  (memory (export "memory") {memory})
  (global $bump (mut {a}) ({a}.const 256))
  (func (export "realloc") (param {a} {a}) (param $align {a}) (param $size {a}) (result {a})
   (local $ptr {a})
   global.get $bump local.get $align {a}.const 1 {a}.sub {a}.add
   local.get $align {a}.const 1 {a}.sub {a}.const -1 {a}.xor {a}.and local.tee $ptr
   local.get $size {a}.add global.set $bump local.get $ptr))
 (core instance $memory (instantiate $memory))
 (core module $state (func (export "drop") (param i64)))
 (core instance $state (instantiate $state))
 (type $r (resource (rep i64) (dtor (core func $state "drop"))))
 (export $R "r" (type $r))
 (type $record (record (field "prefix" u8) (field "resource" (borrow $R)) (field "suffix" u32)))
 (export $Record "record-type" (type $record))
 (type $variant (variant (case "none") (case "resource" (borrow $R)) (case "number" u32)))
 (export $Variant "variant-type" (type $variant))
 (type $spill (tuple {fields}))
 (core func $new (canon resource.new $r))
 (core func $ret (canon task.return (result u64)))
 (core func $cancel (canon task.cancel))
 (core func $get (canon context.get i64 0))
 (core func $set (canon context.set i64 0))
 (core func $get1 (canon context.get i64 1))
 (core func $set1 (canon context.set i64 1))
 (core func $ws-new (canon waitable-set.new))
 (core func $ws-drop (canon waitable-set.drop))
 (core module $code
  (import "" "memory" (memory {memory}))
  (import "" "new" (func $new (param i64) (result i32)))
  (import "" "ret" (func $ret (param i64)))
  (import "" "cancel" (func $cancel))
  (import "" "get" (func $get (result i64)))
  (import "" "set" (func $set (param i64)))
  (import "" "get1" (func $get1 (result i64)))
  (import "" "set1" (func $set1 (param i64)))
  (import "" "ws-new" (func $ws-new (result i32)))
  (import "" "ws-drop" (func $ws-drop (param i32)))
  (func (export "make") (param i64) (result i32) local.get 0 call $new)
  (func (export "direct") (param i64) (result i64) local.get 0)
  (func (export "trap") (param i64) unreachable)
  (func (export "record") (param i32 i64 i32) (result i64)
   local.get 1 local.get 0 i64.extend_i32_u i64.add local.get 2 i64.extend_i32_u i64.add)
  (func (export "variant") (param i32 i64) (result i64)
   local.get 0 if (result i64) local.get 1 else i64.const 0 end)
  (func (export "list") (param $ptr {a}) (param $n {a}) (result i64)
   (local $i {a}) (local $sum i64)
   block $done loop $next
    local.get $i local.get $n {a}.ge_u br_if $done
    local.get $sum local.get $ptr local.get $i {a}.const 8 {a}.mul {a}.add i64.load i64.xor local.set $sum
    local.get $i {a}.const 1 {a}.add local.set $i br $next
   end end local.get $sum)
  (func (export "spill") (param {a}) (result i64)
   local.get 0 i64.load local.get 0 i64.load offset=128 i64.xor)
  (func (export "async") (param i64) (result i32) local.get 0 call $set i32.const 1)
  (func (export "callback") (param i32 i32 i32) (result i32)
   call $get call $ret i32.const 0)
  (func (export "wait") (param i64) (result i32)
   local.get 0 call $set call $ws-new i64.extend_i32_u call $set1
   call $get1 i32.wrap_i64 i32.const 4 i32.shl i32.const 2 i32.or)
  (func (export "wait-callback") (param i32 i32 i32) (result i32)
   local.get 0 i32.const 6 i32.ne if unreachable end
   call $get1 i32.wrap_i64 call $ws-drop call $cancel i32.const 0))
 (core instance $code (instantiate $code (with "" (instance
  (export "memory" (memory $memory "memory")) (export "new" (func $new))
  (export "ret" (func $ret)) (export "cancel" (func $cancel))
  (export "get" (func $get)) (export "set" (func $set))
  (export "get1" (func $get1)) (export "set1" (func $set1))
  (export "ws-new" (func $ws-new)) (export "ws-drop" (func $ws-drop))))))
 (func (export "make") (param "rep" u64) (result (own $R)) (canon lift (core func $code "make")))
 (func (export "direct") (param "value" (borrow $R)) (result u64) (canon lift (core func $code "direct")))
 (func (export "trap") (param "value" (borrow $R)) (canon lift (core func $code "trap")))
 (func (export "record") (param "value" $Record) (result u64) (canon lift (core func $code "record")))
 (func (export "variant") (param "value" $Variant) (result u64) (canon lift (core func $code "variant")))
 (func (export "list") (param "value" (list (borrow $R))) (result u64)
  (canon lift (core func $code "list") (memory (core memory $memory "memory")) (realloc (core func $memory "realloc"))))
 (func (export "spill") (param "value" $spill) (result u64)
  (canon lift (core func $code "spill") (memory (core memory $memory "memory")) (realloc (core func $memory "realloc"))))
 (func (export "async") async (param "value" (borrow $R)) (result u64)
  (canon lift (core func $code "async") async (callback (core func $code "callback"))))
 (func (export "wait") async (param "value" (borrow $R)) (result u64)
  (canon lift (core func $code "wait") async (callback (core func $code "wait-callback")))))'''


def composed(wide):
    a = 'i64' if wide else 'i32'
    memory = 'i64 1' if wide else '1'
    return '(component\n'+provider(wide).replace('(component', '(component $Provider', 1)+f'''
 (instance $provider (instantiate $Provider))
 (component $Client
  (import "provider" (instance $provider
   (export "r" (type $R (sub resource)))
   (export "make" (func (param "rep" u64) (result (own $R))))
   (export "direct" (func (param "value" (borrow $R)) (result u64)))
   (export "list" (func (param "value" (list (borrow $R))) (result u64)))
   (export "wait" (func async (param "value" (borrow $R)) (result u64)))))
  (alias export $provider "r" (type $R))
  (core module $memory (memory (export "memory") {memory}))
  (core instance $memory (instantiate $memory))
  (core func $make (canon lower (func $provider "make")))
  (core func $direct (canon lower (func $provider "direct")))
  (core func $list (canon lower (func $provider "list") (memory (core memory $memory "memory"))))
  (core func $wait (canon lower (func $provider "wait") async (memory (core memory $memory "memory"))))
  (core func $drop (canon resource.drop $R))
  (core func $cancel (canon subtask.cancel async))
  (core func $sub-drop (canon subtask.drop))
  (core module $code
   (import "" "memory" (memory {memory}))
   (import "" "make" (func $make (param i64) (result i32)))
   (import "" "direct" (func $direct (param i32) (result i64)))
   (import "" "list" (func $list (param {a} {a}) (result i64)))
   (import "" "wait" (func $wait (param i32 {a}) (result i32)))
   (import "" "drop" (func $drop (param i32)))
   (import "" "cancel" (func $cancel (param i32) (result i32)))
   (import "" "sub-drop" (func $sub-drop (param i32)))
   (func (export "run") (param i64) (result i64)
    (local $a i32) (local $b i32) (local $out i64)
    local.get 0 call $make local.set $a
    local.get 0 i64.const 1 i64.add call $make local.set $b
    local.get $a call $direct local.set $out
    {a}.const 0 local.get $a i32.store
    {a}.const 4 local.get $b i32.store
    {a}.const 0 {a}.const 2 call $list local.get $out i64.xor local.set $out
    local.get $a call $drop local.get $b call $drop local.get $out)
   (func (export "cancel") (param i64) (result i32)
    (local $r i32) (local $sub i32)
    local.get 0 call $make local.tee $r {a}.const 16 call $wait
    i32.const 4 i32.shr_u local.tee $sub call $cancel
    i32.const 4 i32.ne if unreachable end
    local.get $sub call $sub-drop local.get $r call $drop i32.const 42))
  (core instance $code (instantiate $code (with "" (instance
   (export "memory" (memory $memory "memory")) (export "make" (func $make))
   (export "direct" (func $direct)) (export "list" (func $list)) (export "wait" (func $wait))
   (export "drop" (func $drop)) (export "cancel" (func $cancel)) (export "sub-drop" (func $sub-drop))))))
  (func (export "run") (param "rep" u64) (result u64) (canon lift (core func $code "run")))
  (func (export "cancel") (param "rep" u64) (result u32) (canon lift (core func $code "cancel"))))
 (instance $client (instantiate $Client (with "provider" (instance $provider))))
 (alias export $client "run" (func $run)) (export "run" (func $run))
 (alias export $client "cancel" (func $cancel)) (export "cancel" (func $cancel)))'''


fixtures = {}
for wide in (False, True):
    suffix = '64' if wide else '32'
    fixtures['cm64-borrow-local'+suffix] = provider(wide)
    fixtures['cm64-borrow-composed'+suffix] = composed(wide)
    for kind, pointer in [('align', 1), ('range', 1 << 32 if wide else 65528)]:
        text = provider(wide)
        start = text.index('  (func (export "realloc")')
        end = text.index(' (core instance $memory', start)
        address = 'i64' if wide else 'i32'
        replacement = f'  (func (export "realloc") (param {address} {address} {address} {address}) (result {address}) {address}.const {pointer}))\n'
        fixtures['cm64-borrow-bad-'+kind+suffix] = text[:start]+replacement+text[end:]
legacy = OUT/'ownership64-borrow.wat'
fixtures['cm64-borrow-direct'] = legacy.read_text().replace('(dtor (func $destructor))', '(dtor (core func $destructor))').replace('(func (export "borrow-rep") (param i32) (result i32) local.get 0)', '(func (export "borrow-rep") (param i64) (result i64) local.get 0)').replace('(param "self" (borrow 1)) (result u32)', '(param "self" (borrow 1)) (result u64)')
sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
spec_path = ROOT/'testdata/component-cm64/manifest.json'
spec = json.loads(spec_path.read_text())
assert spec['revision'] == SPEC
assert sha(spec_path.with_name('inputs.tar.gz')) == spec['archiveSHA256']
with tarfile.open(spec_path.with_name('inputs.tar.gz')) as archive:
    assert {m.name for m in archive.getmembers()} == set(spec['sha256'])
    for member in archive.getmembers():
        assert member.isfile() and hashlib.sha256(archive.extractfile(member).read()).hexdigest() == spec['sha256'][member.name]

manifest = {'componentModelCorrection': SPEC, 'producer': 'wasm-tools 1.258.0 parse',
            'generatorSHA256': sha(Path(__file__)), 'specManifestSHA256': sha(spec_path), 'sourceSHA256': {'ownership64-borrow.wat': sha(legacy)}, 'fixtures': {}}
for name, text in fixtures.items():
    wat = OUT/(name+'.wat')
    wat.write_text(text.rstrip()+'\n')
    subprocess.run(['wasm-tools', 'parse', wat, '-o', wat.with_suffix('.wasm')], check=True)
    manifest['fixtures'][name] = {ext: sha(OUT/(name+'.'+ext)) for ext in ('wat','wasm')}
(OUT/'cm64-borrows.json').write_text(json.dumps(manifest, indent=2)+'\n')
print('Built', len(fixtures), 'cm64 borrow fixtures; independent old-producer validation is a recorded boundary')
