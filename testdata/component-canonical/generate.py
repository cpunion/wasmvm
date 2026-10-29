import importlib.util,sys,random,json,hashlib
from pathlib import Path
from types import SimpleNamespace
root=Path(__file__).resolve().parents[2];src=Path(__file__).resolve().parent;p=src/'definitions.py'
spec=importlib.util.spec_from_file_location('canonical_reference',p);m=importlib.util.module_from_spec(spec);sys.modules[spec.name]=m;spec.loader.exec_module(m)
primitive={'bool':127,'s8':126,'u8':125,'s16':124,'u16':123,'s32':122,'u32':121,'s64':120,'u64':119,'f32':118,'f64':117,'char':116,'string':115}
classes={name:getattr(m,name.upper()+'Type' if name[0] in 'suf' and name!='string' else name.title()+'Type') for name in primitive}
rng=random.Random(20261030)
def generate(depth=0):
 if depth>=4 or rng.randrange(4)==0:
  name=rng.choice(list(primitive));return {'Kind':primitive[name]},classes[name]()
 kind=rng.choice(['list','tuple','record','variant','flags','enum','option','result'])
 if kind=='list':j,t=generate(depth+1);return {'Kind':112,'Element':j},m.ListType(t)
 if kind in ['tuple','record']:
  children=[generate(depth+1) for _ in range(rng.randrange(1,6))]
  fields=[{'Name':'f'+str(i),'Type':j} for i,(j,t) in enumerate(children)]
  return {'Kind':111 if kind=='tuple' else 114,'Fields':fields},m.TupleType([t for j,t in children]) if kind=='tuple' else m.RecordType([m.FieldType('f'+str(i),t) for i,(j,t) in enumerate(children)])
 if kind in ['flags','enum']:
  n=rng.choice([1,8,9,16,17,32]) if kind=='flags' else rng.choice([1,2,255,256,257])
  labels=['v'+str(i) for i in range(n)];return {'Kind':110 if kind=='flags' else 109,'Labels':labels},(m.FlagsType if kind=='flags' else m.EnumType)(labels)
 if kind=='option':j,t=generate(depth+1);return {'Kind':107,'Element':j},m.OptionType(t)
 if kind in ['result','variant']:
  children=[generate(depth+1) if rng.randrange(3)!=0 else (None,None) for _ in range(2 if kind=='result' else rng.randrange(1,6))]
  fields=[{'Name':'c'+str(i),'Type':j} for i,(j,t) in enumerate(children)]
  return {'Kind':106 if kind=='result' else 113,'Cases':fields},m.ResultType(children[0][1],children[1][1]) if kind=='result' else m.VariantType([m.CaseType('c'+str(i),t) for i,(j,t) in enumerate(children)])
class Memory:
 def __init__(self,width):self.width=width
 def ptr_type(self):return self.width
rows=[]
for i in range(512):
 j,t=generate()
 for wide,width in [(False,'i32'),(True,'i64')]:
  opts=SimpleNamespace(memory=Memory(width),async_=False,callback=None)
  rows.append({'Type':j,'Memory64':wide,'Size':m.elem_size(t,width),'Align':m.alignment(t,width),'Flat':m.flatten_type(t,opts)})
out=root/'testdata/component-canonical';(out/'layout-oracle.json').write_text(json.dumps(rows,separators=(',',':'))+'\n')
(out/'LICENSE-APACHE').write_bytes((src/'LICENSE-APACHE').read_bytes())
(out/'definitions.py').write_bytes(p.read_bytes())
(out/'generate.py').write_bytes(Path(__file__).read_bytes())
files=['layout-oracle.json','definitions.py','generate.py','LICENSE-APACHE']
manifest={'revision':'d095e2f60ddae1f9a1e9b1e2fb547ec1114da677','sourcePath':'design/mvp/canonical-abi/definitions.py','seed':20261030,'uniqueTypes':512,'cases':len(rows),'sha256':{n:hashlib.sha256((out/n).read_bytes()).hexdigest() for n in files}}
(out/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n');print(len(rows),'independent reference layouts', (out/'layout-oracle.json').stat().st_size,'bytes')
