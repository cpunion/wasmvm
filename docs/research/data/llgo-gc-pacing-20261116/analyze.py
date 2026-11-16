import pathlib,gzip,json,math,random,statistics,sys
for root in sys.argv[1:]:
 root=pathlib.Path(root)
 for payload in (16,65536):
  folder=root/f'p3-{payload}'
  with gzip.open(folder/'rows.jsonl.gz','rt') as f: rows=[json.loads(line)for line in f]
  for r in rows:
   batches=r['record']['service']['Batches'][3:23]
   r['gc']=sum(b['GCAfter']-b['GCBefore']for b in batches)
   r['heapMiB']=max(b['HeapAfter']for b in batches)/1048576
   r['physicalMiB']=max(q['after']['ri_phys_footprint']for q in r['rounds'])/1048576
  labels={r['id']for r in rows};by={(r['block'],r['id']):r for r in rows}
  result={'median':{},'paired':{}}
  for label in sorted(labels):result['median'][label]={k:statistics.median(r[k]for r in rows if r['id']==label)for k in ('cpuUs','instructions','throughput','gc','heapMiB','physicalMiB','allocsPerRequest','bytesPerRequest')}
  for left,right in [('head','base'),('base-aa','base'),('head-aa','head')]:
   ratios={}
   for k in ('cpuUs','instructions','throughput'):
    v=[math.log(by[b,left][k]/by[b,right][k])for b in range(20)]
    rng=random.Random(20261116)
    boot=sorted(math.exp(statistics.mean(rng.choices(v,k=20)))for _ in range(10000))
    ratios[k]=[math.exp(statistics.mean(v)),boot[250],boot[9749]]
   result['paired'][f'{left}/{right}']=ratios
  (folder/'analysis.json').write_text(json.dumps(result,indent=2)+'\n')
  print(root.name,payload,json.dumps(result),flush=True)
