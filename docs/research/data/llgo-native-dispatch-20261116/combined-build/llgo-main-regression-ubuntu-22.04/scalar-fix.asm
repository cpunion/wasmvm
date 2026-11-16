
/home/runner/work/wasmvm/wasmvm/main-regression/scalar-fix:	file format elf64-x86-64

Disassembly of section .text:

00000000003a5600 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a5600:      	push	rbp
  3a5601:      	mov	rbp, rsp
  3a5604:      	push	r15
  3a5606:      	push	r14
  3a5608:      	push	r13
  3a560a:      	push	r12
  3a560c:      	push	rbx
  3a560d:      	sub	rsp, 0x2a8
  3a5614:      	mov	rcx, qword ptr [rdi + 0x48]
  3a5618:      	nop	dword ptr [rax + rax]
  3a5620:      	test	rcx, rcx
  3a5623:      	jle	0x3a7415 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e15>
  3a5629:      	mov	rbx, rsi
  3a562c:      	mov	r11, rdi
  3a562f:      	movzx	eax, byte ptr [rdi + 0x20]
  3a5633:      	mov	byte ptr [rbp - 0x39], al
  3a5636:      	lea	rax, [rbp - 0x90]
  3a563d:      	mov	qword ptr [rbp - 0x68], rax
  3a5641:      	mov	qword ptr [rbp - 0x38], rdi
  3a5645:      	mov	rdi, qword ptr [r11 + 0x40]
  3a5649:      	shl	rcx, 0x6
  3a564d:      	mov	byte ptr [rdi + rcx - 0x30], 0x0
  3a5652:      	mov	rsi, qword ptr [rdi + rcx - 0x40]
  3a5657:      	mov	r10, qword ptr [rsi + 0x48]
  3a565b:      	mov	r13, qword ptr [r10 + 0x98]
  3a5662:      	mov	r9, qword ptr [rdi + rcx - 0x28]
  3a5667:      	mov	rdx, qword ptr [r11 + 0x88]
  3a566e:      	lea	rax, [r9 + r13]
  3a5672:      	cmp	rax, rdx
  3a5675:      	ja	0x3a732d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d2d>
  3a567b:      	nop	dword ptr [rax + rax]
  3a5680:      	cmp	r9, rax
  3a5683:      	ja	0x3a733f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3f>
  3a5689:      	mov	r15, qword ptr [r10 + 0xb8]
  3a5690:      	mov	r12d, dword ptr [rdi + rcx - 0x38]
  3a5695:      	sub	rdx, r9
  3a5698:      	shl	r9, 0x4
  3a569c:      	test	rdx, rdx
  3a569f:      	mov	eax, 0x0
  3a56a4:      	cmovle	r9, rax
  3a56a8:      	cmp	r15, r12
  3a56ab:      	jbe	0x3a72f1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cf1>
  3a56b1:      	add	rdi, rcx
  3a56b4:      	mov	qword ptr [rbp - 0x60], rdi
  3a56b8:      	mov	qword ptr [rbp - 0xa8], rsi
  3a56bf:      	mov	rax, qword ptr [rsi + 0x40]
  3a56c3:      	mov	qword ptr [rbp - 0x50], rax
  3a56c7:      	mov	rcx, qword ptr [r10 + 0xb0]
  3a56ce:      	add	r9, qword ptr [r11 + 0x78]
  3a56d2:      	mov	r14d, r12d
  3a56d5:      	mov	qword ptr [rbp - 0x78], r13
  3a56d9:      	mov	qword ptr [rbp - 0x30], r10
  3a56dd:      	mov	qword ptr [rbp - 0x48], r9
  3a56e1:      	mov	qword ptr [rbp - 0xb0], r15
  3a56e8:      	mov	qword ptr [rbp - 0x58], rcx
  3a56ec:      	mov	esi, 0xffffffff
  3a56f1:      	jmp	0x3a5729 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x129>
  3a56f3:      	nop	word ptr cs:[rax + rax]
  3a5700:      	inc	r14d
  3a5703:      	mov	r15d, r14d
  3a5706:      	mov	rcx, qword ptr [rbp - 0x58]
  3a570a:      	mov	r10, r8
  3a570d:      	mov	r12d, r15d
  3a5710:      	mov	r14d, r15d
  3a5713:      	mov	r15, qword ptr [rbp - 0xb0]
  3a571a:      	nop	word ptr [rax + rax]
  3a5720:      	cmp	r15, r12
  3a5723:      	jbe	0x3a72f1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cf1>
  3a5729:      	shl	r12, 0x5
  3a572d:      	add	r12, rcx
  3a5730:      	test	byte ptr [rbp - 0x39], 0x1
  3a5734:      	je	0x3a5796 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196>
  3a5736:      	mov	rsi, qword ptr [r10 + 0x80]
  3a573d:      	mov	edi, dword ptr [r12 + 0x4]
  3a5742:      	cmp	rsi, rdi
  3a5745:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a574b:      	mov	r15, qword ptr [r10 + 0x78]
  3a574f:      	imul	r13, rdi, 0xe0
  3a5756:      	mov	eax, dword ptr [r15 + r13]
  3a575a:      	cmp	eax, 0x5
  3a575d:      	setne	cl
  3a5760:      	cmp	eax, 0xb
  3a5763:      	setne	al
  3a5766:      	and	al, cl
  3a5768:      	movzx	ecx, al
  3a576b:      	mov	rax, qword ptr [r11 + 0x150]
  3a5772:      	sub	rax, rcx
  3a5775:      	jb	0x3a7174 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b74>
  3a577b:      	mov	qword ptr [r11 + 0x150], rax
  3a5782:      	mov	r13, qword ptr [rbp - 0x78]
  3a5786:      	mov	r15, qword ptr [rbp - 0xb0]
  3a578d:      	mov	rcx, qword ptr [rbp - 0x58]
  3a5791:      	mov	esi, 0xffffffff
  3a5796:      	inc	rbx
  3a5799:      	nop	dword ptr [rax]
  3a57a0:      	test	bl, 0x3f
  3a57a3:      	jne	0x3a5826 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x226>
  3a57a9:      	mov	rax, qword ptr [r11 + 0x160]
  3a57b0:      	test	rax, rax
  3a57b3:      	je	0x3a57c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c8>
  3a57b5:      	test	r11, r11
  3a57b8:      	je	0x3a58ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2ec>
  3a57be:      	mov	eax, dword ptr [rax]
  3a57c0:      	test	eax, eax
  3a57c2:      	je	0x3a5dd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7d2>
  3a57c8:      	mov	rax, qword ptr [r11 + 0x8]
  3a57cc:      	test	rax, rax
  3a57cf:      	je	0x3a7406 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e06>
  3a57d5:      	mov	rdi, qword ptr [r11 + 0x10]
  3a57d9:      	mov	rcx, qword ptr [rax + 0x8]
  3a57dd:      	test	byte ptr [rcx + 0x17], 0x20
  3a57e1:      	je	0x3a57fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa>
  3a57e3:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a57e7:      	and	cl, 0x1f
  3a57ea:      	cmp	cl, 0x16
  3a57ed:      	je	0x3a57fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa>
  3a57ef:      	mov	qword ptr [rbp - 0x90], rdi
  3a57f6:      	mov	rdi, qword ptr [rbp - 0x68]
  3a57fa:      	call	qword ptr [rax + 0x28]
  3a57fd:      	mov	esi, 0xffffffff
  3a5802:      	mov	r10, qword ptr [rbp - 0x30]
  3a5806:      	mov	r9, qword ptr [rbp - 0x48]
  3a580a:      	mov	rcx, qword ptr [rbp - 0x58]
  3a580e:      	test	rax, rax
  3a5811:      	je	0x3a5910 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x310>
  3a5817:      	cmp	qword ptr [rax + 0x8], 0x0
  3a581c:      	mov	r11, qword ptr [rbp - 0x38]
  3a5820:      	jne	0x3a71c6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc6>
  3a5826:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a582c:      	je	0x3a5940 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a5832:      	test	byte ptr [rbp - 0x39], 0x1
  3a5836:      	je	0x3a5880 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x280>
  3a5838:      	mov	edx, dword ptr [r12]
  3a583c:      	nop	dword ptr [rax]
  3a5840:      	cmp	edx, 0x20
  3a5843:      	je	0x3a5940 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a5849:      	cmp	edx, 0x41
  3a584c:      	je	0x3a5940 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a5852:      	jmp	0x3a5944 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x344>
  3a5857:      	nop	word ptr cs:[rax + rax]
  3a5866:      	nop	word ptr cs:[rax + rax]
  3a5875:      	nop	word ptr cs:[rax + rax]
  3a5880:      	mov	edx, dword ptr [r12]
  3a5884:      	cmp	edx, 0x20
  3a5887:      	je	0x3a5892 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x292>
  3a5889:      	cmp	edx, 0x41
  3a588c:      	jne	0x3a5944 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x344>
  3a5892:      	mov	eax, ebx
  3a5894:      	not	eax
  3a5896:      	test	al, 0x3f
  3a5898:      	je	0x3a5940 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a589e:      	nop
  3a58a0:      	cmp	edx, 0x41
  3a58a3:      	jne	0x3a58c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c9>
  3a58a5:      	mov	edi, dword ptr [r12 + 0x8]
  3a58aa:      	cmp	r13, rdi
  3a58ad:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a58b3:      	mov	eax, dword ptr [r12 + 0x18]
  3a58b8:      	shl	rdi, 0x4
  3a58bc:      	mov	qword ptr [r9 + rdi], rax
  3a58c0:      	mov	qword ptr [r9 + rdi + 0x8], 0x0
  3a58c9:      	inc	r14d
  3a58cc:      	cmp	r15, r14
  3a58cf:      	jbe	0x3a72d9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cd9>
  3a58d5:      	mov	rax, r14
  3a58d8:      	shl	rax, 0x5
  3a58dc:      	lea	r12, [rcx + rax]
  3a58e0:      	inc	rbx
  3a58e3:      	cmp	dword ptr [rcx + rax + 0x1c], 0x0
  3a58e8:      	jne	0x3a5880 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x280>
  3a58ea:      	jmp	0x3a5940 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a58ec:      	mov	edi, 0x1
  3a58f1:      	call	0x61eac0 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a58f6:      	mov	rcx, qword ptr [rbp - 0x58]
  3a58fa:      	mov	r9, qword ptr [rbp - 0x48]
  3a58fe:      	mov	r10, qword ptr [rbp - 0x30]
  3a5902:      	mov	esi, 0xffffffff
  3a5907:      	test	rax, rax
  3a590a:      	jne	0x3a5817 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x217>
  3a5910:      	mov	r11, qword ptr [rbp - 0x38]
  3a5914:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a591a:      	nop	word ptr [rax + rax]
  3a5920:      	jne	0x3a5832 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x232>
  3a5926:      	nop	word ptr cs:[rax + rax]
  3a5935:      	nop	word ptr cs:[rax + rax]
  3a5940:      	mov	edx, dword ptr [r12]
  3a5944:      	lea	eax, [rdx - 0xe0]
  3a594a:      	cmp	eax, 0xffff20
  3a594f:      	jae	0x3a5c00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x600>
  3a5955:      	mov	eax, edx
  3a5957:      	shr	eax, 0x10
  3a595a:      	nop	word ptr [rax + rax]
  3a5960:      	cmp	eax, 0xfb
  3a5965:      	je	0x3a5b35 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x535>
  3a596b:      	cmp	eax, 0xfc
  3a5970:      	je	0x3a5a10 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x410>
  3a5976:      	nop	word ptr cs:[rax + rax]
  3a5980:      	cmp	eax, 0xfd
  3a5985:      	jne	0x3a5bc5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5c5>
  3a598b:      	mov	edi, dword ptr [r12 + 0xc]
  3a5990:      	cmp	rdi, rsi
  3a5993:      	je	0x3a5e8c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x88c>
  3a5999:      	nop	dword ptr [rax]
  3a59a0:      	cmp	r13, rdi
  3a59a3:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a59a9:      	shl	rdi, 0x4
  3a59ad:      	mov	rcx, qword ptr [r9 + rdi]
  3a59b1:      	mov	r10, qword ptr [r9 + rdi + 0x8]
  3a59b6:      	mov	edi, dword ptr [r12 + 0x10]
  3a59bb:      	nop	dword ptr [rax + rax]
  3a59c0:      	cmp	rdi, rsi
  3a59c3:      	je	0x3a5e9f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x89f>
  3a59c9:      	mov	esi, 0xffffffff
  3a59ce:      	cmp	r13, rdi
  3a59d1:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a59d7:      	shl	rdi, 0x4
  3a59db:      	mov	rax, qword ptr [r9 + rdi]
  3a59df:      	mov	r8, qword ptr [r9 + rdi + 0x8]
  3a59e4:      	mov	edi, dword ptr [r12 + 0x14]
  3a59e9:      	cmp	rdi, rsi
  3a59ec:      	je	0x3a5eb7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8b7>
  3a59f2:      	cmp	r13, rdi
  3a59f5:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a59fb:      	shl	rdi, 0x4
  3a59ff:      	mov	rsi, r9
  3a5a02:      	mov	r9, qword ptr [r9 + rdi]
  3a5a06:      	mov	r11, qword ptr [rsi + rdi + 0x8]
  3a5a0b:      	jmp	0x3a5ebd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8bd>
  3a5a10:      	add	edx, 0xff03ffed
  3a5a16:      	cmp	edx, 0x4
  3a5a19:      	jae	0x3a5c00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x600>
  3a5a1f:      	mov	rdx, qword ptr [r10 + 0xd8]
  3a5a26:      	mov	edi, dword ptr [r12 + 0x18]
  3a5a2b:      	mov	eax, dword ptr [r12 + 0x1c]
  3a5a30:      	lea	ecx, [rdi + rax]
  3a5a33:      	add	ecx, 0x2
  3a5a36:      	cmp	rdx, rcx
  3a5a39:      	jb	0x3a7399 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d99>
  3a5a3f:      	nop
  3a5a40:      	cmp	edi, ecx
  3a5a42:      	ja	0x3a738a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d8a>
  3a5a48:      	sub	rdx, rdi
  3a5a4b:      	lea	r8, [4*rdi]
  3a5a53:      	test	rdx, rdx
  3a5a56:      	mov	eax, 0x0
  3a5a5b:      	cmovle	r8, rax
  3a5a5f:      	nop
  3a5a60:      	cmp	ecx, edi
  3a5a62:      	je	0x3a7381 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d81>
  3a5a68:      	add	r8, qword ptr [r10 + 0xc8]
  3a5a6f:      	mov	eax, dword ptr [r8]
  3a5a72:      	cmp	r13, rax
  3a5a75:      	jbe	0x3a7322 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d22>
  3a5a7b:      	sub	rcx, rdi
  3a5a7e:      	nop
  3a5a80:      	cmp	ecx, 0x1
  3a5a83:      	je	0x3a7372 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d72>
  3a5a89:      	mov	edi, dword ptr [r8 + 0x4]
  3a5a8d:      	cmp	r13, rdi
  3a5a90:      	jbe	0x3a7312 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d12>
  3a5a96:      	shl	rax, 0x4
  3a5a9a:      	mov	rdx, r9
  3a5a9d:      	mov	r9, qword ptr [r9 + rax]
  3a5aa1:      	shl	rdi, 0x4
  3a5aa5:      	mov	rsi, qword ptr [rdx + rdi]
  3a5aa9:      	mov	r10d, dword ptr [r12]
  3a5aad:      	lea	eax, [r10 - 0xfc0013]
  3a5ab4:      	cmp	eax, 0x2
  3a5ab7:      	jae	0x3a600e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa0e>
  3a5abd:      	nop	dword ptr [rax]
  3a5ac0:      	cmp	ecx, 0x3
  3a5ac3:      	jb	0x3a73b7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1db7>
  3a5ac9:      	mov	edi, dword ptr [r8 + 0x8]
  3a5acd:      	cmp	r13, rdi
  3a5ad0:      	jbe	0x3a7312 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d12>
  3a5ad6:      	cmp	ecx, 0x3
  3a5ad9:      	je	0x3a73a8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1da8>
  3a5adf:      	mov	eax, dword ptr [r8 + 0xc]
  3a5ae3:      	cmp	r13, rax
  3a5ae6:      	jbe	0x3a7322 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d22>
  3a5aec:      	shl	rdi, 0x4
  3a5af0:      	mov	rdx, qword ptr [rbp - 0x48]
  3a5af4:      	mov	rdi, qword ptr [rdx + rdi]
  3a5af8:      	shl	rax, 0x4
  3a5afc:      	mov	rdx, qword ptr [rdx + rax]
  3a5b00:      	cmp	r10d, 0xfc0013
  3a5b07:      	jne	0x3a6155 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb55>
  3a5b0d:      	lea	rax, [rdi + r9]
  3a5b11:      	mov	r10, rdi
  3a5b14:      	and	r10, r9
  3a5b17:      	or	rdi, r9
  3a5b1a:      	mov	r9, rax
  3a5b1d:      	not	r9
  3a5b20:      	and	r9, rdi
  3a5b23:      	or	r9, r10
  3a5b26:      	shr	r9, 0x3f
  3a5b2a:      	add	rdx, rsi
  3a5b2d:      	add	rdx, r9
  3a5b30:      	jmp	0x3a617a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb7a>
  3a5b35:      	add	edx, -0x2
  3a5b38:      	nop	dword ptr [rax + rax]
  3a5b40:      	cmp	dx, 0x21
  3a5b44:      	jae	0x3a5d65 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x765>
  3a5b4a:      	lea	rax, [rsi + 0x1c003e10]
  3a5b51:      	movzx	ecx, dx
  3a5b54:      	bt	rax, rcx
  3a5b58:      	jae	0x3a5d65 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x765>
  3a5b5e:      	mov	edi, dword ptr [r12 + 0xc]
  3a5b63:      	mov	eax, 0xffffffff
  3a5b68:      	cmp	rdi, rsi
  3a5b6b:      	je	0x3a61f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf7>
  3a5b71:      	cmp	r13, rdi
  3a5b74:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a5b7a:      	shl	rdi, 0x4
  3a5b7e:      	mov	r8, qword ptr [r9 + rdi]
  3a5b82:      	mov	edi, dword ptr [r12 + 0x10]
  3a5b87:      	cmp	rdi, rax
  3a5b8a:      	je	0x3a6209 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc09>
  3a5b90:      	cmp	r13, rdi
  3a5b93:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a5b99:      	shl	rdi, 0x4
  3a5b9d:      	mov	rax, qword ptr [r9 + rdi]
  3a5ba1:      	mov	r9, qword ptr [r9 + rdi + 0x8]
  3a5ba6:      	mov	r10d, 0x0
  3a5bac:      	mov	r11d, 0x0
  3a5bb2:      	cmp	dx, 0xc
  3a5bb6:      	je	0x3a6226 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc26>
  3a5bbc:      	nop	dword ptr [rax]
  3a5bc0:      	jmp	0x3a626d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc6d>
  3a5bc5:      	cmp	edx, 0xe7
  3a5bcb:      	jb	0x3a70ea <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aea>
  3a5bd1:      	nop	word ptr cs:[rax + rax]
  3a5be0:      	nop	word ptr cs:[rax + rax]
  3a5bef:      	nop	word ptr cs:[rax + rax]
  3a5bfe:      	nop
  3a5c00:      	mov	edi, dword ptr [r12 + 0xc]
  3a5c05:      	cmp	rdi, rsi
  3a5c08:      	je	0x3a5c80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x680>
  3a5c0a:      	cmp	r13, rdi
  3a5c0d:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a5c13:      	mov	r8, r10
  3a5c16:      	mov	rax, rdi
  3a5c19:      	shl	rax, 0x4
  3a5c1d:      	mov	r10, qword ptr [r9 + rax]
  3a5c21:      	mov	ecx, dword ptr [r12 + 0x10]
  3a5c26:      	cmp	rcx, rsi
  3a5c29:      	je	0x3a5c90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x690>
  3a5c2b:      	cmp	r13, rcx
  3a5c2e:      	jbe	0x3a7307 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d07>
  3a5c34:      	mov	rax, rcx
  3a5c37:      	shl	rax, 0x4
  3a5c3b:      	mov	rdx, qword ptr [r9 + rax]
  3a5c3f:      	mov	r9d, dword ptr [r12]
  3a5c43:      	lea	eax, [r9 - 0x45]
  3a5c47:      	cmp	eax, 0x7f
  3a5c4a:      	ja	0x3a5c9f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x69f>
  3a5c4c:      	ja	0x3a6b50 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1550>
  3a5c52:      	jmp	qword ptr [8*rax + 0x203ab8]
  3a5c59:      	mov	edi, r9d
  3a5c5c:      	mov	rsi, r10
  3a5c5f:      	call	0x3d7240 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a5c64:      	test	dl, dl
  3a5c66:      	jne	0x3a71d1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd1>
  3a5c6c:      	xor	ecx, ecx
  3a5c6e:      	mov	r11, qword ptr [rbp - 0x38]
  3a5c72:      	mov	r8, qword ptr [rbp - 0x30]
  3a5c76:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a5c7b:      	nop	dword ptr [rax + rax]
  3a5c80:      	mov	r8, r10
  3a5c83:      	xor	r10d, r10d
  3a5c86:      	mov	ecx, dword ptr [r12 + 0x10]
  3a5c8b:      	cmp	rcx, rsi
  3a5c8e:      	jne	0x3a5c2b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x62b>
  3a5c90:      	xor	edx, edx
  3a5c92:      	mov	r9d, dword ptr [r12]
  3a5c96:      	lea	eax, [r9 - 0x45]
  3a5c9a:      	cmp	eax, 0x7f
  3a5c9d:      	jbe	0x3a5c4c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x64c>
  3a5c9f:      	lea	eax, [r9 - 0x1]
  3a5ca3:      	cmp	eax, 0xd3
  3a5ca8:      	mov	esi, 0xffffffff
  3a5cad:      	ja	0x3a5da5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7a5>
  3a5cb3:      	jmp	qword ptr [8*rax + 0x2033d8]
  3a5cba:      	mov	rsi, qword ptr [r8 + 0x80]
  3a5cc1:      	mov	edi, dword ptr [r12 + 0x4]
  3a5cc6:      	cmp	rsi, rdi
  3a5cc9:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a5ccf:      	imul	rax, rdi, 0xe0
  3a5cd6:      	nop	word ptr cs:[rax + rax]
  3a5ce0:      	add	rax, qword ptr [r8 + 0x78]
  3a5ce4:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5cea:      	mov	rcx, qword ptr [rbp - 0x50]
  3a5cee:      	mov	rsi, qword ptr [rcx + 0x78]
  3a5cf2:      	mov	edi, dword ptr [rax + 0x18]
  3a5cf5:      	cmp	rsi, rdi
  3a5cf8:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a5cfe:      	shl	rdi, 0x3
  3a5d02:      	add	rdi, qword ptr [rcx + 0x70]
  3a5d06:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5d0c:      	mov	rcx, qword ptr [rdi]
  3a5d0f:      	test	rcx, rcx
  3a5d12:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5d18:      	mov	rsi, qword ptr [rcx + 0x10]
  3a5d1c:      	mov	ecx, dword ptr [rax + 0x14]
  3a5d1f:      	mov	r8d, dword ptr [rax + 0x1c]
  3a5d23:      	shl	r8, 0x20
  3a5d27:      	or	r8, rcx
  3a5d2a:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a5d2f:      	je	0x3a5d3c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x73c>
  3a5d31:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a5d36:      	je	0x3a6130 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb30>
  3a5d3c:      	xorps	xmm0, xmm0
  3a5d3f:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a5d46:      	movups	xmmword ptr [rsp], xmm0
  3a5d4a:      	mov	r15, qword ptr [rbp - 0x68]
  3a5d4e:      	mov	rdi, r15
  3a5d51:      	mov	edx, r9d
  3a5d54:      	mov	rcx, r10
  3a5d57:      	xor	r9d, r9d
  3a5d5a:      	call	0x3cc700 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a5d5f:      	nop
  3a5d60:      	jmp	0x3a6111 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb11>
  3a5d65:      	mov	r15, qword ptr [rbp - 0x60]
  3a5d69:      	mov	dword ptr [r15 - 0x38], r14d
  3a5d6d:      	mov	rdi, r11
  3a5d70:      	call	0x3a76c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a5d75:      	mov	r9, qword ptr [rbp - 0x48]
  3a5d79:      	mov	r10, qword ptr [rbp - 0x30]
  3a5d7d:      	mov	esi, 0xffffffff
  3a5d82:      	mov	r11, qword ptr [rbp - 0x38]
  3a5d86:      	test	al, al
  3a5d88:      	mov	rcx, qword ptr [rbp - 0x58]
  3a5d8c:      	jne	0x3a72af <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1caf>
  3a5d92:      	inc	r14d
  3a5d95:      	mov	dword ptr [r15 - 0x38], r14d
  3a5d99:      	mov	r15d, r14d
  3a5d9c:      	nop	dword ptr [rax]
  3a5da0:      	jmp	0x3a570d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10d>
  3a5da5:      	lea	eax, [r9 - 0xfc0000]
  3a5dac:      	cmp	eax, 0x8
  3a5daf:      	jae	0x3a5fa8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9a8>
  3a5db5:      	mov	edi, r9d
  3a5db8:      	mov	rsi, r10
  3a5dbb:      	call	0x3d7240 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a5dc0:      	mov	r8, qword ptr [rbp - 0x30]
  3a5dc4:      	mov	r11, qword ptr [rbp - 0x38]
  3a5dc8:      	mov	r15d, edx
  3a5dcb:      	xor	ecx, ecx
  3a5dcd:      	jmp	0x3a6bb0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b0>
  3a5dd2:      	xor	eax, eax
  3a5dd4:      	xor	edx, edx
  3a5dd6:      	test	rax, rax
  3a5dd9:      	jne	0x3a5817 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x217>
  3a5ddf:      	nop
  3a5de0:      	jmp	0x3a5910 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x310>
  3a5de5:      	mov	rsi, qword ptr [r8 + 0x80]
  3a5dec:      	mov	edi, dword ptr [r12 + 0x4]
  3a5df1:      	cmp	rsi, rdi
  3a5df4:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a5dfa:      	imul	rax, rdi, 0xe0
  3a5e01:      	add	rax, qword ptr [r8 + 0x78]
  3a5e05:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5e0b:      	mov	rcx, qword ptr [rbp - 0x50]
  3a5e0f:      	mov	rsi, qword ptr [rcx + 0x78]
  3a5e13:      	mov	edi, dword ptr [rax + 0x18]
  3a5e16:      	cmp	rsi, rdi
  3a5e19:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a5e1f:      	shl	rdi, 0x3
  3a5e23:      	add	rdi, qword ptr [rcx + 0x70]
  3a5e27:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5e2d:      	mov	rcx, qword ptr [rdi]
  3a5e30:      	test	rcx, rcx
  3a5e33:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5e39:      	mov	rdi, qword ptr [rcx + 0x10]
  3a5e3d:      	mov	esi, dword ptr [rax + 0x14]
  3a5e40:      	mov	ecx, dword ptr [rax + 0x1c]
  3a5e43:      	shl	rcx, 0x20
  3a5e47:      	or	rcx, rsi
  3a5e4a:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a5e4f:      	je	0x3a5e5c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x85c>
  3a5e51:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a5e56:      	je	0x3a63bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdbb>
  3a5e5c:      	mov	qword ptr [rbp - 0x90], rdx
  3a5e63:      	mov	qword ptr [rbp - 0x88], 0x0
  3a5e6e:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a5e75:      	movups	xmmword ptr [rsp], xmm0
  3a5e79:      	mov	esi, r9d
  3a5e7c:      	mov	rdx, r10
  3a5e7f:      	xor	r8d, r8d
  3a5e82:      	call	0x3cdd60 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a5e87:      	jmp	0x3a6ba1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15a1>
  3a5e8c:      	xor	r10d, r10d
  3a5e8f:      	xor	ecx, ecx
  3a5e91:      	mov	edi, dword ptr [r12 + 0x10]
  3a5e96:      	cmp	rdi, rsi
  3a5e99:      	jne	0x3a59c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3c9>
  3a5e9f:      	mov	esi, 0xffffffff
  3a5ea4:      	xor	eax, eax
  3a5ea6:      	xor	r8d, r8d
  3a5ea9:      	mov	edi, dword ptr [r12 + 0x14]
  3a5eae:      	cmp	rdi, rsi
  3a5eb1:      	jne	0x3a59f2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3f2>
  3a5eb7:      	xor	r9d, r9d
  3a5eba:      	xor	r11d, r11d
  3a5ebd:      	mov	r15, qword ptr [rbp - 0x30]
  3a5ec1:      	mov	rsi, qword ptr [r15 + 0x80]
  3a5ec8:      	mov	edi, dword ptr [r12 + 0x4]
  3a5ecd:      	cmp	rsi, rdi
  3a5ed0:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a5ed6:      	imul	rdi, rdi, 0xe0
  3a5edd:      	nop	dword ptr [rax]
  3a5ee0:      	add	rdi, qword ptr [r15 + 0x78]
  3a5ee4:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5eea:      	movzx	esi, dx
  3a5eed:      	cmp	esi, 0xb
  3a5ef0:      	setb	r15b
  3a5ef4:      	mov	esi, edx
  3a5ef6:      	and	esi, 0xfffc
  3a5efc:      	cmp	esi, 0x54
  3a5eff:      	sete	r13b
  3a5f03:      	or	r13b, r15b
  3a5f06:      	jne	0x3a5f1a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x91a>
  3a5f08:      	movzx	r15d, dx
  3a5f0c:      	lea	r13d, [r15 - 0x5c]
  3a5f10:      	cmp	r13d, 0x2
  3a5f14:      	jae	0x3a6045 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa45>
  3a5f1a:      	mov	r10, qword ptr [rbp - 0x50]
  3a5f1e:      	mov	rsi, qword ptr [r10 + 0x78]
  3a5f22:      	mov	r9d, dword ptr [rdi + 0x18]
  3a5f26:      	cmp	rsi, r9
  3a5f29:      	jbe	0x3a731a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d1a>
  3a5f2f:      	shl	r9, 0x3
  3a5f33:      	add	r9, qword ptr [r10 + 0x70]
  3a5f37:      	mov	r15, qword ptr [rbp - 0x68]
  3a5f3b:      	nop	dword ptr [rax + rax]
  3a5f40:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5f46:      	mov	rsi, qword ptr [r9]
  3a5f49:      	test	rsi, rsi
  3a5f4c:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a5f52:      	mov	rsi, qword ptr [rsi + 0x10]
  3a5f56:      	mov	r9d, dword ptr [rdi + 0x14]
  3a5f5a:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a5f5e:      	shl	r10, 0x20
  3a5f62:      	or	r10, r9
  3a5f65:      	movzx	edi, byte ptr [rdi + 0x40]
  3a5f69:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a5f6e:      	je	0x3a5f7b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x97b>
  3a5f70:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a5f75:      	je	0x3a60e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xae5>
  3a5f7b:      	mov	qword ptr [rbp - 0xa0], rax
  3a5f82:      	mov	qword ptr [rbp - 0x98], r8
  3a5f89:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a5f90:      	movups	xmmword ptr [rsp], xmm0
  3a5f94:      	movzx	r9d, dil
  3a5f98:      	mov	rdi, r15
  3a5f9b:      	mov	r8, r10
  3a5f9e:      	call	0x3cc700 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a5fa3:      	jmp	0x3a610d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb0d>
  3a5fa8:      	lea	eax, [r9 - 0x1000001]
  3a5faf:      	cmp	eax, 0x8
  3a5fb2:      	ja	0x3a6d57 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1757>
  3a5fb8:      	mov	ecx, 0x117
  3a5fbd:      	bt	ecx, eax
  3a5fc0:      	jae	0x3a6d45 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1745>
  3a5fc6:      	movzx	eax, r9b
  3a5fca:      	xor	edi, edi
  3a5fcc:      	add	eax, -0x2
  3a5fcf:      	cmp	eax, 0x7
  3a5fd2:      	ja	0x3a6c45 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1645>
  3a5fd8:      	mov	r9, qword ptr [rbp - 0x48]
  3a5fdc:      	nop	dword ptr [rax]
  3a5fe0:      	jmp	qword ptr [8*rax + 0x203a78]
  3a5fe7:      	xor	edi, edi
  3a5fe9:      	test	r10d, r10d
  3a5fec:      	sete	dil
  3a5ff0:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a5ff7:      	add	edi, dword ptr [r12 + 0x18]
  3a5ffc:      	nop	dword ptr [rax]
  3a6000:      	cmp	rsi, rdi
  3a6003:      	ja	0x3a648d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe8d>
  3a6009:      	jmp	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a600e:      	mov	rax, rsi
  3a6011:      	mul	r9
  3a6014:      	cmp	dword ptr [r12], 0xfc0015
  3a601c:      	nop	dword ptr [rax]
  3a6020:      	jne	0x3a617a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb7a>
  3a6026:      	mov	rdi, r9
  3a6029:      	sar	rdi, 0x3f
  3a602d:      	and	rdi, rsi
  3a6030:      	sar	rsi, 0x3f
  3a6034:      	and	rsi, r9
  3a6037:      	add	rsi, rdi
  3a603a:      	sub	rdx, rsi
  3a603d:      	nop	dword ptr [rax]
  3a6040:      	jmp	0x3a617a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb7a>
  3a6045:      	cmp	r15d, 0xb
  3a6049:      	je	0x3a6054 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa54>
  3a604b:      	cmp	esi, 0x58
  3a604e:      	jne	0x3a6425 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe25>
  3a6054:      	mov	r10, qword ptr [rbp - 0x50]
  3a6058:      	mov	rsi, qword ptr [r10 + 0x78]
  3a605c:      	mov	r9d, dword ptr [rdi + 0x18]
  3a6060:      	cmp	rsi, r9
  3a6063:      	mov	r13, qword ptr [rbp - 0x78]
  3a6067:      	jbe	0x3a731a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d1a>
  3a606d:      	shl	r9, 0x3
  3a6071:      	add	r9, qword ptr [r10 + 0x70]
  3a6075:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a607b:      	mov	rsi, qword ptr [r9]
  3a607e:      	nop
  3a6080:      	test	rsi, rsi
  3a6083:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6089:      	mov	rsi, qword ptr [rsi + 0x10]
  3a608d:      	mov	r10d, dword ptr [rdi + 0x14]
  3a6091:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a6095:      	shl	r9, 0x20
  3a6099:      	or	r9, r10
  3a609c:      	movzx	edi, byte ptr [rdi + 0x40]
  3a60a0:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a60a5:      	je	0x3a60b2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xab2>
  3a60a7:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a60ac:      	je	0x3a6b74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1574>
  3a60b2:      	mov	qword ptr [rbp - 0x90], rax
  3a60b9:      	mov	qword ptr [rbp - 0x88], r8
  3a60c0:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a60c7:      	movups	xmmword ptr [rsp], xmm0
  3a60cb:      	movzx	r8d, dil
  3a60cf:      	mov	rdi, rsi
  3a60d2:      	mov	esi, edx
  3a60d4:      	mov	rdx, rcx
  3a60d7:      	mov	rcx, r9
  3a60da:      	call	0x3cdd60 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a60df:      	nop
  3a60e0:      	jmp	0x3a6ba1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15a1>
  3a60e5:      	mov	qword ptr [rbp - 0xa0], rax
  3a60ec:      	mov	qword ptr [rbp - 0x98], r8
  3a60f3:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a60fa:      	movups	xmmword ptr [rsp], xmm0
  3a60fe:      	movzx	r9d, dil
  3a6102:      	mov	rdi, r15
  3a6105:      	mov	r8, r10
  3a6108:      	call	0x3cc860 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a610d:      	mov	r13, qword ptr [rbp - 0x78]
  3a6111:      	mov	rax, qword ptr [r15]
  3a6114:      	mov	rcx, qword ptr [r15 + 0x8]
  3a6118:      	mov	qword ptr [rbp - 0x68], r15
  3a611c:      	movzx	r15d, byte ptr [r15 + 0x10]
  3a6121:      	jmp	0x3a6ba8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15a8>
  3a6126:      	mov	rax, qword ptr [r12 + 0x18]
  3a612b:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6130:      	xorps	xmm0, xmm0
  3a6133:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a613a:      	movups	xmmword ptr [rsp], xmm0
  3a613e:      	mov	r15, qword ptr [rbp - 0x68]
  3a6142:      	mov	rdi, r15
  3a6145:      	mov	edx, r9d
  3a6148:      	mov	rcx, r10
  3a614b:      	xor	r9d, r9d
  3a614e:      	call	0x3cc860 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a6153:      	jmp	0x3a6111 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb11>
  3a6155:      	mov	rax, r9
  3a6158:      	sub	rax, rdi
  3a615b:      	not	r9
  3a615e:      	mov	r10, rdi
  3a6161:      	and	r10, r9
  3a6164:      	xor	r9, rdi
  3a6167:      	and	r9, rax
  3a616a:      	or	r9, r10
  3a616d:      	shr	r9, 0x3f
  3a6171:      	add	rdx, r9
  3a6174:      	sub	rsi, rdx
  3a6177:      	mov	rdx, rsi
  3a617a:      	mov	r10, qword ptr [rbp - 0x30]
  3a617e:      	mov	edi, dword ptr [r12 + 0x1c]
  3a6183:      	cmp	ecx, edi
  3a6185:      	jbe	0x3a7354 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d54>
  3a618b:      	lea	esi, [rdi + 0x1]
  3a618e:      	cmp	ecx, esi
  3a6190:      	mov	r9, qword ptr [rbp - 0x60]
  3a6194:      	jbe	0x3a735c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d5c>
  3a619a:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a619e:      	nop
  3a61a0:      	cmp	r13, rcx
  3a61a3:      	jbe	0x3a7367 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d67>
  3a61a9:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a61ad:      	shl	rcx, 0x4
  3a61b1:      	mov	rsi, qword ptr [rbp - 0x48]
  3a61b5:      	mov	qword ptr [rsi + rcx], rax
  3a61b9:      	mov	qword ptr [rsi + rcx + 0x8], 0x0
  3a61c2:      	cmp	r13, rdi
  3a61c5:      	jbe	0x3a7312 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d12>
  3a61cb:      	shl	rdi, 0x4
  3a61cf:      	mov	qword ptr [rsi + rdi], rdx
  3a61d3:      	mov	qword ptr [rsi + rdi + 0x8], 0x0
  3a61dc:      	inc	r14d
  3a61df:      	mov	dword ptr [r9 - 0x38], r14d
  3a61e3:      	mov	r15d, r14d
  3a61e6:      	mov	r9, rsi
  3a61e9:      	mov	esi, 0xffffffff
  3a61ee:      	mov	rcx, qword ptr [rbp - 0x58]
  3a61f2:      	jmp	0x3a570d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10d>
  3a61f7:      	xor	r8d, r8d
  3a61fa:      	mov	edi, dword ptr [r12 + 0x10]
  3a61ff:      	nop
  3a6200:      	cmp	rdi, rax
  3a6203:      	jne	0x3a5b90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x590>
  3a6209:      	xor	eax, eax
  3a620b:      	xor	r9d, r9d
  3a620e:      	mov	r10d, 0x0
  3a6214:      	mov	r11d, 0x0
  3a621a:      	nop	word ptr [rax + rax]
  3a6220:      	cmp	dx, 0xc
  3a6224:      	jne	0x3a626d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc6d>
  3a6226:      	mov	rcx, qword ptr [rbp - 0x30]
  3a622a:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a6231:      	mov	edi, dword ptr [r12 + 0x18]
  3a6236:      	add	edi, 0x2
  3a6239:      	nop	dword ptr [rax]
  3a6240:      	cmp	rsi, rdi
  3a6243:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a6249:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a6250:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a6253:      	cmp	r13, rdi
  3a6256:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a625c:      	shl	rdi, 0x4
  3a6260:      	mov	rcx, qword ptr [rbp - 0x48]
  3a6264:      	mov	r10, qword ptr [rcx + rdi]
  3a6268:      	mov	r11, qword ptr [rcx + rdi + 0x8]
  3a626d:      	cmp	qword ptr [rbp - 0x38], 0x0
  3a6272:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6278:      	mov	rdx, qword ptr [rbp - 0x30]
  3a627c:      	mov	rsi, qword ptr [rdx + 0x80]
  3a6283:      	mov	edi, dword ptr [r12 + 0x4]
  3a6288:      	cmp	rsi, rdi
  3a628b:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a6291:      	imul	rcx, rdi, 0xe0
  3a6298:      	nop	dword ptr [rax + rax]
  3a62a0:      	add	rcx, qword ptr [rdx + 0x78]
  3a62a4:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a62aa:      	mov	rdx, qword ptr [rbp - 0x38]
  3a62ae:      	mov	rsi, qword ptr [rdx]
  3a62b1:      	mov	qword ptr [rbp - 0xd0], rax
  3a62b8:      	mov	qword ptr [rbp - 0xc8], r9
  3a62bf:      	mov	qword ptr [rbp - 0xc0], r10
  3a62c6:      	mov	qword ptr [rbp - 0xb8], r11
  3a62cd:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a62d4:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a62d9:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a62e0:      	movups	xmmword ptr [rsp], xmm0
  3a62e4:      	lea	rdi, [rbp - 0x160]
  3a62eb:      	mov	rdx, qword ptr [rbp - 0x50]
  3a62ef:      	call	0x3bb0c0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a62f4:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a62fb:      	mov	r15, qword ptr [rbp - 0x150]
  3a6302:      	test	r15, r15
  3a6305:      	je	0x3a63e6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xde6>
  3a630b:      	cmp	qword ptr [r15 + 0x8], 0x0
  3a6310:      	mov	esi, 0xffffffff
  3a6315:      	mov	r10, qword ptr [rbp - 0x30]
  3a6319:      	mov	r9, qword ptr [rbp - 0x48]
  3a631d:      	mov	rcx, qword ptr [rbp - 0x58]
  3a6321:      	je	0x3a63f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdf7>
  3a6327:      	jmp	0x3a71d6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd6>
  3a632c:      	cmp	r13, rdi
  3a632f:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a6335:      	mov	edx, dword ptr [r12 + 0x18]
  3a633a:      	cmp	r13, rdx
  3a633d:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6341:      	jbe	0x3a73e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1de7>
  3a6347:      	shl	rdi, 0x4
  3a634b:      	mov	rax, qword ptr [rsi + rdi]
  3a634f:      	mov	rcx, qword ptr [rsi + rdi + 0x8]
  3a6354:      	shl	rdx, 0x4
  3a6358:      	mov	qword ptr [rsi + rdx], rax
  3a635c:      	mov	qword ptr [rsi + rdx + 0x8], rcx
  3a6361:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a6366:      	mov	eax, dword ptr [r12 + 0x14]
  3a636b:      	cmp	rax, rsi
  3a636e:      	je	0x3a6beb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15eb>
  3a6374:      	cmp	r13, rax
  3a6377:      	mov	rdx, qword ptr [rbp - 0x48]
  3a637b:      	nop	dword ptr [rax + rax]
  3a6380:      	jbe	0x3a72fc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cfc>
  3a6386:      	cmp	r13, rcx
  3a6389:      	jbe	0x3a73f2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1df2>
  3a638f:      	shl	rax, 0x4
  3a6393:      	cmp	dword ptr [rdx + rax], 0x0
  3a6397:      	je	0x3a6bf8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15f8>
  3a639d:      	nop	dword ptr [rax]
  3a63a0:      	cmp	r13, rdi
  3a63a3:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a63a9:      	shl	rdi, 0x4
  3a63ad:      	mov	rax, qword ptr [rdx + rdi]
  3a63b1:      	mov	rcx, qword ptr [rdx + rdi + 0x8]
  3a63b6:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a63bb:      	mov	qword ptr [rbp - 0x90], rdx
  3a63c2:      	mov	qword ptr [rbp - 0x88], 0x0
  3a63cd:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a63d4:      	movups	xmmword ptr [rsp], xmm0
  3a63d8:      	mov	esi, r9d
  3a63db:      	mov	rdx, r10
  3a63de:      	xor	r8d, r8d
  3a63e1:      	jmp	0x3a6b9c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x159c>
  3a63e6:      	mov	esi, 0xffffffff
  3a63eb:      	mov	r10, qword ptr [rbp - 0x30]
  3a63ef:      	mov	r9, qword ptr [rbp - 0x48]
  3a63f3:      	mov	rcx, qword ptr [rbp - 0x58]
  3a63f7:      	mov	edi, dword ptr [r12 + 0x8]
  3a63fc:      	cmp	rdi, rsi
  3a63ff:      	mov	r11, qword ptr [rbp - 0x38]
  3a6403:      	mov	r15, qword ptr [rbp - 0x60]
  3a6407:      	je	0x3a5d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x792>
  3a640d:      	cmp	r13, rdi
  3a6410:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a6416:      	shl	rdi, 0x4
  3a641a:      	movups	xmmword ptr [r9 + rdi], xmm0
  3a641f:      	nop
  3a6420:      	jmp	0x3a5d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x792>
  3a6425:      	mov	qword ptr [rsp], r11
  3a6429:      	mov	rsi, rcx
  3a642c:      	mov	rdx, r10
  3a642f:      	mov	rcx, rax
  3a6432:      	call	0x3d5200 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a6437:      	mov	rcx, rdx
  3a643a:      	mov	r11, qword ptr [rbp - 0x38]
  3a643e:      	mov	r13, qword ptr [rbp - 0x78]
  3a6442:      	mov	r8, qword ptr [rbp - 0x30]
  3a6446:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a644b:      	mov	edi, dword ptr [r12 + 0x1c]
  3a6450:      	dec	edi
  3a6452:      	cmp	edi, r10d
  3a6455:      	cmovae	edi, r10d
  3a6459:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a6460:      	add	edi, dword ptr [r12 + 0x18]
  3a6465:      	cmp	rsi, rdi
  3a6468:      	ja	0x3a648d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe8d>
  3a646a:      	jmp	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a646f:      	xor	edi, edi
  3a6471:      	test	r10, r10
  3a6474:      	setne	dil
  3a6478:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a647f:      	add	edi, dword ptr [r12 + 0x18]
  3a6484:      	cmp	rsi, rdi
  3a6487:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a648d:      	mov	r10, r8
  3a6490:      	mov	rax, qword ptr [r8 + 0xe0]
  3a6497:      	lea	rcx, [rdi + 2*rdi]
  3a649b:      	mov	r15d, dword ptr [rax + 4*rcx]
  3a649f:      	mov	r12d, dword ptr [rax + 4*rcx + 0x4]
  3a64a4:      	mov	esi, dword ptr [rax + 4*rcx + 0x8]
  3a64a8:      	cmp	r15d, r14d
  3a64ab:      	ja	0x3a6c89 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1689>
  3a64b1:      	mov	rax, qword ptr [r11 + 0x160]
  3a64b8:      	test	rax, rax
  3a64bb:      	je	0x3a64d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xed3>
  3a64bd:      	nop	dword ptr [rax]
  3a64c0:      	test	r11, r11
  3a64c3:      	je	0x3a6c65 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1665>
  3a64c9:      	mov	eax, dword ptr [rax]
  3a64cb:      	test	eax, eax
  3a64cd:      	je	0x3a6fa5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19a5>
  3a64d3:      	mov	rax, qword ptr [r11 + 0x8]
  3a64d7:      	nop	word ptr [rax + rax]
  3a64e0:      	test	rax, rax
  3a64e3:      	je	0x3a7406 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e06>
  3a64e9:      	mov	rdi, qword ptr [r11 + 0x10]
  3a64ed:      	mov	rcx, qword ptr [rax + 0x8]
  3a64f1:      	test	byte ptr [rcx + 0x17], 0x20
  3a64f5:      	jne	0x3a6c07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1607>
  3a64fb:      	mov	r14, rsi
  3a64fe:      	nop
  3a6500:      	jmp	0x3a6c21 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1621>
  3a6505:      	xor	eax, eax
  3a6507:      	cmp	r10, rdx
  3a650a:      	sete	al
  3a650d:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6512:      	mov	rax, qword ptr [rbp - 0x50]
  3a6516:      	mov	rsi, qword ptr [rax + 0x78]
  3a651a:      	mov	edi, dword ptr [r12 + 0x18]
  3a651f:      	nop
  3a6520:      	cmp	rsi, rdi
  3a6523:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a6529:      	shl	rdi, 0x3
  3a652d:      	add	rdi, qword ptr [rax + 0x70]
  3a6531:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6537:      	nop	word ptr [rax + rax]
  3a6540:      	test	r11, r11
  3a6543:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6549:      	mov	r15, qword ptr [rdi]
  3a654c:      	test	r15, r15
  3a654f:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6555:      	mov	rdi, qword ptr [r11]
  3a6558:      	mov	rax, qword ptr [r15 + 0x10]
  3a655c:      	movzx	eax, byte ptr [rax + 0x31]
  3a6560:      	mov	byte ptr [rbp - 0x69], al
  3a6563:      	mov	edx, r10d
  3a6566:      	test	al, 0x1
  3a6568:      	cmovne	rdx, r10
  3a656c:      	mov	rsi, r15
  3a656f:      	call	0x3c62c0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a6574:      	test	dl, 0x1
  3a6577:      	jne	0x3a5c6c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x66c>
  3a657d:      	mov	rax, qword ptr [r15 + 0x10]
  3a6581:      	movzx	eax, byte ptr [rax + 0x31]
  3a6585:      	mov	byte ptr [rbp - 0x6a], al
  3a6588:      	and	al, 0x1
  3a658a:      	neg	al
  3a658c:      	mov	eax, 0x0
  3a6591:      	sbb	rax, rax
  3a6594:      	mov	ecx, 0xffffffff
  3a6599:      	or	rax, rcx
  3a659c:      	nop	dword ptr [rax]
  3a65a0:      	jmp	0x3a5c6c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x66c>
  3a65a5:      	mov	rax, qword ptr [rbp - 0x50]
  3a65a9:      	mov	rsi, qword ptr [rax + 0x60]
  3a65ad:      	mov	edi, dword ptr [r12 + 0x18]
  3a65b2:      	cmp	rsi, rdi
  3a65b5:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a65bb:      	mov	rax, qword ptr [rax + 0x58]
  3a65bf:      	lea	rcx, [rax + 8*rdi]
  3a65c3:      	test	rcx, rcx
  3a65c6:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a65cc:      	mov	rax, qword ptr [rax + 8*rdi]
  3a65d0:      	test	rax, rax
  3a65d3:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a65d9:      	cmp	byte ptr [rax + 0x29], 0x0
  3a65dd:      	mov	ecx, r10d
  3a65e0:      	cmovne	rcx, r10
  3a65e4:      	cmp	rcx, qword ptr [rax + 0x40]
  3a65e8:      	jae	0x3a722e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c2e>
  3a65ee:      	mov	rax, qword ptr [rax + 0x38]
  3a65f2:      	mov	rax, qword ptr [rax + 8*rcx]
  3a65f6:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a65fb:      	mov	rax, qword ptr [rbp - 0x50]
  3a65ff:      	mov	rsi, qword ptr [rax + 0x48]
  3a6603:      	mov	edi, dword ptr [r12 + 0x18]
  3a6608:      	cmp	rsi, rdi
  3a660b:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a6611:      	mov	rax, qword ptr [rax + 0x40]
  3a6615:      	mov	rax, qword ptr [rax + 8*rdi]
  3a6619:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a661e:      	nop
  3a6620:      	test	r10, r10
  3a6623:      	jne	0x3a670c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x110c>
  3a6629:      	jmp	0x3a7238 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c38>
  3a662e:      	mov	rax, qword ptr [rbp - 0x50]
  3a6632:      	mov	rsi, qword ptr [rax + 0x90]
  3a6639:      	mov	eax, dword ptr [r12 + 0x18]
  3a663e:      	nop
  3a6640:      	cmp	rsi, rax
  3a6643:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dfe>
  3a6649:      	cmp	r13, rdi
  3a664c:      	mov	rdx, qword ptr [rbp - 0x48]
  3a6650:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a6656:      	mov	rcx, qword ptr [rbp - 0x50]
  3a665a:      	mov	rcx, qword ptr [rcx + 0x88]
  3a6661:      	mov	rax, qword ptr [rcx + 8*rax]
  3a6665:      	shl	rdi, 0x4
  3a6669:      	movups	xmm0, xmmword ptr [rdx + rdi]
  3a666d:      	movups	xmmword ptr [rax + 0x20], xmm0
  3a6671:      	jmp	0x3a66f8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10f8>
  3a6676:      	mov	rax, qword ptr [rbp - 0x50]
  3a667a:      	mov	rsi, qword ptr [rax + 0x90]
  3a6681:      	mov	edi, dword ptr [r12 + 0x18]
  3a6686:      	cmp	rsi, rdi
  3a6689:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a668f:      	mov	rax, qword ptr [rax + 0x88]
  3a6696:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a669a:      	mov	rax, qword ptr [rcx + 0x20]
  3a669e:      	mov	rcx, qword ptr [rcx + 0x28]
  3a66a2:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a66a7:      	mov	rax, qword ptr [rbp - 0x50]
  3a66ab:      	mov	rsi, qword ptr [rax + 0x60]
  3a66af:      	mov	edi, dword ptr [r12 + 0x18]
  3a66b4:      	cmp	rsi, rdi
  3a66b7:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a66bd:      	mov	rax, qword ptr [rax + 0x58]
  3a66c1:      	lea	rcx, [rax + 8*rdi]
  3a66c5:      	test	rcx, rcx
  3a66c8:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a66ce:      	mov	rax, qword ptr [rax + 8*rdi]
  3a66d2:      	test	rax, rax
  3a66d5:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a66db:      	cmp	byte ptr [rax + 0x29], 0x0
  3a66df:      	mov	ecx, r10d
  3a66e2:      	cmovne	rcx, r10
  3a66e6:      	cmp	rcx, qword ptr [rax + 0x40]
  3a66ea:      	jae	0x3a722e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c2e>
  3a66f0:      	mov	rax, qword ptr [rax + 0x38]
  3a66f4:      	mov	qword ptr [rax + 8*rcx], rdx
  3a66f8:      	xor	eax, eax
  3a66fa:      	nop	word ptr [rax + rax]
  3a6700:      	xor	ecx, ecx
  3a6702:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a6707:      	mov	edi, dword ptr [r12 + 0x18]
  3a670c:      	cmp	r13, rdi
  3a670f:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a6715:      	shl	rdi, 0x4
  3a6719:      	mov	rcx, qword ptr [rbp - 0x48]
  3a671d:      	mov	rax, qword ptr [rcx + rdi]
  3a6721:      	mov	rcx, qword ptr [rcx + rdi + 0x8]
  3a6726:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a672b:      	xor	eax, eax
  3a672d:      	test	r10, r10
  3a6730:      	sete	al
  3a6733:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6735:      	mov	rax, qword ptr [rbp - 0x50]
  3a6739:      	mov	rsi, qword ptr [rax + 0x78]
  3a673d:      	mov	edi, dword ptr [r12 + 0x18]
  3a6742:      	cmp	rsi, rdi
  3a6745:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a674b:      	shl	rdi, 0x3
  3a674f:      	add	rdi, qword ptr [rax + 0x70]
  3a6753:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6759:      	mov	rax, qword ptr [rdi]
  3a675c:      	nop	dword ptr [rax]
  3a6760:      	test	rax, rax
  3a6763:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6769:      	mov	rdi, qword ptr [rax + 0x10]
  3a676d:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a6772:      	mov	rax, qword ptr [rdi + 0x60]
  3a6776:      	je	0x3a6f65 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1965>
  3a677c:      	nop	dword ptr [rax]
  3a6780:      	test	rax, rax
  3a6783:      	jne	0x3a6f6e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196e>
  3a6789:      	call	0x3ca140 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a678e:      	mov	r8, qword ptr [rbp - 0x30]
  3a6792:      	mov	r11, qword ptr [rbp - 0x38]
  3a6796:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a679b:      	mov	eax, dword ptr [r12 + 0x18]
  3a67a0:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a67a5:      	xor	rdx, r10
  3a67a8:      	jmp	0x3a6b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a67ad:      	mov	eax, r10d
  3a67b0:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a67b5:      	mov	ecx, edx
  3a67b7:      	shr	r10, cl
  3a67ba:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a67bf:      	mov	ecx, edx
  3a67c1:      	shr	r10d, cl
  3a67c4:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a67c9:      	or	rdx, r10
  3a67cc:      	jmp	0x3a6b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a67d1:      	mov	ecx, edx
  3a67d3:      	sar	r10d, cl
  3a67d6:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a67db:      	xor	edx, r10d
  3a67de:      	nop
  3a67e0:      	jmp	0x3a6b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a67e5:      	mov	rax, r10
  3a67e8:      	shr	rax
  3a67eb:      	movabs	rcx, 0x5555555555555555
  3a67f5:      	and	rax, rcx
  3a67f8:      	and	r10, rcx
  3a67fb:      	add	r10, rax
  3a67fe:      	mov	rax, r10
  3a6801:      	shr	rax, 0x2
  3a6805:      	movabs	rcx, 0x3333333333333333
  3a680f:      	and	rax, rcx
  3a6812:      	and	r10, rcx
  3a6815:      	add	r10, rax
  3a6818:      	mov	rax, r10
  3a681b:      	shr	rax, 0x4
  3a681f:      	add	rax, r10
  3a6822:      	movabs	rcx, 0xf0f0f0f0f0f0f0f
  3a682c:      	and	rcx, rax
  3a682f:      	mov	rax, rcx
  3a6832:      	shr	rax, 0x8
  3a6836:      	add	rax, rcx
  3a6839:      	mov	rcx, rax
  3a683c:      	shr	rcx, 0x10
  3a6840:      	add	rcx, rax
  3a6843:      	mov	rax, rcx
  3a6846:      	shr	rax, 0x20
  3a684a:      	add	eax, ecx
  3a684c:      	and	eax, 0x7f
  3a684f:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6854:      	xor	eax, eax
  3a6856:      	cmp	r10, rdx
  3a6859:      	setb	al
  3a685c:      	nop	dword ptr [rax]
  3a6860:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6865:      	xor	eax, eax
  3a6867:      	cmp	r10, rdx
  3a686a:      	setbe	al
  3a686d:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6872:      	xor	eax, eax
  3a6874:      	cmp	r10, rdx
  3a6877:      	seta	al
  3a687a:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a687f:      	imul	edx, r10d
  3a6883:      	jmp	0x3a6b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a6888:      	mov	ecx, r10d
  3a688b:      	shr	ecx, 0x10
  3a688e:      	xor	eax, eax
  3a6890:      	cmp	r10d, 0x10000
  3a6897:      	setb	al
  3a689a:      	cmovb	ecx, r10d
  3a689e:      	shl	eax, 0x4
  3a68a1:      	mov	edx, ecx
  3a68a3:      	shr	edx, 0x8
  3a68a6:      	xor	esi, esi
  3a68a8:      	cmp	ecx, 0x100
  3a68ae:      	setb	sil
  3a68b2:      	lea	eax, [rax + 8*rsi]
  3a68b5:      	cmovb	edx, ecx
  3a68b8:      	movzx	ecx, dl
  3a68bb:      	movzx	ecx, byte ptr [rcx + 0x3282d1]
  3a68c2:      	sub	rax, rcx
  3a68c5:      	add	rax, 0x8
  3a68c9:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a68ce:      	mov	ecx, edx
  3a68d0:      	rol	r10d, cl
  3a68d3:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a68d8:      	xor	eax, eax
  3a68da:      	cmp	r10, rdx
  3a68dd:      	setg	al
  3a68e0:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a68e5:      	xor	eax, eax
  3a68e7:      	cmp	r10d, edx
  3a68ea:      	setl	al
  3a68ed:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a68f2:      	xor	eax, eax
  3a68f4:      	cmp	r10, rdx
  3a68f7:      	setne	al
  3a68fa:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a68ff:      	nop
  3a6900:      	test	r10d, r10d
  3a6903:      	je	0x3a6fb7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b7>
  3a6909:      	tzcnt	eax, r10d
  3a690e:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6913:      	add	edx, r10d
  3a6916:      	jmp	0x3a6b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a691b:      	xor	eax, eax
  3a691d:      	cmp	r10d, edx
  3a6920:      	setne	al
  3a6923:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6928:      	xor	eax, eax
  3a692a:      	cmp	r10, rdx
  3a692d:      	setle	al
  3a6930:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6935:      	mov	ecx, edx
  3a6937:      	ror	r10d, cl
  3a693a:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a693f:      	xor	eax, eax
  3a6941:      	cmp	r10d, edx
  3a6944:      	setbe	al
  3a6947:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a694c:      	xor	eax, eax
  3a694e:      	cmp	r10d, edx
  3a6951:      	setge	al
  3a6954:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6959:      	xor	eax, eax
  3a695b:      	test	r10d, r10d
  3a695e:      	sete	al
  3a6961:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6966:      	xor	eax, eax
  3a6968:      	cmp	r10d, edx
  3a696b:      	setb	al
  3a696e:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6973:      	mov	rdx, r10
  3a6976:      	mov	eax, edx
  3a6978:      	shr	eax, 0x18
  3a697b:      	mov	ecx, edx
  3a697d:      	shr	ecx, 0x10
  3a6980:      	movzx	ecx, cl
  3a6983:      	movzx	ecx, byte ptr [rcx + 0x3283d1]
  3a698a:      	add	cl, byte ptr [rax + 0x3283d1]
  3a6990:      	movzx	eax, dh
  3a6993:      	add	cl, byte ptr [rax + 0x3283d1]
  3a6999:      	movzx	eax, dl
  3a699c:      	add	cl, byte ptr [rax + 0x3283d1]
  3a69a2:      	movzx	eax, cl
  3a69a5:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a69aa:      	add	rdx, r10
  3a69ad:      	jmp	0x3a6b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a69b2:      	mov	ecx, edx
  3a69b4:      	shl	r10d, cl
  3a69b7:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a69bc:      	xor	eax, eax
  3a69be:      	test	r10, r10
  3a69c1:      	sete	al
  3a69c4:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a69c9:      	and	rdx, r10
  3a69cc:      	jmp	0x3a6b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a69d1:      	sub	r10, rdx
  3a69d4:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a69d9:      	xor	eax, eax
  3a69db:      	cmp	r10d, edx
  3a69de:      	sete	al
  3a69e1:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a69e6:      	mov	ecx, edx
  3a69e8:      	rol	r10, cl
  3a69eb:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a69f0:      	xor	eax, eax
  3a69f2:      	cmp	r10d, edx
  3a69f5:      	setae	al
  3a69f8:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a69fd:      	and	r10d, edx
  3a6a00:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a6a05:      	xor	eax, eax
  3a6a07:      	cmp	r10d, edx
  3a6a0a:      	setle	al
  3a6a0d:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6a12:      	xor	eax, eax
  3a6a14:      	cmp	r10d, edx
  3a6a17:      	setg	al
  3a6a1a:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6a1f:      	xor	eax, eax
  3a6a21:      	mov	rcx, r10
  3a6a24:      	shr	rcx, 0x20
  3a6a28:      	cmove	rcx, r10
  3a6a2c:      	setne	al
  3a6a2f:      	shl	eax, 0x5
  3a6a32:      	mov	rdx, rcx
  3a6a35:      	shr	rdx, 0x10
  3a6a39:      	lea	rsi, [rax + 0x10]
  3a6a3d:      	cmp	rcx, 0x10000
  3a6a44:      	cmovb	rdx, rcx
  3a6a48:      	cmovb	rsi, rax
  3a6a4c:      	mov	eax, edx
  3a6a4e:      	shr	eax, 0x8
  3a6a51:      	lea	rcx, [rsi + 0x8]
  3a6a55:      	cmp	rdx, 0x100
  3a6a5c:      	cmovb	rax, rdx
  3a6a60:      	cmovb	rcx, rsi
  3a6a64:      	movzx	eax, al
  3a6a67:      	movzx	edx, byte ptr [rax + 0x3282d1]
  3a6a6e:      	add	rdx, rcx
  3a6a71:      	mov	eax, 0x40
  3a6a76:      	sub	rax, rdx
  3a6a79:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6a7e:      	xor	eax, eax
  3a6a80:      	cmp	r10, rdx
  3a6a83:      	setl	al
  3a6a86:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6a8b:      	mov	ecx, edx
  3a6a8d:      	ror	r10, cl
  3a6a90:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a6a95:      	xor	eax, eax
  3a6a97:      	cmp	r10d, edx
  3a6a9a:      	seta	al
  3a6a9d:      	nop	dword ptr [rax]
  3a6aa0:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6aa5:      	mov	eax, r10d
  3a6aa8:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6aad:      	movsx	rax, r10w
  3a6ab1:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6ab6:      	movsxd	rax, r10d
  3a6ab9:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6abe:      	xor	eax, eax
  3a6ac0:      	cmp	r10, rdx
  3a6ac3:      	sete	al
  3a6ac6:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6acb:      	movsxd	rax, r10d
  3a6ace:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6ad3:      	xor	eax, eax
  3a6ad5:      	cmp	r10, rdx
  3a6ad8:      	setae	al
  3a6adb:      	nop	dword ptr [rax + rax]
  3a6ae0:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6ae5:      	xor	eax, eax
  3a6ae7:      	cmp	r10, rdx
  3a6aea:      	setge	al
  3a6aed:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6af2:      	mov	ecx, edx
  3a6af4:      	sar	r10, cl
  3a6af7:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a6af9:      	or	edx, r10d
  3a6afc:      	jmp	0x3a6b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a6afe:      	sub	r10d, edx
  3a6b01:      	jmp	0x3a6b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1530>
  3a6b03:      	imul	rdx, r10
  3a6b07:      	xor	ecx, ecx
  3a6b09:      	mov	rax, rdx
  3a6b0c:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a6b11:      	movsx	rax, r10w
  3a6b15:      	mov	eax, eax
  3a6b17:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6b1c:      	mov	eax, 0x40
  3a6b21:      	tzcnt	rax, r10
  3a6b26:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6b2b:      	mov	ecx, edx
  3a6b2d:      	shl	r10, cl
  3a6b30:      	xor	ecx, ecx
  3a6b32:      	mov	rax, r10
  3a6b35:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a6b3a:      	movsx	rax, r10b
  3a6b3e:      	nop
  3a6b40:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6b45:      	movsx	rax, r10b
  3a6b49:      	mov	eax, eax
  3a6b4b:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6b50:      	movzx	esi, r9b
  3a6b54:      	add	esi, -0x61
  3a6b57:      	nop	word ptr [rax + rax]
  3a6b60:      	cmp	esi, 0x5e
  3a6b63:      	ja	0x3a66f8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10f8>
  3a6b69:      	xor	eax, eax
  3a6b6b:      	xor	ecx, ecx
  3a6b6d:      	jmp	qword ptr [8*rsi + 0x203eb8]
  3a6b74:      	mov	qword ptr [rbp - 0x90], rax
  3a6b7b:      	mov	qword ptr [rbp - 0x88], r8
  3a6b82:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a6b89:      	movups	xmmword ptr [rsp], xmm0
  3a6b8d:      	movzx	r8d, dil
  3a6b91:      	mov	rdi, rsi
  3a6b94:      	mov	esi, edx
  3a6b96:      	mov	rdx, rcx
  3a6b99:      	mov	rcx, r9
  3a6b9c:      	call	0x3ce060 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a6ba1:      	mov	r15d, eax
  3a6ba4:      	xor	eax, eax
  3a6ba6:      	xor	ecx, ecx
  3a6ba8:      	mov	r11, qword ptr [rbp - 0x38]
  3a6bac:      	mov	r8, qword ptr [rbp - 0x30]
  3a6bb0:      	test	r15b, r15b
  3a6bb3:      	jne	0x3a723b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3b>
  3a6bb9:      	mov	edi, dword ptr [r12 + 0x8]
  3a6bbe:      	mov	esi, 0xffffffff
  3a6bc3:      	cmp	rdi, rsi
  3a6bc6:      	mov	r9, qword ptr [rbp - 0x48]
  3a6bca:      	je	0x3a5700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a6bd0:      	cmp	r13, rdi
  3a6bd3:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a6bd9:      	shl	rdi, 0x4
  3a6bdd:      	mov	qword ptr [r9 + rdi], rax
  3a6be1:      	mov	qword ptr [r9 + rdi + 0x8], rcx
  3a6be6:      	jmp	0x3a5700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a6beb:      	cmp	r13, rcx
  3a6bee:      	mov	rdx, qword ptr [rbp - 0x48]
  3a6bf2:      	jbe	0x3a73f2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1df2>
  3a6bf8:      	shl	rcx, 0x4
  3a6bfc:      	mov	rax, qword ptr [rdx + rcx]
  3a6c00:      	mov	rcx, qword ptr [rdx + rcx + 0x8]
  3a6c05:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a6c07:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a6c0b:      	and	cl, 0x1f
  3a6c0e:      	mov	r14, rsi
  3a6c11:      	cmp	cl, 0x16
  3a6c14:      	je	0x3a6c21 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1621>
  3a6c16:      	mov	qword ptr [rbp - 0x90], rdi
  3a6c1d:      	mov	rdi, qword ptr [rbp - 0x68]
  3a6c21:      	call	qword ptr [rax + 0x28]
  3a6c24:      	mov	r10, qword ptr [rbp - 0x30]
  3a6c28:      	mov	r9, qword ptr [rbp - 0x48]
  3a6c2c:      	mov	rsi, r14
  3a6c2f:      	test	rax, rax
  3a6c32:      	je	0x3a6c85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1685>
  3a6c34:      	cmp	qword ptr [rax + 0x8], 0x0
  3a6c39:      	mov	r11, qword ptr [rbp - 0x38]
  3a6c3d:      	je	0x3a6c89 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1689>
  3a6c3f:      	nop
  3a6c40:      	jmp	0x3a71c6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc6>
  3a6c45:      	mov	r9, qword ptr [rbp - 0x48]
  3a6c49:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a6c50:      	add	edi, dword ptr [r12 + 0x18]
  3a6c55:      	cmp	rsi, rdi
  3a6c58:      	ja	0x3a648d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe8d>
  3a6c5e:      	nop
  3a6c60:      	jmp	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a6c65:      	mov	edi, 0x1
  3a6c6a:      	mov	r14, rsi
  3a6c6d:      	call	0x61eac0 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a6c72:      	mov	rsi, r14
  3a6c75:      	mov	r9, qword ptr [rbp - 0x48]
  3a6c79:      	mov	r10, qword ptr [rbp - 0x30]
  3a6c7d:      	nop	dword ptr [rax]
  3a6c80:      	test	rax, rax
  3a6c83:      	jne	0x3a6c34 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1634>
  3a6c85:      	mov	r11, qword ptr [rbp - 0x38]
  3a6c89:      	mov	rax, qword ptr [r10 + 0x108]
  3a6c90:      	lea	edx, [rsi + r12]
  3a6c94:      	cmp	rax, rdx
  3a6c97:      	jb	0x3a73d5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dd5>
  3a6c9d:      	nop	dword ptr [rax]
  3a6ca0:      	cmp	r12d, edx
  3a6ca3:      	ja	0x3a73c6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc6>
  3a6ca9:      	sub	rax, r12
  3a6cac:      	lea	rcx, [8*r12]
  3a6cb4:      	test	rax, rax
  3a6cb7:      	mov	eax, 0x0
  3a6cbc:      	cmovle	rcx, rax
  3a6cc0:      	test	esi, esi
  3a6cc2:      	je	0x3a6d33 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1733>
  3a6cc4:      	mov	rax, qword ptr [r10 + 0xf8]
  3a6ccb:      	sub	rdx, r12
  3a6cce:      	add	rcx, rax
  3a6cd1:      	add	rcx, 0x4
  3a6cd5:      	xor	esi, esi
  3a6cd7:      	nop	word ptr cs:[rax + rax]
  3a6ce6:      	nop	word ptr cs:[rax + rax]
  3a6cf5:      	nop	word ptr cs:[rax + rax]
  3a6d00:      	mov	edi, dword ptr [rcx + 8*rsi]
  3a6d03:      	cmp	r13, rdi
  3a6d06:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a6d0c:      	mov	eax, dword ptr [rcx + 8*rsi - 0x4]
  3a6d10:      	cmp	r13, rax
  3a6d13:      	jbe	0x3a72fc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cfc>
  3a6d19:      	shl	rdi, 0x4
  3a6d1d:      	shl	rax, 0x4
  3a6d21:      	movups	xmm0, xmmword ptr [r9 + rdi]
  3a6d26:      	movups	xmmword ptr [r9 + rax], xmm0
  3a6d2b:      	inc	rsi
  3a6d2e:      	cmp	rdx, rsi
  3a6d31:      	jne	0x3a6d00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1700>
  3a6d33:      	mov	rcx, qword ptr [rbp - 0x58]
  3a6d37:      	mov	esi, 0xffffffff
  3a6d3c:      	nop	dword ptr [rax]
  3a6d40:      	jmp	0x3a570d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10d>
  3a6d45:      	cmp	eax, 0x3
  3a6d48:      	je	0x3a7147 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6d4e:      	cmp	eax, 0x6
  3a6d51:      	je	0x3a70ea <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aea>
  3a6d57:      	lea	eax, [r9 - 0x12]
  3a6d5b:      	nop	dword ptr [rax + rax]
  3a6d60:      	cmp	eax, 0x4
  3a6d63:      	jb	0x3a711d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b1d>
  3a6d69:      	mov	edi, dword ptr [r12 + 0x14]
  3a6d6e:      	cmp	rdi, rsi
  3a6d71:      	je	0x3a6dd9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17d9>
  3a6d73:      	cmp	r13, rdi
  3a6d76:      	jbe	0x3a72e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a6d7c:      	shl	rdi, 0x4
  3a6d80:      	mov	rax, qword ptr [rbp - 0x48]
  3a6d84:      	mov	r15, qword ptr [rax + rdi]
  3a6d88:      	mov	eax, r9d
  3a6d8b:      	and	eax, 0xffff0000
  3a6d90:      	cmp	eax, 0xfe0000
  3a6d95:      	jne	0x3a6deb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17eb>
  3a6d97:      	movzx	r9d, r9w
  3a6d9b:      	nop	dword ptr [rax + rax]
  3a6da0:      	cmp	r9d, 0x3
  3a6da4:      	jne	0x3a6fc5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c5>
  3a6daa:      	test	r11, r11
  3a6dad:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6db3:      	mov	rax, qword ptr [r11]
  3a6db6:      	test	rax, rax
  3a6db9:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6dbf:      	mov	rdi, qword ptr [rax + 0x8]
  3a6dc3:      	call	0x3c95c0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a6dc8:      	mov	r11, qword ptr [rbp - 0x38]
  3a6dcc:      	xor	eax, eax
  3a6dce:      	xor	ecx, ecx
  3a6dd0:      	mov	r8, qword ptr [rbp - 0x30]
  3a6dd4:      	jmp	0x3a6bb9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b9>
  3a6dd9:      	xor	r15d, r15d
  3a6ddc:      	mov	eax, r9d
  3a6ddf:      	and	eax, 0xffff0000
  3a6de4:      	cmp	eax, 0xfe0000
  3a6de9:      	je	0x3a6d97 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1797>
  3a6deb:      	add	r9d, 0xff03fff8
  3a6df2:      	cmp	r9d, 0xa
  3a6df6:      	jae	0x3a7424 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e24>
  3a6dfc:      	mov	rcx, qword ptr [rbp - 0x30]
  3a6e00:      	mov	rsi, qword ptr [rcx + 0x80]
  3a6e07:      	mov	edi, dword ptr [r12 + 0x4]
  3a6e0c:      	cmp	rsi, rdi
  3a6e0f:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a6e15:      	imul	rax, rdi, 0xe0
  3a6e1c:      	nop	dword ptr [rax]
  3a6e20:      	add	rax, qword ptr [rcx + 0x78]
  3a6e24:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6e2a:      	movups	xmm0, xmmword ptr [rax + 0x98]
  3a6e31:      	movups	xmm1, xmmword ptr [rax + 0x88]
  3a6e38:      	movups	xmm2, xmmword ptr [rax + 0x78]
  3a6e3c:      	movups	xmm3, xmmword ptr [rax + 0x68]
  3a6e40:      	movups	xmm4, xmmword ptr [rax + 0x58]
  3a6e44:      	movups	xmm5, xmmword ptr [rax + 0x48]
  3a6e48:      	movzx	ecx, byte ptr [rax + 0x40]
  3a6e4c:      	movups	xmm6, xmmword ptr [rax + 0x30]
  3a6e50:      	movups	xmm7, xmmword ptr [rax + 0x20]
  3a6e54:      	movups	xmm8, xmmword ptr [rax + 0x10]
  3a6e59:      	mov	rsi, qword ptr [rax + 0x8]
  3a6e5d:      	mov	eax, dword ptr [rax]
  3a6e5f:      	mov	dword ptr [rbp - 0x208], eax
  3a6e65:      	mov	qword ptr [rbp - 0x200], rsi
  3a6e6c:      	movups	xmmword ptr [rbp - 0x1f8], xmm8
  3a6e74:      	movups	xmmword ptr [rbp - 0x1e8], xmm7
  3a6e7b:      	movups	xmmword ptr [rbp - 0x1d8], xmm6
  3a6e82:      	mov	byte ptr [rbp - 0x1c8], cl
  3a6e88:      	movups	xmmword ptr [rbp - 0x1c0], xmm5
  3a6e8f:      	movups	xmmword ptr [rbp - 0x1b0], xmm4
  3a6e96:      	movups	xmmword ptr [rbp - 0x1a0], xmm3
  3a6e9d:      	movups	xmmword ptr [rbp - 0x190], xmm2
  3a6ea4:      	movups	xmmword ptr [rbp - 0x180], xmm1
  3a6eab:      	movups	xmmword ptr [rbp - 0x170], xmm0
  3a6eb2:      	mov	qword ptr [rbp - 0xf0], rdx
  3a6eb9:      	mov	qword ptr [rbp - 0xe8], 0x0
  3a6ec4:      	mov	qword ptr [rbp - 0xe0], r15
  3a6ecb:      	mov	qword ptr [rbp - 0xd8], 0x0
  3a6ed6:      	mov	ecx, 0x15
  3a6edb:      	mov	rdi, rsp
  3a6ede:      	lea	rsi, [rbp - 0x208]
  3a6ee5:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a6ee8:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a6eef:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a6ef7:      	movups	xmm0, xmmword ptr [rbp - 0xf0]
  3a6efe:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a6f06:      	lea	rdi, [rbp - 0x118]
  3a6f0d:      	mov	rsi, r11
  3a6f10:      	mov	rcx, r10
  3a6f13:      	mov	rdx, qword ptr [rbp - 0x50]
  3a6f17:      	call	0x3ac7c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a6f1c:      	mov	rax, qword ptr [rbp - 0x118]
  3a6f23:      	mov	rcx, qword ptr [rbp - 0x110]
  3a6f2a:      	movzx	r15d, byte ptr [rbp - 0x108]
  3a6f32:      	mov	rsi, qword ptr [rbp - 0x100]
  3a6f39:      	nop	dword ptr [rax]
  3a6f40:      	test	rsi, rsi
  3a6f43:      	je	0x3a70ab <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aab>
  3a6f49:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a6f4e:      	mov	r11, qword ptr [rbp - 0x38]
  3a6f52:      	mov	r8, qword ptr [rbp - 0x30]
  3a6f56:      	je	0x3a6bb0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b0>
  3a6f5c:      	nop	dword ptr [rax]
  3a6f60:      	jmp	0x3a72c4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cc4>
  3a6f65:      	test	rax, rax
  3a6f68:      	je	0x3a70b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ab9>
  3a6f6e:      	mov	rax, qword ptr [rax]
  3a6f71:      	test	rdi, rdi
  3a6f74:      	je	0x3a70c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ac9>
  3a6f7a:      	mov	rcx, qword ptr [rdi + 0x38]
  3a6f7e:      	test	rcx, rcx
  3a6f81:      	mov	edx, 0x10000
  3a6f86:      	cmove	rcx, rdx
  3a6f8a:      	mov	rdx, rax
  3a6f8d:      	or	rdx, rcx
  3a6f90:      	shr	rdx, 0x20
  3a6f94:      	je	0x3a70b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ab0>
  3a6f9a:      	xor	edx, edx
  3a6f9c:      	div	rcx
  3a6f9f:      	nop
  3a6fa0:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6fa5:      	xor	eax, eax
  3a6fa7:      	xor	edx, edx
  3a6fa9:      	test	rax, rax
  3a6fac:      	jne	0x3a6c34 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1634>
  3a6fb2:      	jmp	0x3a6c85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1685>
  3a6fb7:      	mov	eax, 0x20
  3a6fbc:      	nop	dword ptr [rax]
  3a6fc0:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a6fc5:      	mov	rax, qword ptr [rbp - 0x30]
  3a6fc9:      	mov	rsi, qword ptr [rax + 0x80]
  3a6fd0:      	mov	edi, dword ptr [r12 + 0x4]
  3a6fd5:      	cmp	rsi, rdi
  3a6fd8:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a6fde:      	imul	r11, rdi, 0xe0
  3a6fe5:      	add	r11, qword ptr [rax + 0x78]
  3a6fe9:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a6fef:      	mov	rax, qword ptr [rbp - 0x50]
  3a6ff3:      	mov	rsi, qword ptr [rax + 0x78]
  3a6ff7:      	mov	edi, dword ptr [r11 + 0x18]
  3a6ffb:      	nop	dword ptr [rax + rax]
  3a7000:      	cmp	rsi, rdi
  3a7003:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a7009:      	shl	rdi, 0x3
  3a700d:      	add	rdi, qword ptr [rax + 0x70]
  3a7011:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a7017:      	mov	rsi, qword ptr [rdi]
  3a701a:      	nop	word ptr [rax + rax]
  3a7020:      	test	rsi, rsi
  3a7023:      	je	0x3a7165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a7029:      	mov	rdi, qword ptr [rbp - 0x38]
  3a702d:      	mov	rcx, qword ptr [rdi]
  3a7030:      	mov	rax, qword ptr [rdi + 0x8]
  3a7034:      	mov	r8, qword ptr [rcx + 0x8]
  3a7038:      	mov	rcx, qword ptr [rdi + 0x10]
  3a703c:      	mov	rsi, qword ptr [rsi + 0x10]
  3a7040:      	mov	edi, dword ptr [r11 + 0x14]
  3a7044:      	mov	r11d, dword ptr [r11 + 0x1c]
  3a7048:      	shl	r11, 0x20
  3a704c:      	or	r11, rdi
  3a704f:      	mov	qword ptr [rsp + 0x18], r15
  3a7054:      	mov	qword ptr [rsp + 0x10], rdx
  3a7059:      	mov	qword ptr [rsp + 0x8], r11
  3a705e:      	mov	qword ptr [rsp], r10
  3a7062:      	lea	rdi, [rbp - 0x140]
  3a7069:      	mov	rdx, rax
  3a706c:      	call	0x3caa80 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a7071:      	mov	rax, qword ptr [rbp - 0x140]
  3a7078:      	mov	rcx, qword ptr [rbp - 0x138]
  3a707f:      	movzx	r15d, byte ptr [rbp - 0x130]
  3a7087:      	mov	rsi, qword ptr [rbp - 0x128]
  3a708e:      	test	rsi, rsi
  3a7091:      	je	0x3a70e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ae5>
  3a7093:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a7098:      	mov	r11, qword ptr [rbp - 0x38]
  3a709c:      	mov	r8, qword ptr [rbp - 0x30]
  3a70a0:      	je	0x3a6bb0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15b0>
  3a70a6:      	jmp	0x3a72cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ccd>
  3a70ab:      	jmp	0x3a6ba8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15a8>
  3a70b0:      	xor	edx, edx
  3a70b2:      	div	ecx
  3a70b4:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a70b9:      	mov	rax, qword ptr [rdi + 0x8]
  3a70bd:      	nop	dword ptr [rax]
  3a70c0:      	test	rdi, rdi
  3a70c3:      	jne	0x3a6f7a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x197a>
  3a70c9:      	mov	edi, 0x1
  3a70ce:      	call	0x61eac0 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a70d3:      	mov	r8, qword ptr [rbp - 0x30]
  3a70d7:      	mov	r11, qword ptr [rbp - 0x38]
  3a70db:      	nop	dword ptr [rax + rax]
  3a70e0:      	jmp	0x3a6700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1100>
  3a70e5:      	jmp	0x3a6ba8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15a8>
  3a70ea:      	mov	rax, qword ptr [rbp - 0x60]
  3a70ee:      	mov	dword ptr [rax - 0x38], r14d
  3a70f2:      	mov	rdi, r11
  3a70f5:      	mov	r14, r11
  3a70f8:      	call	0x3a9a20 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a70fd:      	nop	dword ptr [rax]
  3a7100:      	test	al, al
  3a7102:      	jne	0x3a72af <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1caf>
  3a7108:      	mov	rcx, qword ptr [r14 + 0x48]
  3a710c:      	test	rcx, rcx
  3a710f:      	mov	r11, r14
  3a7112:      	jg	0x3a5645 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x45>
  3a7118:      	jmp	0x3a7415 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e15>
  3a711d:      	mov	rax, qword ptr [rbp - 0x60]
  3a7121:      	mov	dword ptr [rax - 0x38], r14d
  3a7125:      	mov	rdi, r11
  3a7128:      	mov	r14, r11
  3a712b:      	call	0x3a8700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a7130:      	jmp	0x3a70fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1afd>
  3a7132:      	mov	rax, qword ptr [rbp - 0x60]
  3a7136:      	mov	dword ptr [rax - 0x38], r14d
  3a713a:      	mov	rdi, r11
  3a713d:      	mov	r14, r11
  3a7140:      	call	0x3a9500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a7145:      	jmp	0x3a70fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1afd>
  3a7147:      	mov	rax, qword ptr [rbp - 0x60]
  3a714b:      	mov	dword ptr [rax - 0x38], r14d
  3a714f:      	mov	rdi, r11
  3a7152:      	mov	r14, r11
  3a7155:      	call	0x3a8200 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a715a:      	test	al, al
  3a715c:      	je	0x3a7108 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b08>
  3a715e:      	mov	al, 0x1
  3a7160:      	jmp	0x3a72af <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1caf>
  3a7165:      	mov	edi, 0x1
  3a716a:      	call	0x61eac0 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a716f:      	jmp	0x3a72af <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1caf>
  3a7174:      	mov	edi, 0x28
  3a7179:      	call	0x638250 <GC_malloc@plt>
  3a717e:      	nop
  3a7180:      	test	rax, rax
  3a7183:      	je	0x3a7433 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e33>
  3a7189:      	mov	r14, rax
  3a718c:      	add	r15, r13
  3a718f:      	mov	edi, 0x28
  3a7194:      	call	0x619d80 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a7199:      	mov	rax, qword ptr [r15 + 0x8]
  3a719d:      	movaps	xmm0, xmmword ptr [rip + 0x34ea7c] # 0x6f5c20 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a71a4:      	movups	xmmword ptr [r14], xmm0
  3a71a8:      	mov	qword ptr [r14 + 0x10], rax
  3a71ac:      	mov	qword ptr [r14 + 0x18], 0x299e78
  3a71b4:      	mov	qword ptr [r14 + 0x20], 0xe
  3a71bc:      	mov	esi, 0x2ba878
  3a71c1:      	jmp	0x3a72a1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ca1>
  3a71c6:      	mov	rdi, r11
  3a71c9:      	mov	rsi, rax
  3a71cc:      	jmp	0x3a72a8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ca8>
  3a71d1:      	mov	r15d, edx
  3a71d4:      	jmp	0x3a723b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3b>
  3a71d6:      	mov	r14, qword ptr [rbp - 0x148]
  3a71dd:      	mov	edi, 0x8
  3a71e2:      	call	0x638250 <GC_malloc@plt>
  3a71e7:      	test	rax, rax
  3a71ea:      	je	0x3a7433 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e33>
  3a71f0:      	mov	r12, rax
  3a71f3:      	mov	edi, 0x8
  3a71f8:      	call	0x619d80 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a71fd:      	mov	edx, 0x299de8
  3a7202:      	mov	rdi, r15
  3a7205:      	mov	rsi, r14
  3a7208:      	mov	rcx, r12
  3a720b:      	call	0x5eb5c0 <errors.As>
  3a7210:      	test	al, 0x1
  3a7212:      	je	0x3a7225 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c25>
  3a7214:      	mov	rax, qword ptr [r12]
  3a7218:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a721f:      	mov	ecx, dword ptr [rcx + 0x50]
  3a7222:      	mov	dword ptr [rax + 0x4], ecx
  3a7225:      	mov	rdi, qword ptr [rbp - 0x38]
  3a7229:      	mov	rsi, r15
  3a722c:      	jmp	0x3a72a5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ca5>
  3a722e:      	mov	r15b, 0x3
  3a7231:      	jmp	0x3a723b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3b>
  3a7233:      	mov	r15b, 0x1
  3a7236:      	jmp	0x3a723b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3b>
  3a7238:      	mov	r15b, 0x9
  3a723b:      	mov	edi, 0x10
  3a7240:      	call	0x6381e0 <GC_malloc_atomic@plt>
  3a7245:      	test	rax, rax
  3a7248:      	je	0x3a7433 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e33>
  3a724e:      	mov	r14, rax
  3a7251:      	mov	edi, 0x10
  3a7256:      	call	0x619d80 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a725b:      	xorps	xmm0, xmm0
  3a725e:      	movups	xmmword ptr [r14], xmm0
  3a7262:      	mov	rax, qword ptr [rbp - 0x30]
  3a7266:      	mov	rsi, qword ptr [rax + 0x80]
  3a726d:      	mov	edi, dword ptr [r12 + 0x4]
  3a7272:      	cmp	rsi, rdi
  3a7275:      	jbe	0x3a72ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cec>
  3a7277:      	mov	rax, qword ptr [rax + 0x78]
  3a727b:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a7282:      	mov	ecx, dword ptr [rcx + 0x50]
  3a7285:      	imul	rdx, rdi, 0xe0
  3a728c:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a7291:      	mov	byte ptr [r14], r15b
  3a7294:      	mov	dword ptr [r14 + 0x4], ecx
  3a7298:      	mov	qword ptr [r14 + 0x8], rax
  3a729c:      	mov	esi, 0x29a200
  3a72a1:      	mov	rdi, qword ptr [rbp - 0x38]
  3a72a5:      	mov	rdx, r14
  3a72a8:      	call	0x3a4f20 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a72ad:      	mov	al, 0x3
  3a72af:      	mov	rdx, rbx
  3a72b2:      	add	rsp, 0x2a8
  3a72b9:      	pop	rbx
  3a72ba:      	pop	r12
  3a72bc:      	pop	r13
  3a72be:      	pop	r14
  3a72c0:      	pop	r15
  3a72c2:      	pop	rbp
  3a72c3:      	ret
  3a72c4:      	mov	rdx, qword ptr [rbp - 0xf8]
  3a72cb:      	jmp	0x3a72d4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cd4>
  3a72cd:      	mov	rdx, qword ptr [rbp - 0x120]
  3a72d4:      	mov	rdi, r11
  3a72d7:      	jmp	0x3a72a8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ca8>
  3a72d9:      	mov	rdi, r14
  3a72dc:      	mov	rsi, r15
  3a72df:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a72e4:      	mov	rsi, r13
  3a72e7:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a72ec:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a72f1:      	mov	rdi, r12
  3a72f4:      	mov	rsi, r15
  3a72f7:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a72fc:      	mov	rdi, rax
  3a72ff:      	mov	rsi, r13
  3a7302:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a7307:      	mov	rdi, rcx
  3a730a:      	mov	rsi, r13
  3a730d:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a7312:      	mov	rsi, r13
  3a7315:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a731a:      	mov	rdi, r9
  3a731d:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7322:      	mov	rdi, rax
  3a7325:      	mov	rsi, r13
  3a7328:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a732d:      	mov	rdi, rax
  3a7330:      	mov	esi, 0x1
  3a7335:      	mov	ecx, 0x2
  3a733a:      	call	0x61ae40 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a733f:      	mov	rdi, r9
  3a7342:      	mov	esi, 0x1
  3a7347:      	mov	rdx, rax
  3a734a:      	mov	ecx, 0x3
  3a734f:      	call	0x61ae40 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a7354:      	mov	rsi, rcx
  3a7357:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a735c:      	mov	rdi, rsi
  3a735f:      	mov	rsi, rcx
  3a7362:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7367:      	mov	rdi, rcx
  3a736a:      	mov	rsi, r13
  3a736d:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7372:      	mov	edi, 0x1
  3a7377:      	mov	esi, 0x1
  3a737c:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a7381:      	xor	edi, edi
  3a7383:      	xor	esi, esi
  3a7385:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a738a:      	xor	esi, esi
  3a738c:      	mov	rdx, rcx
  3a738f:      	mov	ecx, 0x3
  3a7394:      	call	0x61ae40 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a7399:      	mov	rdi, rcx
  3a739c:      	xor	esi, esi
  3a739e:      	mov	ecx, 0x2
  3a73a3:      	call	0x61ae40 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a73a8:      	mov	edi, 0x3
  3a73ad:      	mov	esi, 0x3
  3a73b2:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a73b7:      	mov	edi, 0x2
  3a73bc:      	mov	esi, 0x2
  3a73c1:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a73c6:      	mov	rdi, r12
  3a73c9:      	xor	esi, esi
  3a73cb:      	mov	ecx, 0x3
  3a73d0:      	call	0x61ae40 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a73d5:      	mov	rdi, rdx
  3a73d8:      	xor	esi, esi
  3a73da:      	mov	rdx, rax
  3a73dd:      	mov	ecx, 0x2
  3a73e2:      	call	0x61ae40 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a73e7:      	mov	rdi, rdx
  3a73ea:      	mov	rsi, r13
  3a73ed:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a73f2:      	mov	rdi, rcx
  3a73f5:      	mov	rsi, qword ptr [rbp - 0x78]
  3a73f9:      	call	0x619e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a73fe:      	mov	rdi, rax
  3a7401:      	call	0x61b6e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7406:      	mov	edi, 0x3293b0
  3a740b:      	mov	esi, 0x3296e0
  3a7410:      	call	0x619a80 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a7415:      	mov	edi, 0x333868
  3a741a:      	mov	esi, 0x299e68
  3a741f:      	call	0x619a80 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a7424:      	mov	edi, 0x333868
  3a7429:      	mov	esi, 0x299e88
  3a742e:      	call	0x619a80 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a7433:      	mov	edi, 0x333868
  3a7438:      	mov	esi, 0x329388
  3a743d:      	call	0x619a80 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a7442:      	int3
  3a7443:      	int3
  3a7444:      	int3
  3a7445:      	int3
  3a7446:      	int3
  3a7447:      	int3
  3a7448:      	int3
  3a7449:      	int3
  3a744a:      	int3
  3a744b:      	int3
  3a744c:      	int3
  3a744d:      	int3
  3a744e:      	int3
  3a744f:      	int3
  3a7450:      	int3
  3a7451:      	int3
  3a7452:      	int3
  3a7453:      	int3
  3a7454:      	int3
  3a7455:      	int3
  3a7456:      	int3
  3a7457:      	int3
  3a7458:      	int3
  3a7459:      	int3
  3a745a:      	int3
  3a745b:      	int3
  3a745c:      	int3
  3a745d:      	int3
  3a745e:      	int3
  3a745f:      	int3
