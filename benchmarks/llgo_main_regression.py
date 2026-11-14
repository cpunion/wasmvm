"""Verify frozen source regressions with a separately pinned current LLGo."""
import hashlib,json,os,shutil,subprocess,sys,tarfile
from pathlib import Path
ROOT=Path.cwd();OUT=ROOT/'main-regression';OUT.mkdir()
sys.path.insert(0,str(ROOT/'control/benchmarks'))
import own_build
sys.path.insert(0,str(ROOT/'head/benchmarks/native'))
from memory_fixtures import checksum
REV='c0d07ce2b58c8f2f62b25133234486d239021ee1'
SYMBOL='github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop'
run=lambda args,**kw:subprocess.check_output(args,text=True,**kw)
cli=shutil.which('llgo');assert REV[:12]in run([cli,'version'])
nm=shutil.which('llvm-nm-22')or shutil.which('llvm-nm')
objdump=shutil.which('llvm-objdump-22')or shutil.which('llvm-objdump')
assert nm and objdump
result={'compiler':REV,'cliVersion':run([cli,'version']),'cliSha256':own_build.digest(Path(cli)),'cliGoBuildInfo':run(['go','version','-m',cli]),'clang':run(['clang','--version']),'profile':['-O3','-lto=full'],'roles':{},'scope':'Separate compiler baseline event. Identical frozen engine sources; only common harness compiler provenance changes. Static code generation and checked guest execution, not timing acceptance. Production pin remains56fa9ae.'}
for role in ['counter','dense','scalar']:
 folder=ROOT/('source-'+role);folder.mkdir()
 with tarfile.open(ROOT/'downloads'/role/'poll-llgo.tar.gz')as t:t.extractall(folder,filter='data')
 base=folder/'routing-build';manifest=json.loads((base/'poll-llgo.json').read_text());record=manifest['builds']['poll-llgo'];source=base/'poll-llgo'
 assert own_build.hashes(source,own_build.ENGINE_PATHS,own_build.HARNESS)==record['sourceHashes']
 gen=source/'internal/bench/phase/provenance_gen.go'
 gen.write_text('package phase\nfunc init(){sourceRevision='+json.dumps(manifest['candidateRevision'])+';compilerRevision='+json.dumps(REV)+'}\n')
 binary=OUT/(role+'-runner');build_cmd=[cli,'build','-O3','-lto=full','-o',str(binary),'./benchmarks/embedded/wasmvm']
 subprocess.run(build_cmd,cwd=source,check=True,timeout=180)
 assert own_build.hashes(source,own_build.ENGINE_PATHS,own_build.HARNESS)==record['sourceHashes']
 asm=run([objdump,'--disassemble-symbols='+SYMBOL,'--no-show-raw-insn','--x86-asm-syntax=intel',str(binary)])
 (OUT/(role+'.asm')).write_text(asm)
 symbols=[s for s in run([nm,'--print-size',str(binary)]).splitlines()if s.endswith(' '+SYMBOL)]
 expected=checksum('dependent-address',0x12345678,100000)
 cmd=[str(binary),'--mode','api','--backend','interpreter','--wasm',str(ROOT/'head/benchmarks/native/memory-fixtures/dependent-address.wasm'),'--iterations','100000','--seed',str(0x12345678),'--expected',str(expected),'--samples','1','--warmups','0']
 proc=subprocess.run(cmd,input='run\n',text=True,capture_output=True,timeout=30)
 (OUT/(role+'.stdout')).write_text(proc.stdout);(OUT/(role+'.stderr')).write_text(proc.stderr)
 assert proc.returncode==0,proc.stderr
 receipt=json.loads(proc.stdout.splitlines()[-1]);assert receipt['sourceRevision']==manifest['candidateRevision']and receipt['compilerRevision']==REV and receipt['warmResults']==[expected]and receipt['result']==expected
 result['roles'][role]={'source':manifest['candidateRevision'],'binarySha256':own_build.digest(binary),'symbol':symbols,'buildCommand':build_cmd,'executionCommand':cmd,'checkedGuest':receipt,'prologue':asm.splitlines()[5:20]}
 (OUT/'results.json').write_text(json.dumps(result,indent=2)+'\n');print(role,symbols,flush=True)

for path in OUT.glob("*.llgo-debug-lto"):
 if path.is_dir():shutil.rmtree(path)
