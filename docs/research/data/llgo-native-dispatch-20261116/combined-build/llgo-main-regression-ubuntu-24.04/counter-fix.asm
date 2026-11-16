
/home/runner/work/wasmvm/wasmvm/main-regression/counter-fix:	file format elf64-x86-64

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
  3a5623:      	jle	0x3a71b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bb9>
  3a5629:      	mov	rbx, rsi
  3a562c:      	mov	r11, rdi
  3a562f:      	movzx	r12d, byte ptr [rdi + 0x20]
  3a5634:      	mov	r8d, 0xffffffff
  3a563a:      	lea	r14, [rbp - 0x90]
  3a5641:      	mov	qword ptr [rbp - 0x40], rdi
  3a5645:      	mov	byte ptr [rbp - 0x29], r12b
  3a5649:      	mov	r10, qword ptr [r11 + 0x40]
  3a564d:      	shl	rcx, 0x6
  3a5651:      	mov	byte ptr [r10 + rcx - 0x30], 0x0
  3a5657:      	mov	rsi, qword ptr [r10 + rcx - 0x40]
  3a565c:      	mov	rax, qword ptr [rsi + 0x48]
  3a5660:      	mov	qword ptr [rbp - 0x58], rax
  3a5664:      	mov	r9, qword ptr [rax + 0x98]
  3a566b:      	mov	rdi, qword ptr [r10 + rcx - 0x28]
  3a5670:      	mov	rdx, qword ptr [r11 + 0x88]
  3a5677:      	lea	rax, [r9 + rdi]
  3a567b:      	nop	dword ptr [rax + rax]
  3a5680:      	cmp	rax, rdx
  3a5683:      	ja	0x3a737f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d7f>
  3a5689:      	mov	qword ptr [rbp - 0x68], r14
  3a568d:      	cmp	rdi, rax
  3a5690:      	ja	0x3a7391 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d91>
  3a5696:      	mov	r14, qword ptr [rbp - 0x58]
  3a569a:      	mov	r13, qword ptr [r14 + 0xb8]
  3a56a1:      	mov	r15d, dword ptr [r10 + rcx - 0x38]
  3a56a6:      	sub	rdx, rdi
  3a56a9:      	shl	rdi, 0x4
  3a56ad:      	test	rdx, rdx
  3a56b0:      	mov	eax, 0x0
  3a56b5:      	cmovle	rdi, rax
  3a56b9:      	nop	dword ptr [rax]
  3a56c0:      	cmp	r13, r15
  3a56c3:      	jbe	0x3a734f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4f>
  3a56c9:      	add	r10, rcx
  3a56cc:      	mov	qword ptr [rbp - 0x70], r10
  3a56d0:      	mov	qword ptr [rbp - 0xa8], rsi
  3a56d7:      	mov	rax, qword ptr [rsi + 0x40]
  3a56db:      	mov	qword ptr [rbp - 0x60], rax
  3a56df:      	mov	rsi, qword ptr [r14 + 0xb0]
  3a56e6:      	add	rdi, qword ptr [r11 + 0x78]
  3a56ea:      	mov	r14d, r15d
  3a56ed:      	mov	qword ptr [rbp - 0x38], r9
  3a56f1:      	mov	r10, rdi
  3a56f4:      	mov	qword ptr [rbp - 0x50], rdi
  3a56f8:      	mov	qword ptr [rbp - 0xb0], r13
  3a56ff:      	mov	qword ptr [rbp - 0x48], rsi
  3a5703:      	jmp	0x3a5769 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x169>
  3a5705:      	nop	word ptr cs:[rax + rax]
  3a5714:      	nop	word ptr cs:[rax + rax]
  3a5723:      	nop	word ptr cs:[rax + rax]
  3a5732:      	nop	word ptr cs:[rax + rax]
  3a5740:      	inc	r14d
  3a5743:      	mov	r13d, r14d
  3a5746:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a574b:      	mov	r15d, r13d
  3a574e:      	mov	r14d, r13d
  3a5751:      	mov	r13, qword ptr [rbp - 0xb0]
  3a5758:      	nop	dword ptr [rax + rax]
  3a5760:      	cmp	r13, r15
  3a5763:      	jbe	0x3a734f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4f>
  3a5769:      	shl	r15, 0x5
  3a576d:      	add	r15, rsi
  3a5770:      	test	r12b, 0x1
  3a5774:      	je	0x3a57d6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d6>
  3a5776:      	mov	rax, qword ptr [rbp - 0x58]
  3a577a:      	mov	rsi, qword ptr [rax + 0x80]
  3a5781:      	mov	edi, dword ptr [r15 + 0x4]
  3a5785:      	cmp	rsi, rdi
  3a5788:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a578e:      	mov	r13, qword ptr [rax + 0x78]
  3a5792:      	imul	r12, rdi, 0xe0
  3a5799:      	mov	eax, dword ptr [r13 + r12]
  3a579e:      	cmp	eax, 0x5
  3a57a1:      	setne	cl
  3a57a4:      	cmp	eax, 0xb
  3a57a7:      	setne	al
  3a57aa:      	and	al, cl
  3a57ac:      	movzx	ecx, al
  3a57af:      	mov	rax, qword ptr [r11 + 0x150]
  3a57b6:      	sub	rax, rcx
  3a57b9:      	jb	0x3a71d7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd7>
  3a57bf:      	mov	qword ptr [r11 + 0x150], rax
  3a57c6:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a57cb:      	mov	r13, qword ptr [rbp - 0xb0]
  3a57d2:      	mov	rsi, qword ptr [rbp - 0x48]
  3a57d6:      	inc	rbx
  3a57d9:      	nop	dword ptr [rax]
  3a57e0:      	test	bl, 0x3f
  3a57e3:      	jne	0x3a5867 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x267>
  3a57e9:      	mov	rax, qword ptr [r11 + 0x160]
  3a57f0:      	test	rax, rax
  3a57f3:      	je	0x3a5808 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x208>
  3a57f5:      	test	r11, r11
  3a57f8:      	je	0x3a592a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x32a>
  3a57fe:      	mov	eax, dword ptr [rax]
  3a5800:      	test	eax, eax
  3a5802:      	je	0x3a5e6b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x86b>
  3a5808:      	mov	rax, qword ptr [r11 + 0x8]
  3a580c:      	test	rax, rax
  3a580f:      	je	0x3a744d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e4d>
  3a5815:      	mov	rdi, qword ptr [r11 + 0x10]
  3a5819:      	mov	rcx, qword ptr [rax + 0x8]
  3a581d:      	test	byte ptr [rcx + 0x17], 0x20
  3a5821:      	je	0x3a583a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x23a>
  3a5823:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a5827:      	and	cl, 0x1f
  3a582a:      	cmp	cl, 0x16
  3a582d:      	je	0x3a583a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x23a>
  3a582f:      	mov	qword ptr [rbp - 0x90], rdi
  3a5836:      	mov	rdi, qword ptr [rbp - 0x68]
  3a583a:      	call	qword ptr [rax + 0x28]
  3a583d:      	mov	r8d, 0xffffffff
  3a5843:      	mov	r9, qword ptr [rbp - 0x38]
  3a5847:      	mov	r10, qword ptr [rbp - 0x50]
  3a584b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a584f:      	test	rax, rax
  3a5852:      	je	0x3a594f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x34f>
  3a5858:      	cmp	qword ptr [rax + 0x8], 0x0
  3a585d:      	mov	r11, qword ptr [rbp - 0x40]
  3a5861:      	jne	0x3a7227 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c27>
  3a5867:      	cmp	dword ptr [r15 + 0x1c], 0x0
  3a586c:      	je	0x3a5980 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x380>
  3a5872:      	test	r12b, 0x1
  3a5876:      	je	0x3a58c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a5878:      	mov	edx, dword ptr [r15]
  3a587b:      	nop	dword ptr [rax + rax]
  3a5880:      	cmp	edx, 0x20
  3a5883:      	je	0x3a5980 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x380>
  3a5889:      	cmp	edx, 0x41
  3a588c:      	je	0x3a5980 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x380>
  3a5892:      	jmp	0x3a5983 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x383>
  3a5897:      	nop	word ptr cs:[rax + rax]
  3a58a6:      	nop	word ptr cs:[rax + rax]
  3a58b5:      	nop	word ptr cs:[rax + rax]
  3a58c0:      	mov	edx, dword ptr [r15]
  3a58c3:      	cmp	edx, 0x20
  3a58c6:      	je	0x3a58d1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d1>
  3a58c8:      	cmp	edx, 0x41
  3a58cb:      	jne	0x3a5983 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x383>
  3a58d1:      	mov	eax, ebx
  3a58d3:      	not	eax
  3a58d5:      	test	al, 0x3f
  3a58d7:      	je	0x3a5980 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x380>
  3a58dd:      	nop	dword ptr [rax]
  3a58e0:      	cmp	edx, 0x41
  3a58e3:      	jne	0x3a5907 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x307>
  3a58e5:      	mov	edi, dword ptr [r15 + 0x8]
  3a58e9:      	cmp	r9, rdi
  3a58ec:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a58f2:      	mov	eax, dword ptr [r15 + 0x18]
  3a58f6:      	shl	rdi, 0x4
  3a58fa:      	mov	qword ptr [r10 + rdi], rax
  3a58fe:      	mov	qword ptr [r10 + rdi + 0x8], 0x0
  3a5907:      	inc	r14d
  3a590a:      	cmp	r13, r14
  3a590d:      	jbe	0x3a7344 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d44>
  3a5913:      	mov	rax, r14
  3a5916:      	shl	rax, 0x5
  3a591a:      	lea	r15, [rsi + rax]
  3a591e:      	inc	rbx
  3a5921:      	cmp	dword ptr [rsi + rax + 0x1c], 0x0
  3a5926:      	jne	0x3a58c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a5928:      	jmp	0x3a5980 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x380>
  3a592a:      	mov	edi, 0x1
  3a592f:      	call	0x61eb00 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a5934:      	mov	rsi, qword ptr [rbp - 0x48]
  3a5938:      	mov	r10, qword ptr [rbp - 0x50]
  3a593c:      	mov	r9, qword ptr [rbp - 0x38]
  3a5940:      	mov	r8d, 0xffffffff
  3a5946:      	test	rax, rax
  3a5949:      	jne	0x3a5858 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x258>
  3a594f:      	mov	r11, qword ptr [rbp - 0x40]
  3a5953:      	cmp	dword ptr [r15 + 0x1c], 0x0
  3a5958:      	jne	0x3a5872 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x272>
  3a595e:      	nop	word ptr cs:[rax + rax]
  3a596d:      	nop	word ptr cs:[rax + rax]
  3a597c:      	nop	dword ptr [rax]
  3a5980:      	mov	edx, dword ptr [r15]
  3a5983:      	lea	eax, [rdx - 0xe0]
  3a5989:      	cmp	eax, 0xffff20
  3a598e:      	jae	0x3a5c40 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x640>
  3a5994:      	mov	eax, edx
  3a5996:      	shr	eax, 0x10
  3a5999:      	nop	dword ptr [rax]
  3a59a0:      	cmp	eax, 0xfb
  3a59a5:      	je	0x3a5ba5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5a5>
  3a59ab:      	cmp	eax, 0xfc
  3a59b0:      	je	0x3a5a65 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x465>
  3a59b6:      	nop	word ptr cs:[rax + rax]
  3a59c0:      	cmp	eax, 0xfd
  3a59c5:      	jne	0x3a5c31 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x631>
  3a59cb:      	mov	edi, dword ptr [r15 + 0xc]
  3a59cf:      	cmp	rdi, r8
  3a59d2:      	je	0x3a5f25 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x925>
  3a59d8:      	cmp	r9, rdi
  3a59db:      	mov	r12, qword ptr [rbp - 0x58]
  3a59df:      	nop
  3a59e0:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a59e6:      	shl	rdi, 0x4
  3a59ea:      	mov	rcx, qword ptr [r10 + rdi]
  3a59ee:      	mov	r10, qword ptr [r10 + rdi + 0x8]
  3a59f3:      	mov	edi, dword ptr [r15 + 0x10]
  3a59f7:      	mov	esi, 0xffffffff
  3a59fc:      	nop	dword ptr [rax]
  3a5a00:      	cmp	rdi, r8
  3a5a03:      	je	0x3a5f49 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x949>
  3a5a09:      	cmp	r9, rdi
  3a5a0c:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a5a12:      	shl	rdi, 0x4
  3a5a16:      	mov	r8, qword ptr [rbp - 0x50]
  3a5a1a:      	mov	rax, qword ptr [r8 + rdi]
  3a5a1e:      	mov	r8, qword ptr [r8 + rdi + 0x8]
  3a5a23:      	mov	edi, dword ptr [r15 + 0x14]
  3a5a27:      	cmp	rdi, rsi
  3a5a2a:      	je	0x3a5f5b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x95b>
  3a5a30:      	cmp	r9, rdi
  3a5a33:      	jbe	0x3a7365 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d65>
  3a5a39:      	shl	rdi, 0x4
  3a5a3d:      	mov	rsi, qword ptr [rbp - 0x50]
  3a5a41:      	mov	r9, qword ptr [rsi + rdi]
  3a5a45:      	mov	r11, qword ptr [rsi + rdi + 0x8]
  3a5a4a:      	mov	rsi, qword ptr [r12 + 0x80]
  3a5a52:      	mov	edi, dword ptr [r15 + 0x4]
  3a5a56:      	cmp	rsi, rdi
  3a5a59:      	ja	0x3a5f76 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x976>
  3a5a5f:      	nop
  3a5a60:      	jmp	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a5a65:      	add	edx, 0xff03ffed
  3a5a6b:      	cmp	edx, 0x4
  3a5a6e:      	jae	0x3a5c40 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x640>
  3a5a74:      	mov	rax, qword ptr [rbp - 0x58]
  3a5a78:      	mov	rdx, qword ptr [rax + 0xd8]
  3a5a7f:      	mov	edi, dword ptr [r15 + 0x18]
  3a5a83:      	mov	eax, dword ptr [r15 + 0x1c]
  3a5a87:      	lea	ecx, [rdi + rax]
  3a5a8a:      	add	ecx, 0x2
  3a5a8d:      	cmp	rdx, rcx
  3a5a90:      	jb	0x3a73f3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1df3>
  3a5a96:      	cmp	edi, ecx
  3a5a98:      	ja	0x3a73e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1de4>
  3a5a9e:      	sub	rdx, rdi
  3a5aa1:      	lea	r8, [4*rdi]
  3a5aa9:      	test	rdx, rdx
  3a5aac:      	mov	eax, 0x0
  3a5ab1:      	cmovle	r8, rax
  3a5ab5:      	cmp	ecx, edi
  3a5ab7:      	je	0x3a73db <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ddb>
  3a5abd:      	mov	rax, qword ptr [rbp - 0x58]
  3a5ac1:      	add	r8, qword ptr [rax + 0xc8]
  3a5ac8:      	mov	eax, dword ptr [r8]
  3a5acb:      	cmp	r9, rax
  3a5ace:      	jbe	0x3a73a3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1da3>
  3a5ad4:      	sub	rcx, rdi
  3a5ad7:      	nop	word ptr [rax + rax]
  3a5ae0:      	cmp	ecx, 0x1
  3a5ae3:      	je	0x3a73cc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dcc>
  3a5ae9:      	mov	edi, dword ptr [r8 + 0x4]
  3a5aed:      	cmp	r9, rdi
  3a5af0:      	jbe	0x3a7376 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d76>
  3a5af6:      	shl	rax, 0x4
  3a5afa:      	mov	r9, qword ptr [r10 + rax]
  3a5afe:      	shl	rdi, 0x4
  3a5b02:      	mov	rsi, qword ptr [r10 + rdi]
  3a5b06:      	mov	r10d, dword ptr [r15]
  3a5b09:      	lea	eax, [r10 - 0xfc0013]
  3a5b10:      	cmp	eax, 0x2
  3a5b13:      	jae	0x3a6095 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa95>
  3a5b19:      	nop	dword ptr [rax]
  3a5b20:      	cmp	ecx, 0x3
  3a5b23:      	jb	0x3a741d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e1d>
  3a5b29:      	mov	edi, dword ptr [r8 + 0x8]
  3a5b2d:      	cmp	qword ptr [rbp - 0x38], rdi
  3a5b31:      	jbe	0x3a7376 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d76>
  3a5b37:      	nop	word ptr [rax + rax]
  3a5b40:      	cmp	ecx, 0x3
  3a5b43:      	je	0x3a740e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e0e>
  3a5b49:      	mov	eax, dword ptr [r8 + 0xc]
  3a5b4d:      	cmp	qword ptr [rbp - 0x38], rax
  3a5b51:      	jbe	0x3a7402 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e02>
  3a5b57:      	shl	rdi, 0x4
  3a5b5b:      	mov	rdx, qword ptr [rbp - 0x50]
  3a5b5f:      	mov	rdi, qword ptr [rdx + rdi]
  3a5b63:      	shl	rax, 0x4
  3a5b67:      	mov	rdx, qword ptr [rdx + rax]
  3a5b6b:      	cmp	r10d, 0xfc0013
  3a5b72:      	jne	0x3a61d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbd0>
  3a5b78:      	lea	rax, [rdi + r9]
  3a5b7c:      	mov	r10, rdi
  3a5b7f:      	and	r10, r9
  3a5b82:      	or	rdi, r9
  3a5b85:      	mov	r9, rax
  3a5b88:      	not	r9
  3a5b8b:      	and	r9, rdi
  3a5b8e:      	or	r9, r10
  3a5b91:      	shr	r9, 0x3f
  3a5b95:      	add	rdx, rsi
  3a5b98:      	add	rdx, r9
  3a5b9b:      	nop	dword ptr [rax + rax]
  3a5ba0:      	jmp	0x3a61f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf5>
  3a5ba5:      	add	edx, -0x2
  3a5ba8:      	cmp	dx, 0x21
  3a5bac:      	jae	0x3a5df5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7f5>
  3a5bb2:      	lea	rax, [r8 + 0x1c003e10]
  3a5bb9:      	movzx	ecx, dx
  3a5bbc:      	bt	rax, rcx
  3a5bc0:      	jae	0x3a5df5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7f5>
  3a5bc6:      	mov	edi, dword ptr [r15 + 0xc]
  3a5bca:      	mov	eax, 0xffffffff
  3a5bcf:      	cmp	rdi, r8
  3a5bd2:      	je	0x3a6278 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc78>
  3a5bd8:      	nop	dword ptr [rax + rax]
  3a5be0:      	cmp	r9, rdi
  3a5be3:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a5be9:      	shl	rdi, 0x4
  3a5bed:      	mov	r8, qword ptr [r10 + rdi]
  3a5bf1:      	mov	edi, dword ptr [r15 + 0x10]
  3a5bf5:      	cmp	rdi, rax
  3a5bf8:      	je	0x3a6289 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc89>
  3a5bfe:      	nop
  3a5c00:      	cmp	r9, rdi
  3a5c03:      	jbe	0x3a7365 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d65>
  3a5c09:      	shl	rdi, 0x4
  3a5c0d:      	mov	rax, qword ptr [r10 + rdi]
  3a5c11:      	mov	r9, qword ptr [r10 + rdi + 0x8]
  3a5c16:      	mov	r10d, 0x0
  3a5c1c:      	mov	r11d, 0x0
  3a5c22:      	cmp	dx, 0xc
  3a5c26:      	je	0x3a62a6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xca6>
  3a5c2c:      	jmp	0x3a62f3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcf3>
  3a5c31:      	cmp	edx, 0xe7
  3a5c37:      	jb	0x3a7118 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b18>
  3a5c3d:      	nop	dword ptr [rax]
  3a5c40:      	mov	edi, dword ptr [r15 + 0xc]
  3a5c44:      	cmp	rdi, r8
  3a5c47:      	je	0x3a5d00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x700>
  3a5c4d:      	cmp	r9, rdi
  3a5c50:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a5c56:      	shl	rdi, 0x4
  3a5c5a:      	mov	r12, qword ptr [r10 + rdi]
  3a5c5e:      	mov	rdx, qword ptr [r10 + rdi + 0x8]
  3a5c63:      	mov	edi, dword ptr [r15 + 0x10]
  3a5c67:      	cmp	rdi, r8
  3a5c6a:      	je	0x3a5d12 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x712>
  3a5c70:      	cmp	r9, rdi
  3a5c73:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a5c79:      	shl	rdi, 0x4
  3a5c7d:      	mov	rax, r10
  3a5c80:      	mov	r10, qword ptr [r10 + rdi]
  3a5c84:      	mov	rax, qword ptr [rax + rdi + 0x8]
  3a5c89:      	mov	r9d, dword ptr [r15]
  3a5c8c:      	lea	ecx, [r9 - 0x45]
  3a5c90:      	cmp	ecx, 0x7f
  3a5c93:      	ja	0x3a5d29 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x729>
  3a5c99:      	ja	0x3a6b29 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1529>
  3a5c9f:      	nop
  3a5ca0:      	jmp	qword ptr [8*rcx + 0x203ab8]
  3a5ca7:      	mov	edi, r9d
  3a5caa:      	mov	rsi, r12
  3a5cad:      	mov	rdx, r10
  3a5cb0:      	call	0x3d7280 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a5cb5:      	test	dl, dl
  3a5cb7:      	jne	0x3a7232 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c32>
  3a5cbd:      	xor	edx, edx
  3a5cbf:      	mov	r11, qword ptr [rbp - 0x40]
  3a5cc3:      	mov	r8d, 0xffffffff
  3a5cc9:      	mov	rsi, qword ptr [rbp - 0x48]
  3a5ccd:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a5cd2:      	nop	word ptr cs:[rax + rax]
  3a5ce1:      	nop	word ptr cs:[rax + rax]
  3a5cf0:      	nop	word ptr cs:[rax + rax]
  3a5cff:      	nop
  3a5d00:      	xor	r12d, r12d
  3a5d03:      	xor	edx, edx
  3a5d05:      	mov	edi, dword ptr [r15 + 0x10]
  3a5d09:      	cmp	rdi, r8
  3a5d0c:      	jne	0x3a5c70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x670>
  3a5d12:      	xor	r10d, r10d
  3a5d15:      	xor	eax, eax
  3a5d17:      	mov	r9d, dword ptr [r15]
  3a5d1a:      	lea	ecx, [r9 - 0x45]
  3a5d1e:      	nop
  3a5d20:      	cmp	ecx, 0x7f
  3a5d23:      	jbe	0x3a5c99 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x699>
  3a5d29:      	lea	ecx, [r9 - 0x1]
  3a5d2d:      	cmp	ecx, 0xd3
  3a5d33:      	ja	0x3a5e37 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x837>
  3a5d39:      	nop	dword ptr [rax]
  3a5d40:      	jmp	qword ptr [8*rcx + 0x2033d8]
  3a5d47:      	mov	rcx, qword ptr [rbp - 0x58]
  3a5d4b:      	mov	rsi, qword ptr [rcx + 0x80]
  3a5d52:      	mov	edi, dword ptr [r15 + 0x4]
  3a5d56:      	cmp	rsi, rdi
  3a5d59:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a5d5f:      	imul	rax, rdi, 0xe0
  3a5d66:      	add	rax, qword ptr [rcx + 0x78]
  3a5d6a:      	mov	rcx, qword ptr [rbp - 0x60]
  3a5d6e:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5d74:      	mov	rsi, qword ptr [rcx + 0x78]
  3a5d78:      	mov	edi, dword ptr [rax + 0x18]
  3a5d7b:      	nop	dword ptr [rax + rax]
  3a5d80:      	cmp	rsi, rdi
  3a5d83:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a5d89:      	shl	rdi, 0x3
  3a5d8d:      	add	rdi, qword ptr [rcx + 0x70]
  3a5d91:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5d97:      	mov	rcx, qword ptr [rdi]
  3a5d9a:      	nop	word ptr [rax + rax]
  3a5da0:      	test	rcx, rcx
  3a5da3:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5da9:      	mov	rsi, qword ptr [rcx + 0x10]
  3a5dad:      	mov	ecx, dword ptr [rax + 0x14]
  3a5db0:      	mov	r8d, dword ptr [rax + 0x1c]
  3a5db4:      	shl	r8, 0x20
  3a5db8:      	or	r8, rcx
  3a5dbb:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a5dc0:      	je	0x3a5dcd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7cd>
  3a5dc2:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a5dc7:      	je	0x3a6197 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb97>
  3a5dcd:      	xorps	xmm0, xmm0
  3a5dd0:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a5dd7:      	movups	xmmword ptr [rsp], xmm0
  3a5ddb:      	mov	r13, qword ptr [rbp - 0x68]
  3a5ddf:      	mov	rdi, r13
  3a5de2:      	mov	edx, r9d
  3a5de5:      	mov	rcx, r12
  3a5de8:      	xor	r9d, r9d
  3a5deb:      	call	0x3cc740 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a5df0:      	jmp	0x3a61ba <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbba>
  3a5df5:      	mov	r15, qword ptr [rbp - 0x70]
  3a5df9:      	mov	dword ptr [r15 - 0x38], r14d
  3a5dfd:      	mov	rdi, r11
  3a5e00:      	call	0x3a7700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a5e05:      	mov	rsi, qword ptr [rbp - 0x48]
  3a5e09:      	mov	r10, qword ptr [rbp - 0x50]
  3a5e0d:      	mov	r9, qword ptr [rbp - 0x38]
  3a5e11:      	mov	r8d, 0xffffffff
  3a5e17:      	mov	r11, qword ptr [rbp - 0x40]
  3a5e1b:      	nop	dword ptr [rax + rax]
  3a5e20:      	test	al, al
  3a5e22:      	jne	0x3a730e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d0e>
  3a5e28:      	inc	r14d
  3a5e2b:      	mov	dword ptr [r15 - 0x38], r14d
  3a5e2f:      	mov	r13d, r14d
  3a5e32:      	jmp	0x3a574b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14b>
  3a5e37:      	lea	ecx, [r9 - 0xfc0000]
  3a5e3e:      	nop
  3a5e40:      	cmp	ecx, 0x8
  3a5e43:      	jae	0x3a6045 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa45>
  3a5e49:      	mov	edi, r9d
  3a5e4c:      	mov	rsi, r12
  3a5e4f:      	mov	rdx, r10
  3a5e52:      	call	0x3d7280 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a5e57:      	mov	r8d, 0xffffffff
  3a5e5d:      	mov	r11, qword ptr [rbp - 0x40]
  3a5e61:      	mov	r13d, edx
  3a5e64:      	xor	edx, edx
  3a5e66:      	jmp	0x3a6b85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1585>
  3a5e6b:      	xor	eax, eax
  3a5e6d:      	xor	edx, edx
  3a5e6f:      	test	rax, rax
  3a5e72:      	jne	0x3a5858 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x258>
  3a5e78:      	jmp	0x3a594f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x34f>
  3a5e7d:      	mov	rdx, qword ptr [rbp - 0x58]
  3a5e81:      	mov	rsi, qword ptr [rdx + 0x80]
  3a5e88:      	mov	edi, dword ptr [r15 + 0x4]
  3a5e8c:      	cmp	rsi, rdi
  3a5e8f:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a5e95:      	imul	rcx, rdi, 0xe0
  3a5e9c:      	add	rcx, qword ptr [rdx + 0x78]
  3a5ea0:      	mov	rdx, qword ptr [rbp - 0x60]
  3a5ea4:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5eaa:      	mov	rsi, qword ptr [rdx + 0x78]
  3a5eae:      	mov	edi, dword ptr [rcx + 0x18]
  3a5eb1:      	cmp	rsi, rdi
  3a5eb4:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a5eba:      	shl	rdi, 0x3
  3a5ebe:      	nop
  3a5ec0:      	add	rdi, qword ptr [rdx + 0x70]
  3a5ec4:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5eca:      	mov	rdx, qword ptr [rdi]
  3a5ecd:      	test	rdx, rdx
  3a5ed0:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5ed6:      	mov	rdi, qword ptr [rdx + 0x10]
  3a5eda:      	mov	edx, dword ptr [rcx + 0x14]
  3a5edd:      	mov	ecx, dword ptr [rcx + 0x1c]
  3a5ee0:      	shl	rcx, 0x20
  3a5ee4:      	or	rcx, rdx
  3a5ee7:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a5eec:      	je	0x3a5ef9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8f9>
  3a5eee:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a5ef3:      	je	0x3a6411 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe11>
  3a5ef9:      	mov	qword ptr [rbp - 0x90], r10
  3a5f00:      	mov	qword ptr [rbp - 0x88], rax
  3a5f07:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a5f0e:      	movups	xmmword ptr [rsp], xmm0
  3a5f12:      	mov	esi, r9d
  3a5f15:      	mov	rdx, r12
  3a5f18:      	xor	r8d, r8d
  3a5f1b:      	call	0x3cdda0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a5f20:      	jmp	0x3a6b74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1574>
  3a5f25:      	xor	r10d, r10d
  3a5f28:      	xor	ecx, ecx
  3a5f2a:      	mov	r12, qword ptr [rbp - 0x58]
  3a5f2e:      	mov	edi, dword ptr [r15 + 0x10]
  3a5f32:      	mov	esi, 0xffffffff
  3a5f37:      	nop	word ptr [rax + rax]
  3a5f40:      	cmp	rdi, r8
  3a5f43:      	jne	0x3a5a09 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x409>
  3a5f49:      	xor	eax, eax
  3a5f4b:      	xor	r8d, r8d
  3a5f4e:      	mov	edi, dword ptr [r15 + 0x14]
  3a5f52:      	cmp	rdi, rsi
  3a5f55:      	jne	0x3a5a30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x430>
  3a5f5b:      	xor	r9d, r9d
  3a5f5e:      	xor	r11d, r11d
  3a5f61:      	mov	rsi, qword ptr [r12 + 0x80]
  3a5f69:      	mov	edi, dword ptr [r15 + 0x4]
  3a5f6d:      	cmp	rsi, rdi
  3a5f70:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a5f76:      	imul	rdi, rdi, 0xe0
  3a5f7d:      	nop	dword ptr [rax]
  3a5f80:      	add	rdi, qword ptr [r12 + 0x78]
  3a5f85:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5f8b:      	movzx	esi, dx
  3a5f8e:      	cmp	esi, 0xb
  3a5f91:      	setb	r12b
  3a5f95:      	mov	esi, edx
  3a5f97:      	and	esi, 0xfffc
  3a5f9d:      	cmp	esi, 0x54
  3a5fa0:      	sete	r13b
  3a5fa4:      	or	r13b, r12b
  3a5fa7:      	jne	0x3a5fbc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9bc>
  3a5fa9:      	movzx	r12d, dx
  3a5fad:      	lea	r13d, [r12 - 0x5c]
  3a5fb2:      	cmp	r13d, 0x2
  3a5fb6:      	jae	0x3a60c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xac5>
  3a5fbc:      	mov	r10, qword ptr [rbp - 0x60]
  3a5fc0:      	mov	rsi, qword ptr [r10 + 0x78]
  3a5fc4:      	mov	r9d, dword ptr [rdi + 0x18]
  3a5fc8:      	cmp	rsi, r9
  3a5fcb:      	jbe	0x3a736e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d6e>
  3a5fd1:      	shl	r9, 0x3
  3a5fd5:      	add	r9, qword ptr [r10 + 0x70]
  3a5fd9:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5fdf:      	mov	rsi, qword ptr [r9]
  3a5fe2:      	test	rsi, rsi
  3a5fe5:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a5feb:      	mov	rsi, qword ptr [rsi + 0x10]
  3a5fef:      	mov	r9d, dword ptr [rdi + 0x14]
  3a5ff3:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a5ff7:      	shl	r10, 0x20
  3a5ffb:      	or	r10, r9
  3a5ffe:      	movzx	edi, byte ptr [rdi + 0x40]
  3a6002:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a6007:      	je	0x3a6014 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa14>
  3a6009:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a600e:      	je	0x3a6165 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb65>
  3a6014:      	mov	qword ptr [rbp - 0xa0], rax
  3a601b:      	mov	qword ptr [rbp - 0x98], r8
  3a6022:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a6029:      	movups	xmmword ptr [rsp], xmm0
  3a602d:      	movzx	r9d, dil
  3a6031:      	mov	r13, qword ptr [rbp - 0x68]
  3a6035:      	mov	rdi, r13
  3a6038:      	mov	r8, r10
  3a603b:      	call	0x3cc740 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a6040:      	jmp	0x3a61ba <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbba>
  3a6045:      	lea	ecx, [r9 - 0x1000001]
  3a604c:      	cmp	ecx, 0x8
  3a604f:      	ja	0x3a6d52 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1752>
  3a6055:      	mov	edx, 0x117
  3a605a:      	bt	edx, ecx
  3a605d:      	nop	dword ptr [rax]
  3a6060:      	jae	0x3a6d3c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x173c>
  3a6066:      	movzx	eax, r9b
  3a606a:      	xor	edi, edi
  3a606c:      	add	eax, -0x2
  3a606f:      	cmp	eax, 0x7
  3a6072:      	ja	0x3a6ba5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15a5>
  3a6078:      	mov	r9, qword ptr [rbp - 0x38]
  3a607c:      	mov	r10, qword ptr [rbp - 0x50]
  3a6080:      	jmp	qword ptr [8*rax + 0x203a78]
  3a6087:      	xor	edi, edi
  3a6089:      	test	r12d, r12d
  3a608c:      	sete	dil
  3a6090:      	jmp	0x3a6bad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15ad>
  3a6095:      	mov	rax, rsi
  3a6098:      	mul	r9
  3a609b:      	cmp	dword ptr [r15], 0xfc0015
  3a60a2:      	jne	0x3a61f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf5>
  3a60a8:      	mov	rdi, r9
  3a60ab:      	sar	rdi, 0x3f
  3a60af:      	and	rdi, rsi
  3a60b2:      	sar	rsi, 0x3f
  3a60b6:      	and	rsi, r9
  3a60b9:      	add	rsi, rdi
  3a60bc:      	sub	rdx, rsi
  3a60bf:      	nop
  3a60c0:      	jmp	0x3a61f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf5>
  3a60c5:      	cmp	r12d, 0xb
  3a60c9:      	je	0x3a60d4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xad4>
  3a60cb:      	cmp	esi, 0x58
  3a60ce:      	jne	0x3a6477 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe77>
  3a60d4:      	mov	r10, qword ptr [rbp - 0x60]
  3a60d8:      	mov	rsi, qword ptr [r10 + 0x78]
  3a60dc:      	mov	r9d, dword ptr [rdi + 0x18]
  3a60e0:      	cmp	rsi, r9
  3a60e3:      	jbe	0x3a736e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d6e>
  3a60e9:      	shl	r9, 0x3
  3a60ed:      	add	r9, qword ptr [r10 + 0x70]
  3a60f1:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a60f7:      	mov	rsi, qword ptr [r9]
  3a60fa:      	nop	word ptr [rax + rax]
  3a6100:      	test	rsi, rsi
  3a6103:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a6109:      	mov	rsi, qword ptr [rsi + 0x10]
  3a610d:      	mov	r10d, dword ptr [rdi + 0x14]
  3a6111:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a6115:      	shl	r9, 0x20
  3a6119:      	or	r9, r10
  3a611c:      	movzx	edi, byte ptr [rdi + 0x40]
  3a6120:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a6125:      	je	0x3a6132 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb32>
  3a6127:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a612c:      	je	0x3a6b47 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1547>
  3a6132:      	mov	qword ptr [rbp - 0x90], rax
  3a6139:      	mov	qword ptr [rbp - 0x88], r8
  3a6140:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a6147:      	movups	xmmword ptr [rsp], xmm0
  3a614b:      	movzx	r8d, dil
  3a614f:      	mov	rdi, rsi
  3a6152:      	mov	esi, edx
  3a6154:      	mov	rdx, rcx
  3a6157:      	mov	rcx, r9
  3a615a:      	call	0x3cdda0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a615f:      	nop
  3a6160:      	jmp	0x3a6b74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1574>
  3a6165:      	mov	qword ptr [rbp - 0xa0], rax
  3a616c:      	mov	qword ptr [rbp - 0x98], r8
  3a6173:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a617a:      	movups	xmmword ptr [rsp], xmm0
  3a617e:      	movzx	r9d, dil
  3a6182:      	mov	r13, qword ptr [rbp - 0x68]
  3a6186:      	mov	rdi, r13
  3a6189:      	mov	r8, r10
  3a618c:      	jmp	0x3a61b5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbb5>
  3a618e:      	mov	rax, qword ptr [r15 + 0x18]
  3a6192:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6197:      	xorps	xmm0, xmm0
  3a619a:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a61a1:      	movups	xmmword ptr [rsp], xmm0
  3a61a5:      	mov	r13, qword ptr [rbp - 0x68]
  3a61a9:      	mov	rdi, r13
  3a61ac:      	mov	edx, r9d
  3a61af:      	mov	rcx, r12
  3a61b2:      	xor	r9d, r9d
  3a61b5:      	call	0x3cc8a0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a61ba:      	mov	rax, qword ptr [r13]
  3a61be:      	mov	rdx, qword ptr [r13 + 0x8]
  3a61c2:      	mov	qword ptr [rbp - 0x68], r13
  3a61c6:      	movzx	r13d, byte ptr [r13 + 0x10]
  3a61cb:      	jmp	0x3a6b7b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157b>
  3a61d0:      	mov	rax, r9
  3a61d3:      	sub	rax, rdi
  3a61d6:      	not	r9
  3a61d9:      	mov	r10, rdi
  3a61dc:      	and	r10, r9
  3a61df:      	xor	r9, rdi
  3a61e2:      	and	r9, rax
  3a61e5:      	or	r9, r10
  3a61e8:      	shr	r9, 0x3f
  3a61ec:      	add	rdx, r9
  3a61ef:      	sub	rsi, rdx
  3a61f2:      	mov	rdx, rsi
  3a61f5:      	mov	edi, dword ptr [r15 + 0x1c]
  3a61f9:      	cmp	ecx, edi
  3a61fb:      	mov	r9, qword ptr [rbp - 0x38]
  3a61ff:      	nop
  3a6200:      	jbe	0x3a73ae <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dae>
  3a6206:      	lea	esi, [rdi + 0x1]
  3a6209:      	cmp	ecx, esi
  3a620b:      	mov	r10, qword ptr [rbp - 0x70]
  3a620f:      	jbe	0x3a73b6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1db6>
  3a6215:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a6219:      	nop	dword ptr [rax]
  3a6220:      	cmp	r9, rcx
  3a6223:      	jbe	0x3a73c1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc1>
  3a6229:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a622d:      	shl	rcx, 0x4
  3a6231:      	mov	rsi, qword ptr [rbp - 0x50]
  3a6235:      	mov	qword ptr [rsi + rcx], rax
  3a6239:      	mov	qword ptr [rsi + rcx + 0x8], 0x0
  3a6242:      	cmp	r9, rdi
  3a6245:      	jbe	0x3a73c4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc4>
  3a624b:      	shl	rdi, 0x4
  3a624f:      	mov	qword ptr [rsi + rdi], rdx
  3a6253:      	mov	qword ptr [rsi + rdi + 0x8], 0x0
  3a625c:      	inc	r14d
  3a625f:      	mov	dword ptr [r10 - 0x38], r14d
  3a6263:      	mov	r13d, r14d
  3a6266:      	mov	r8d, 0xffffffff
  3a626c:      	mov	r10, rsi
  3a626f:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6273:      	jmp	0x3a574b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14b>
  3a6278:      	xor	r8d, r8d
  3a627b:      	mov	edi, dword ptr [r15 + 0x10]
  3a627f:      	nop
  3a6280:      	cmp	rdi, rax
  3a6283:      	jne	0x3a5bfe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5fe>
  3a6289:      	xor	eax, eax
  3a628b:      	xor	r9d, r9d
  3a628e:      	mov	r10d, 0x0
  3a6294:      	mov	r11d, 0x0
  3a629a:      	nop	word ptr [rax + rax]
  3a62a0:      	cmp	dx, 0xc
  3a62a4:      	jne	0x3a62f3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcf3>
  3a62a6:      	mov	rcx, qword ptr [rbp - 0x58]
  3a62aa:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a62b1:      	mov	edi, dword ptr [r15 + 0x18]
  3a62b5:      	add	edi, 0x2
  3a62b8:      	nop	dword ptr [rax + rax]
  3a62c0:      	cmp	rsi, rdi
  3a62c3:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a62c9:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a62d0:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a62d3:      	cmp	qword ptr [rbp - 0x38], rdi
  3a62d7:      	mov	rcx, qword ptr [rbp - 0x50]
  3a62db:      	nop	dword ptr [rax + rax]
  3a62e0:      	jbe	0x3a7365 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d65>
  3a62e6:      	shl	rdi, 0x4
  3a62ea:      	mov	r10, qword ptr [rcx + rdi]
  3a62ee:      	mov	r11, qword ptr [rcx + rdi + 0x8]
  3a62f3:      	cmp	qword ptr [rbp - 0x40], 0x0
  3a62f8:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a62fe:      	mov	rdx, qword ptr [rbp - 0x58]
  3a6302:      	mov	rsi, qword ptr [rdx + 0x80]
  3a6309:      	mov	edi, dword ptr [r15 + 0x4]
  3a630d:      	cmp	rsi, rdi
  3a6310:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a6316:      	imul	rcx, rdi, 0xe0
  3a631d:      	nop	dword ptr [rax]
  3a6320:      	add	rcx, qword ptr [rdx + 0x78]
  3a6324:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a632a:      	mov	rdx, qword ptr [rbp - 0x40]
  3a632e:      	mov	rsi, qword ptr [rdx]
  3a6331:      	mov	qword ptr [rbp - 0xd0], rax
  3a6338:      	mov	qword ptr [rbp - 0xc8], r9
  3a633f:      	mov	qword ptr [rbp - 0xc0], r10
  3a6346:      	mov	qword ptr [rbp - 0xb8], r11
  3a634d:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a6354:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a6359:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a6360:      	movups	xmmword ptr [rsp], xmm0
  3a6364:      	lea	rdi, [rbp - 0x150]
  3a636b:      	mov	rdx, qword ptr [rbp - 0x60]
  3a636f:      	call	0x3bb100 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a6374:      	movaps	xmm0, xmmword ptr [rbp - 0x150]
  3a637b:      	mov	r13, qword ptr [rbp - 0x140]
  3a6382:      	test	r13, r13
  3a6385:      	je	0x3a6438 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe38>
  3a638b:      	cmp	qword ptr [r13 + 0x8], 0x0
  3a6390:      	mov	r8d, 0xffffffff
  3a6396:      	mov	r9, qword ptr [rbp - 0x38]
  3a639a:      	mov	r10, qword ptr [rbp - 0x50]
  3a639e:      	mov	rsi, qword ptr [rbp - 0x48]
  3a63a2:      	je	0x3a644a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe4a>
  3a63a8:      	jmp	0x3a7237 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c37>
  3a63ad:      	mov	edi, dword ptr [r15 + 0x18]
  3a63b1:      	cmp	qword ptr [rbp - 0x38], rdi
  3a63b5:      	jbe	0x3a7365 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d65>
  3a63bb:      	shl	rdi, 0x4
  3a63bf:      	mov	rax, qword ptr [rbp - 0x50]
  3a63c3:      	mov	qword ptr [rax + rdi], r12
  3a63c7:      	mov	qword ptr [rax + rdi + 0x8], rdx
  3a63cc:      	mov	rax, r12
  3a63cf:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a63d4:      	mov	edi, dword ptr [r15 + 0x14]
  3a63d8:      	nop	dword ptr [rax + rax]
  3a63e0:      	cmp	rdi, r8
  3a63e3:      	je	0x3a6b97 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1597>
  3a63e9:      	cmp	qword ptr [rbp - 0x38], rdi
  3a63ed:      	jbe	0x3a7365 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d65>
  3a63f3:      	shl	rdi, 0x4
  3a63f7:      	mov	rcx, qword ptr [rbp - 0x50]
  3a63fb:      	mov	rcx, qword ptr [rcx + rdi]
  3a63ff:      	test	ecx, ecx
  3a6401:      	cmove	rdx, rax
  3a6405:      	cmove	r12, r10
  3a6409:      	mov	rax, r12
  3a640c:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a6411:      	mov	qword ptr [rbp - 0x90], r10
  3a6418:      	mov	qword ptr [rbp - 0x88], rax
  3a641f:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a6426:      	movups	xmmword ptr [rsp], xmm0
  3a642a:      	mov	esi, r9d
  3a642d:      	mov	rdx, r12
  3a6430:      	xor	r8d, r8d
  3a6433:      	jmp	0x3a6b6f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x156f>
  3a6438:      	mov	r8d, 0xffffffff
  3a643e:      	mov	r9, qword ptr [rbp - 0x38]
  3a6442:      	mov	r10, qword ptr [rbp - 0x50]
  3a6446:      	mov	rsi, qword ptr [rbp - 0x48]
  3a644a:      	mov	edi, dword ptr [r15 + 0x8]
  3a644e:      	cmp	rdi, r8
  3a6451:      	mov	r11, qword ptr [rbp - 0x40]
  3a6455:      	mov	r15, qword ptr [rbp - 0x70]
  3a6459:      	je	0x3a5e28 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x828>
  3a645f:      	nop
  3a6460:      	cmp	r9, rdi
  3a6463:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a6469:      	shl	rdi, 0x4
  3a646d:      	movups	xmmword ptr [r10 + rdi], xmm0
  3a6472:      	jmp	0x3a5e28 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x828>
  3a6477:      	mov	qword ptr [rsp], r11
  3a647b:      	mov	rsi, rcx
  3a647e:      	mov	rdx, r10
  3a6481:      	mov	rcx, rax
  3a6484:      	call	0x3d5240 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a6489:      	jmp	0x3a5cbf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6bf>
  3a648e:      	mov	edi, dword ptr [r15 + 0x1c]
  3a6492:      	dec	edi
  3a6494:      	cmp	edi, r12d
  3a6497:      	cmovae	edi, r12d
  3a649b:      	nop	dword ptr [rax + rax]
  3a64a0:      	jmp	0x3a6bad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15ad>
  3a64a5:      	xor	edi, edi
  3a64a7:      	test	r12, r12
  3a64aa:      	setne	dil
  3a64ae:      	jmp	0x3a6bad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15ad>
  3a64b3:      	xor	eax, eax
  3a64b5:      	cmp	r12, r10
  3a64b8:      	sete	al
  3a64bb:      	nop	dword ptr [rax + rax]
  3a64c0:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a64c5:      	mov	rax, qword ptr [rbp - 0x60]
  3a64c9:      	mov	rsi, qword ptr [rax + 0x78]
  3a64cd:      	mov	edi, dword ptr [r15 + 0x18]
  3a64d1:      	cmp	rsi, rdi
  3a64d4:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a64da:      	shl	rdi, 0x3
  3a64de:      	nop
  3a64e0:      	add	rdi, qword ptr [rax + 0x70]
  3a64e4:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a64ea:      	test	r11, r11
  3a64ed:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a64f3:      	mov	r13, qword ptr [rdi]
  3a64f6:      	test	r13, r13
  3a64f9:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a64ff:      	mov	rdi, qword ptr [r11]
  3a6502:      	mov	rax, qword ptr [r13 + 0x10]
  3a6506:      	movzx	eax, byte ptr [rax + 0x31]
  3a650a:      	mov	byte ptr [rbp - 0x71], al
  3a650d:      	mov	edx, r12d
  3a6510:      	test	al, 0x1
  3a6512:      	cmovne	rdx, r12
  3a6516:      	mov	rsi, r13
  3a6519:      	call	0x3c6300 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a651e:      	nop
  3a6520:      	test	dl, 0x1
  3a6523:      	jne	0x3a5cbd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6bd>
  3a6529:      	mov	rax, qword ptr [r13 + 0x10]
  3a652d:      	movzx	eax, byte ptr [rax + 0x31]
  3a6531:      	mov	byte ptr [rbp - 0x72], al
  3a6534:      	and	al, 0x1
  3a6536:      	neg	al
  3a6538:      	mov	eax, 0x0
  3a653d:      	sbb	rax, rax
  3a6540:      	mov	r8d, 0xffffffff
  3a6546:      	or	rax, r8
  3a6549:      	xor	edx, edx
  3a654b:      	mov	r11, qword ptr [rbp - 0x40]
  3a654f:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6553:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a6558:      	mov	rax, qword ptr [rbp - 0x60]
  3a655c:      	mov	rsi, qword ptr [rax + 0x60]
  3a6560:      	mov	edi, dword ptr [r15 + 0x18]
  3a6564:      	cmp	rsi, rdi
  3a6567:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a656d:      	mov	rax, qword ptr [rax + 0x58]
  3a6571:      	lea	rcx, [rax + 8*rdi]
  3a6575:      	test	rcx, rcx
  3a6578:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a657e:      	mov	rax, qword ptr [rax + 8*rdi]
  3a6582:      	test	rax, rax
  3a6585:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a658b:      	cmp	byte ptr [rax + 0x29], 0x0
  3a658f:      	mov	ecx, r12d
  3a6592:      	cmovne	rcx, r12
  3a6596:      	nop	word ptr cs:[rax + rax]
  3a65a0:      	cmp	rcx, qword ptr [rax + 0x40]
  3a65a4:      	jae	0x3a728e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c8e>
  3a65aa:      	mov	rax, qword ptr [rax + 0x38]
  3a65ae:      	mov	rax, qword ptr [rax + 8*rcx]
  3a65b2:      	jmp	0x3a70dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1adc>
  3a65b7:      	mov	rax, qword ptr [rbp - 0x60]
  3a65bb:      	mov	rsi, qword ptr [rax + 0x48]
  3a65bf:      	mov	edi, dword ptr [r15 + 0x18]
  3a65c3:      	cmp	rsi, rdi
  3a65c6:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a65cc:      	mov	rax, qword ptr [rax + 0x40]
  3a65d0:      	mov	rax, qword ptr [rax + 8*rdi]
  3a65d4:      	jmp	0x3a70dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1adc>
  3a65d9:      	mov	rax, r12
  3a65dc:      	nop	dword ptr [rax]
  3a65e0:      	test	r12, r12
  3a65e3:      	jne	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a65e9:      	jmp	0x3a7298 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c98>
  3a65ee:      	mov	rax, qword ptr [rbp - 0x60]
  3a65f2:      	mov	rsi, qword ptr [rax + 0x90]
  3a65f9:      	mov	edi, dword ptr [r15 + 0x18]
  3a65fd:      	nop	dword ptr [rax]
  3a6600:      	cmp	rsi, rdi
  3a6603:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a6609:      	mov	rax, qword ptr [rax + 0x88]
  3a6610:      	mov	rax, qword ptr [rax + 8*rdi]
  3a6614:      	mov	qword ptr [rax + 0x20], r12
  3a6618:      	mov	qword ptr [rax + 0x28], rdx
  3a661c:      	nop	dword ptr [rax]
  3a6620:      	jmp	0x3a66b2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10b2>
  3a6625:      	mov	rax, qword ptr [rbp - 0x60]
  3a6629:      	mov	rsi, qword ptr [rax + 0x90]
  3a6630:      	mov	edi, dword ptr [r15 + 0x18]
  3a6634:      	cmp	rsi, rdi
  3a6637:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a663d:      	mov	rax, qword ptr [rax + 0x88]
  3a6644:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a6648:      	mov	rax, qword ptr [rcx + 0x20]
  3a664c:      	mov	rdx, qword ptr [rcx + 0x28]
  3a6650:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6654:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a6659:      	mov	rax, qword ptr [rbp - 0x60]
  3a665d:      	mov	rsi, qword ptr [rax + 0x60]
  3a6661:      	mov	edi, dword ptr [r15 + 0x18]
  3a6665:      	cmp	rsi, rdi
  3a6668:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a666e:      	mov	rax, qword ptr [rax + 0x58]
  3a6672:      	lea	rcx, [rax + 8*rdi]
  3a6676:      	test	rcx, rcx
  3a6679:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a667f:      	mov	rax, qword ptr [rax + 8*rdi]
  3a6683:      	test	rax, rax
  3a6686:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a668c:      	cmp	byte ptr [rax + 0x29], 0x0
  3a6690:      	mov	ecx, r12d
  3a6693:      	cmovne	rcx, r12
  3a6697:      	nop	word ptr [rax + rax]
  3a66a0:      	cmp	rcx, qword ptr [rax + 0x40]
  3a66a4:      	jae	0x3a728e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c8e>
  3a66aa:      	mov	rax, qword ptr [rax + 0x38]
  3a66ae:      	mov	qword ptr [rax + 8*rcx], r10
  3a66b2:      	xor	edx, edx
  3a66b4:      	xor	eax, eax
  3a66b6:      	mov	rsi, qword ptr [rbp - 0x48]
  3a66ba:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a66bf:      	mov	edi, dword ptr [r15 + 0x18]
  3a66c3:      	cmp	qword ptr [rbp - 0x38], rdi
  3a66c7:      	jbe	0x3a7365 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d65>
  3a66cd:      	shl	rdi, 0x4
  3a66d1:      	mov	rcx, qword ptr [rbp - 0x50]
  3a66d5:      	mov	rax, qword ptr [rcx + rdi]
  3a66d9:      	mov	rdx, qword ptr [rcx + rdi + 0x8]
  3a66de:      	nop
  3a66e0:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a66e5:      	xor	eax, eax
  3a66e7:      	test	r12, r12
  3a66ea:      	sete	al
  3a66ed:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a66f2:      	mov	rax, qword ptr [rbp - 0x60]
  3a66f6:      	mov	rsi, qword ptr [rax + 0x78]
  3a66fa:      	mov	edi, dword ptr [r15 + 0x18]
  3a66fe:      	nop
  3a6700:      	cmp	rsi, rdi
  3a6703:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a6709:      	shl	rdi, 0x3
  3a670d:      	add	rdi, qword ptr [rax + 0x70]
  3a6711:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a6717:      	mov	rax, qword ptr [rdi]
  3a671a:      	nop	word ptr [rax + rax]
  3a6720:      	test	rax, rax
  3a6723:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a6729:      	mov	rdi, qword ptr [rax + 0x10]
  3a672d:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a6732:      	mov	rax, qword ptr [rdi + 0x60]
  3a6736:      	je	0x3a6f4f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x194f>
  3a673c:      	nop	dword ptr [rax]
  3a6740:      	test	rax, rax
  3a6743:      	jne	0x3a6f58 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1958>
  3a6749:      	call	0x3ca180 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a674e:      	mov	r8d, 0xffffffff
  3a6754:      	mov	r11, qword ptr [rbp - 0x40]
  3a6758:      	jmp	0x3a70dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1adc>
  3a675d:      	mov	eax, dword ptr [r15 + 0x18]
  3a6761:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6766:      	xor	r10, r12
  3a6769:      	jmp	0x3a6add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a676e:      	mov	eax, r12d
  3a6771:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6776:      	mov	ecx, r10d
  3a6779:      	shr	r12, cl
  3a677c:      	nop	dword ptr [rax]
  3a6780:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a6785:      	mov	ecx, r10d
  3a6788:      	shr	r12d, cl
  3a678b:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a6790:      	or	r10, r12
  3a6793:      	jmp	0x3a6add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a6798:      	mov	ecx, r10d
  3a679b:      	sar	r12d, cl
  3a679e:      	nop
  3a67a0:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a67a5:      	xor	r10d, r12d
  3a67a8:      	jmp	0x3a6add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a67ad:      	mov	rax, r12
  3a67b0:      	shr	rax
  3a67b3:      	movabs	rcx, 0x5555555555555555
  3a67bd:      	and	rax, rcx
  3a67c0:      	and	r12, rcx
  3a67c3:      	add	r12, rax
  3a67c6:      	mov	rax, r12
  3a67c9:      	shr	rax, 0x2
  3a67cd:      	movabs	rcx, 0x3333333333333333
  3a67d7:      	and	rax, rcx
  3a67da:      	and	r12, rcx
  3a67dd:      	add	r12, rax
  3a67e0:      	mov	rax, r12
  3a67e3:      	shr	rax, 0x4
  3a67e7:      	add	rax, r12
  3a67ea:      	movabs	rcx, 0xf0f0f0f0f0f0f0f
  3a67f4:      	and	rcx, rax
  3a67f7:      	mov	rax, rcx
  3a67fa:      	shr	rax, 0x8
  3a67fe:      	add	rax, rcx
  3a6801:      	mov	rcx, rax
  3a6804:      	shr	rcx, 0x10
  3a6808:      	add	rcx, rax
  3a680b:      	mov	rax, rcx
  3a680e:      	shr	rax, 0x20
  3a6812:      	add	eax, ecx
  3a6814:      	and	eax, 0x7f
  3a6817:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a681c:      	xor	eax, eax
  3a681e:      	cmp	r12, r10
  3a6821:      	setb	al
  3a6824:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6829:      	xor	eax, eax
  3a682b:      	cmp	r12, r10
  3a682e:      	setbe	al
  3a6831:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6836:      	xor	eax, eax
  3a6838:      	cmp	r12, r10
  3a683b:      	seta	al
  3a683e:      	nop
  3a6840:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6845:      	imul	r10d, r12d
  3a6849:      	jmp	0x3a6add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a684e:      	mov	ecx, r12d
  3a6851:      	shr	ecx, 0x10
  3a6854:      	xor	eax, eax
  3a6856:      	cmp	r12d, 0x10000
  3a685d:      	setb	al
  3a6860:      	cmovb	ecx, r12d
  3a6864:      	shl	eax, 0x4
  3a6867:      	mov	edx, ecx
  3a6869:      	shr	edx, 0x8
  3a686c:      	xor	esi, esi
  3a686e:      	cmp	ecx, 0x100
  3a6874:      	setb	sil
  3a6878:      	lea	eax, [rax + 8*rsi]
  3a687b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a687f:      	cmovb	edx, ecx
  3a6882:      	movzx	ecx, dl
  3a6885:      	movzx	ecx, byte ptr [rcx + 0x3282e9]
  3a688c:      	sub	rax, rcx
  3a688f:      	add	rax, 0x8
  3a6893:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6898:      	mov	ecx, r10d
  3a689b:      	rol	r12d, cl
  3a689e:      	nop
  3a68a0:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a68a5:      	xor	eax, eax
  3a68a7:      	cmp	r12, r10
  3a68aa:      	setg	al
  3a68ad:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a68b2:      	xor	eax, eax
  3a68b4:      	cmp	r12d, r10d
  3a68b7:      	setl	al
  3a68ba:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a68bf:      	xor	eax, eax
  3a68c1:      	cmp	r12, r10
  3a68c4:      	setne	al
  3a68c7:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a68cc:      	test	r12d, r12d
  3a68cf:      	je	0x3a6fae <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ae>
  3a68d5:      	tzcnt	eax, r12d
  3a68da:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a68df:      	add	r10d, r12d
  3a68e2:      	jmp	0x3a6add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a68e7:      	xor	eax, eax
  3a68e9:      	cmp	r12d, r10d
  3a68ec:      	setne	al
  3a68ef:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a68f4:      	xor	eax, eax
  3a68f6:      	cmp	r12, r10
  3a68f9:      	setle	al
  3a68fc:      	nop	dword ptr [rax]
  3a6900:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6905:      	mov	ecx, r10d
  3a6908:      	ror	r12d, cl
  3a690b:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a6910:      	xor	eax, eax
  3a6912:      	cmp	r12d, r10d
  3a6915:      	setbe	al
  3a6918:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a691d:      	xor	eax, eax
  3a691f:      	cmp	r12d, r10d
  3a6922:      	setge	al
  3a6925:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a692a:      	xor	eax, eax
  3a692c:      	test	r12d, r12d
  3a692f:      	sete	al
  3a6932:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6937:      	xor	eax, eax
  3a6939:      	cmp	r12d, r10d
  3a693c:      	setb	al
  3a693f:      	nop
  3a6940:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6945:      	mov	rdx, r12
  3a6948:      	mov	eax, edx
  3a694a:      	shr	eax, 0x18
  3a694d:      	mov	ecx, edx
  3a694f:      	shr	ecx, 0x10
  3a6952:      	movzx	ecx, cl
  3a6955:      	movzx	ecx, byte ptr [rcx + 0x3283e9]
  3a695c:      	add	cl, byte ptr [rax + 0x3283e9]
  3a6962:      	movzx	eax, dh
  3a6965:      	add	cl, byte ptr [rax + 0x3283e9]
  3a696b:      	movzx	eax, dl
  3a696e:      	add	cl, byte ptr [rax + 0x3283e9]
  3a6974:      	movzx	eax, cl
  3a6977:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a697c:      	add	r10, r12
  3a697f:      	nop
  3a6980:      	jmp	0x3a6add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a6985:      	mov	ecx, r10d
  3a6988:      	shl	r12d, cl
  3a698b:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a6990:      	xor	eax, eax
  3a6992:      	test	r12, r12
  3a6995:      	sete	al
  3a6998:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a699d:      	and	r10, r12
  3a69a0:      	jmp	0x3a6add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a69a5:      	sub	r12, r10
  3a69a8:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a69ad:      	xor	eax, eax
  3a69af:      	cmp	r12d, r10d
  3a69b2:      	sete	al
  3a69b5:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a69ba:      	mov	ecx, r10d
  3a69bd:      	rol	r12, cl
  3a69c0:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a69c5:      	xor	eax, eax
  3a69c7:      	cmp	r12d, r10d
  3a69ca:      	setae	al
  3a69cd:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a69d2:      	and	r12d, r10d
  3a69d5:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a69da:      	xor	eax, eax
  3a69dc:      	cmp	r12d, r10d
  3a69df:      	setle	al
  3a69e2:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a69e7:      	xor	eax, eax
  3a69e9:      	cmp	r12d, r10d
  3a69ec:      	setg	al
  3a69ef:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a69f4:      	xor	eax, eax
  3a69f6:      	mov	rcx, r12
  3a69f9:      	shr	rcx, 0x20
  3a69fd:      	cmove	rcx, r12
  3a6a01:      	setne	al
  3a6a04:      	shl	eax, 0x5
  3a6a07:      	mov	rdx, rcx
  3a6a0a:      	shr	rdx, 0x10
  3a6a0e:      	lea	rsi, [rax + 0x10]
  3a6a12:      	cmp	rcx, 0x10000
  3a6a19:      	cmovb	rdx, rcx
  3a6a1d:      	cmovb	rsi, rax
  3a6a21:      	mov	eax, edx
  3a6a23:      	shr	eax, 0x8
  3a6a26:      	lea	rcx, [rsi + 0x8]
  3a6a2a:      	cmp	rdx, 0x100
  3a6a31:      	cmovb	rax, rdx
  3a6a35:      	cmovb	rcx, rsi
  3a6a39:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6a3d:      	movzx	eax, al
  3a6a40:      	movzx	edx, byte ptr [rax + 0x3282e9]
  3a6a47:      	add	rdx, rcx
  3a6a4a:      	mov	eax, 0x40
  3a6a4f:      	sub	rax, rdx
  3a6a52:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6a57:      	xor	eax, eax
  3a6a59:      	cmp	r12, r10
  3a6a5c:      	setl	al
  3a6a5f:      	nop
  3a6a60:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6a65:      	mov	ecx, r10d
  3a6a68:      	ror	r12, cl
  3a6a6b:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a6a70:      	xor	eax, eax
  3a6a72:      	cmp	r12d, r10d
  3a6a75:      	seta	al
  3a6a78:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6a7d:      	mov	eax, r12d
  3a6a80:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6a85:      	movsx	rax, r12w
  3a6a89:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6a8e:      	movsxd	rax, r12d
  3a6a91:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6a96:      	xor	eax, eax
  3a6a98:      	cmp	r12, r10
  3a6a9b:      	sete	al
  3a6a9e:      	nop
  3a6aa0:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6aa5:      	movsxd	rax, r12d
  3a6aa8:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6aad:      	xor	eax, eax
  3a6aaf:      	cmp	r12, r10
  3a6ab2:      	setae	al
  3a6ab5:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6aba:      	xor	eax, eax
  3a6abc:      	cmp	r12, r10
  3a6abf:      	setge	al
  3a6ac2:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6ac7:      	mov	ecx, r10d
  3a6aca:      	sar	r12, cl
  3a6acd:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a6acf:      	or	r10d, r12d
  3a6ad2:      	jmp	0x3a6add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a6ad4:      	sub	r12d, r10d
  3a6ad7:      	jmp	0x3a6b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x150b>
  3a6ad9:      	imul	r10, r12
  3a6add:      	xor	edx, edx
  3a6adf:      	mov	rax, r10
  3a6ae2:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a6ae7:      	movsx	rax, r12w
  3a6aeb:      	mov	eax, eax
  3a6aed:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6af2:      	mov	eax, 0x40
  3a6af7:      	tzcnt	rax, r12
  3a6afc:      	nop	dword ptr [rax]
  3a6b00:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6b05:      	mov	ecx, r10d
  3a6b08:      	shl	r12, cl
  3a6b0b:      	xor	edx, edx
  3a6b0d:      	mov	rax, r12
  3a6b10:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a6b15:      	movsx	rax, r12b
  3a6b19:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6b1e:      	movsx	rax, r12b
  3a6b22:      	mov	eax, eax
  3a6b24:      	jmp	0x3a6fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c0>
  3a6b29:      	movzx	ecx, r9b
  3a6b2d:      	add	ecx, -0x61
  3a6b30:      	cmp	ecx, 0x5e
  3a6b33:      	ja	0x3a6dcc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17cc>
  3a6b39:      	xor	edx, edx
  3a6b3b:      	xor	eax, eax
  3a6b3d:      	nop	dword ptr [rax]
  3a6b40:      	jmp	qword ptr [8*rcx + 0x203eb8]
  3a6b47:      	mov	qword ptr [rbp - 0x90], rax
  3a6b4e:      	mov	qword ptr [rbp - 0x88], r8
  3a6b55:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a6b5c:      	movups	xmmword ptr [rsp], xmm0
  3a6b60:      	movzx	r8d, dil
  3a6b64:      	mov	rdi, rsi
  3a6b67:      	mov	esi, edx
  3a6b69:      	mov	rdx, rcx
  3a6b6c:      	mov	rcx, r9
  3a6b6f:      	call	0x3ce0a0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a6b74:      	mov	r13d, eax
  3a6b77:      	xor	edx, edx
  3a6b79:      	xor	eax, eax
  3a6b7b:      	mov	r11, qword ptr [rbp - 0x40]
  3a6b7f:      	mov	r8d, 0xffffffff
  3a6b85:      	test	r13b, r13b
  3a6b88:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6b8c:      	je	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a6b92:      	jmp	0x3a729b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c9b>
  3a6b97:      	mov	rdx, rax
  3a6b9a:      	mov	rax, r10
  3a6b9d:      	nop	dword ptr [rax]
  3a6ba0:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a6ba5:      	mov	r9, qword ptr [rbp - 0x38]
  3a6ba9:      	mov	r10, qword ptr [rbp - 0x50]
  3a6bad:      	mov	rax, qword ptr [rbp - 0x58]
  3a6bb1:      	mov	rsi, qword ptr [rax + 0xe8]
  3a6bb8:      	add	edi, dword ptr [r15 + 0x18]
  3a6bbc:      	nop	dword ptr [rax]
  3a6bc0:      	cmp	rsi, rdi
  3a6bc3:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a6bc9:      	mov	rax, qword ptr [rax + 0xe0]
  3a6bd0:      	lea	rcx, [rdi + 2*rdi]
  3a6bd4:      	mov	r13d, dword ptr [rax + 4*rcx]
  3a6bd8:      	mov	r15d, dword ptr [rax + 4*rcx + 0x4]
  3a6bdd:      	mov	esi, dword ptr [rax + 4*rcx + 0x8]
  3a6be1:      	cmp	r13d, r14d
  3a6be4:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a6be9:      	ja	0x3a6c9c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x169c>
  3a6bef:      	mov	rax, qword ptr [r11 + 0x160]
  3a6bf6:      	test	rax, rax
  3a6bf9:      	je	0x3a6c0f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x160f>
  3a6bfb:      	nop	dword ptr [rax + rax]
  3a6c00:      	test	r11, r11
  3a6c03:      	je	0x3a6c75 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1675>
  3a6c05:      	mov	eax, dword ptr [rax]
  3a6c07:      	test	eax, eax
  3a6c09:      	je	0x3a6f93 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1993>
  3a6c0f:      	mov	rax, qword ptr [r11 + 0x8]
  3a6c13:      	test	rax, rax
  3a6c16:      	je	0x3a744d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e4d>
  3a6c1c:      	mov	rdi, qword ptr [r11 + 0x10]
  3a6c20:      	mov	rcx, qword ptr [rax + 0x8]
  3a6c24:      	test	byte ptr [rcx + 0x17], 0x20
  3a6c28:      	jne	0x3a6c2f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x162f>
  3a6c2a:      	mov	r14, rsi
  3a6c2d:      	jmp	0x3a6c49 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1649>
  3a6c2f:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a6c33:      	and	cl, 0x1f
  3a6c36:      	mov	r14, rsi
  3a6c39:      	cmp	cl, 0x16
  3a6c3c:      	je	0x3a6c49 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1649>
  3a6c3e:      	mov	qword ptr [rbp - 0x90], rdi
  3a6c45:      	mov	rdi, qword ptr [rbp - 0x68]
  3a6c49:      	call	qword ptr [rax + 0x28]
  3a6c4c:      	mov	r8d, 0xffffffff
  3a6c52:      	mov	r9, qword ptr [rbp - 0x38]
  3a6c56:      	mov	r10, qword ptr [rbp - 0x50]
  3a6c5a:      	mov	rsi, r14
  3a6c5d:      	nop	dword ptr [rax]
  3a6c60:      	test	rax, rax
  3a6c63:      	je	0x3a6c98 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1698>
  3a6c65:      	cmp	qword ptr [rax + 0x8], 0x0
  3a6c6a:      	mov	r11, qword ptr [rbp - 0x40]
  3a6c6e:      	je	0x3a6c9c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x169c>
  3a6c70:      	jmp	0x3a7227 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c27>
  3a6c75:      	mov	edi, 0x1
  3a6c7a:      	mov	r14, rsi
  3a6c7d:      	call	0x61eb00 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a6c82:      	mov	rsi, r14
  3a6c85:      	mov	r10, qword ptr [rbp - 0x50]
  3a6c89:      	mov	r9, qword ptr [rbp - 0x38]
  3a6c8d:      	mov	r8d, 0xffffffff
  3a6c93:      	test	rax, rax
  3a6c96:      	jne	0x3a6c65 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1665>
  3a6c98:      	mov	r11, qword ptr [rbp - 0x40]
  3a6c9c:      	mov	rax, qword ptr [rbp - 0x58]
  3a6ca0:      	mov	rax, qword ptr [rax + 0x108]
  3a6ca7:      	lea	edx, [rsi + r15]
  3a6cab:      	cmp	rax, rdx
  3a6cae:      	jb	0x3a743b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e3b>
  3a6cb4:      	cmp	r15d, edx
  3a6cb7:      	ja	0x3a742c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e2c>
  3a6cbd:      	sub	rax, r15
  3a6cc0:      	lea	rcx, [8*r15]
  3a6cc8:      	test	rax, rax
  3a6ccb:      	mov	eax, 0x0
  3a6cd0:      	cmovle	rcx, rax
  3a6cd4:      	test	esi, esi
  3a6cd6:      	je	0x3a6d33 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1733>
  3a6cdc:      	mov	rax, qword ptr [rbp - 0x58]
  3a6ce0:      	mov	rax, qword ptr [rax + 0xf8]
  3a6ce7:      	sub	rdx, r15
  3a6cea:      	add	rcx, rax
  3a6ced:      	add	rcx, 0x4
  3a6cf1:      	xor	esi, esi
  3a6cf3:      	nop	word ptr cs:[rax + rax]
  3a6d00:      	mov	edi, dword ptr [rcx + 8*rsi]
  3a6d03:      	cmp	r9, rdi
  3a6d06:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a6d0c:      	mov	eax, dword ptr [rcx + 8*rsi - 0x4]
  3a6d10:      	cmp	r9, rax
  3a6d13:      	jbe	0x3a735a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d5a>
  3a6d19:      	shl	rdi, 0x4
  3a6d1d:      	shl	rax, 0x4
  3a6d21:      	movups	xmm0, xmmword ptr [r10 + rdi]
  3a6d26:      	movups	xmmword ptr [r10 + rax], xmm0
  3a6d2b:      	inc	rsi
  3a6d2e:      	cmp	rdx, rsi
  3a6d31:      	jne	0x3a6d00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1700>
  3a6d33:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6d37:      	jmp	0x3a574b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14b>
  3a6d3c:      	nop	dword ptr [rax]
  3a6d40:      	cmp	ecx, 0x3
  3a6d43:      	je	0x3a715e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5e>
  3a6d49:      	cmp	ecx, 0x6
  3a6d4c:      	je	0x3a717f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b7f>
  3a6d52:      	lea	ecx, [r9 - 0x12]
  3a6d56:      	cmp	ecx, 0x4
  3a6d59:      	jb	0x3a7134 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b34>
  3a6d5f:      	mov	edi, dword ptr [r15 + 0x14]
  3a6d63:      	cmp	rdi, r8
  3a6d66:      	je	0x3a6dd5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17d5>
  3a6d68:      	cmp	qword ptr [rbp - 0x38], rdi
  3a6d6c:      	jbe	0x3a7365 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d65>
  3a6d72:      	shl	rdi, 0x4
  3a6d76:      	mov	rcx, qword ptr [rbp - 0x50]
  3a6d7a:      	movups	xmm0, xmmword ptr [rcx + rdi]
  3a6d7e:      	mov	ecx, r9d
  3a6d81:      	and	ecx, 0xffff0000
  3a6d87:      	cmp	ecx, 0xfe0000
  3a6d8d:      	jne	0x3a6de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17e9>
  3a6d8f:      	movzx	r9d, r9w
  3a6d93:      	cmp	r9d, 0x3
  3a6d97:      	jne	0x3a6ffb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19fb>
  3a6d9d:      	nop	dword ptr [rax]
  3a6da0:      	test	r11, r11
  3a6da3:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a6da9:      	mov	rax, qword ptr [r11]
  3a6dac:      	test	rax, rax
  3a6daf:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a6db5:      	mov	rdi, qword ptr [rax + 0x8]
  3a6db9:      	call	0x3c9600 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a6dbe:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6dc2:      	mov	r8d, 0xffffffff
  3a6dc8:      	mov	r11, qword ptr [rbp - 0x40]
  3a6dcc:      	xor	edx, edx
  3a6dce:      	xor	eax, eax
  3a6dd0:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a6dd5:      	xorps	xmm0, xmm0
  3a6dd8:      	mov	ecx, r9d
  3a6ddb:      	and	ecx, 0xffff0000
  3a6de1:      	cmp	ecx, 0xfe0000
  3a6de7:      	je	0x3a6d8f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x178f>
  3a6de9:      	add	r9d, 0xff03fff8
  3a6df0:      	cmp	r9d, 0xa
  3a6df4:      	jae	0x3a745c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e5c>
  3a6dfa:      	mov	rdx, qword ptr [rbp - 0x58]
  3a6dfe:      	mov	rsi, qword ptr [rdx + 0x80]
  3a6e05:      	mov	edi, dword ptr [r15 + 0x4]
  3a6e09:      	cmp	rsi, rdi
  3a6e0c:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a6e12:      	imul	rcx, rdi, 0xe0
  3a6e19:      	nop	dword ptr [rax]
  3a6e20:      	add	rcx, qword ptr [rdx + 0x78]
  3a6e24:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a6e2a:      	movups	xmm1, xmmword ptr [rcx + 0x98]
  3a6e31:      	movups	xmm2, xmmword ptr [rcx + 0x88]
  3a6e38:      	movups	xmm3, xmmword ptr [rcx + 0x78]
  3a6e3c:      	movups	xmm4, xmmword ptr [rcx + 0x68]
  3a6e40:      	movups	xmm5, xmmword ptr [rcx + 0x58]
  3a6e44:      	movups	xmm6, xmmword ptr [rcx + 0x48]
  3a6e48:      	movzx	edx, byte ptr [rcx + 0x40]
  3a6e4c:      	movups	xmm7, xmmword ptr [rcx + 0x30]
  3a6e50:      	movups	xmm8, xmmword ptr [rcx + 0x20]
  3a6e55:      	movups	xmm9, xmmword ptr [rcx + 0x10]
  3a6e5a:      	mov	rsi, qword ptr [rcx + 0x8]
  3a6e5e:      	mov	ecx, dword ptr [rcx]
  3a6e60:      	mov	dword ptr [rbp - 0x208], ecx
  3a6e66:      	mov	qword ptr [rbp - 0x200], rsi
  3a6e6d:      	movups	xmmword ptr [rbp - 0x1f8], xmm9
  3a6e75:      	movups	xmmword ptr [rbp - 0x1e8], xmm8
  3a6e7d:      	movups	xmmword ptr [rbp - 0x1d8], xmm7
  3a6e84:      	mov	byte ptr [rbp - 0x1c8], dl
  3a6e8a:      	movups	xmmword ptr [rbp - 0x1c0], xmm6
  3a6e91:      	movups	xmmword ptr [rbp - 0x1b0], xmm5
  3a6e98:      	movups	xmmword ptr [rbp - 0x1a0], xmm4
  3a6e9f:      	movups	xmmword ptr [rbp - 0x190], xmm3
  3a6ea6:      	movups	xmmword ptr [rbp - 0x180], xmm2
  3a6ead:      	movups	xmmword ptr [rbp - 0x170], xmm1
  3a6eb4:      	mov	qword ptr [rbp - 0xe0], r10
  3a6ebb:      	mov	qword ptr [rbp - 0xd8], rax
  3a6ec2:      	movaps	xmmword ptr [rbp - 0x160], xmm0
  3a6ec9:      	mov	ecx, 0x15
  3a6ece:      	mov	rdi, rsp
  3a6ed1:      	lea	rsi, [rbp - 0x208]
  3a6ed8:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a6edb:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a6ee2:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a6eea:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a6ef1:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a6ef9:      	lea	rdi, [rbp - 0x108]
  3a6f00:      	mov	rsi, r11
  3a6f03:      	mov	rcx, r12
  3a6f06:      	mov	rdx, qword ptr [rbp - 0x60]
  3a6f0a:      	call	0x3ac800 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a6f0f:      	mov	rax, qword ptr [rbp - 0x108]
  3a6f16:      	mov	rdx, qword ptr [rbp - 0x100]
  3a6f1d:      	movzx	r13d, byte ptr [rbp - 0xf8]
  3a6f25:      	mov	rsi, qword ptr [rbp - 0xf0]
  3a6f2c:      	test	rsi, rsi
  3a6f2f:      	je	0x3a70d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ad3>
  3a6f35:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a6f3a:      	mov	r11, qword ptr [rbp - 0x40]
  3a6f3e:      	mov	r8d, 0xffffffff
  3a6f44:      	je	0x3a6b85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1585>
  3a6f4a:      	jmp	0x3a7323 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d23>
  3a6f4f:      	test	rax, rax
  3a6f52:      	je	0x3a70e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ae7>
  3a6f58:      	mov	rax, qword ptr [rax]
  3a6f5b:      	nop	dword ptr [rax + rax]
  3a6f60:      	test	rdi, rdi
  3a6f63:      	je	0x3a70f4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af4>
  3a6f69:      	mov	rcx, qword ptr [rdi + 0x38]
  3a6f6d:      	test	rcx, rcx
  3a6f70:      	mov	edx, 0x10000
  3a6f75:      	cmove	rcx, rdx
  3a6f79:      	mov	rdx, rax
  3a6f7c:      	or	rdx, rcx
  3a6f7f:      	shr	rdx, 0x20
  3a6f83:      	je	0x3a70d8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ad8>
  3a6f89:      	xor	edx, edx
  3a6f8b:      	div	rcx
  3a6f8e:      	jmp	0x3a70dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1adc>
  3a6f93:      	xor	eax, eax
  3a6f95:      	xor	edx, edx
  3a6f97:      	nop	word ptr [rax + rax]
  3a6fa0:      	test	rax, rax
  3a6fa3:      	jne	0x3a6c65 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1665>
  3a6fa9:      	jmp	0x3a6c98 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1698>
  3a6fae:      	mov	eax, 0x20
  3a6fb3:      	nop	word ptr cs:[rax + rax]
  3a6fc0:      	xor	edx, edx
  3a6fc2:      	mov	edi, dword ptr [r15 + 0x8]
  3a6fc6:      	cmp	rdi, r8
  3a6fc9:      	mov	r9, qword ptr [rbp - 0x38]
  3a6fcd:      	mov	r10, qword ptr [rbp - 0x50]
  3a6fd1:      	je	0x3a5740 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x140>
  3a6fd7:      	nop	word ptr [rax + rax]
  3a6fe0:      	cmp	r9, rdi
  3a6fe3:      	jbe	0x3a733c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d3c>
  3a6fe9:      	shl	rdi, 0x4
  3a6fed:      	mov	qword ptr [r10 + rdi], rax
  3a6ff1:      	mov	qword ptr [r10 + rdi + 0x8], rdx
  3a6ff6:      	jmp	0x3a5740 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x140>
  3a6ffb:      	mov	rcx, qword ptr [rbp - 0x58]
  3a6fff:      	mov	rsi, qword ptr [rcx + 0x80]
  3a7006:      	mov	edi, dword ptr [r15 + 0x4]
  3a700a:      	cmp	rsi, rdi
  3a700d:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a7013:      	imul	rax, rdi, 0xe0
  3a701a:      	nop	word ptr [rax + rax]
  3a7020:      	add	rax, qword ptr [rcx + 0x78]
  3a7024:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a702a:      	mov	rcx, qword ptr [rbp - 0x60]
  3a702e:      	mov	rsi, qword ptr [rcx + 0x78]
  3a7032:      	mov	edi, dword ptr [rax + 0x18]
  3a7035:      	cmp	rsi, rdi
  3a7038:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a703e:      	shl	rdi, 0x3
  3a7042:      	add	rdi, qword ptr [rcx + 0x70]
  3a7046:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a704c:      	mov	rsi, qword ptr [rdi]
  3a704f:      	test	rsi, rsi
  3a7052:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a7058:      	mov	rcx, qword ptr [r11]
  3a705b:      	mov	rdx, qword ptr [r11 + 0x8]
  3a705f:      	mov	r8, qword ptr [rcx + 0x8]
  3a7063:      	mov	rcx, qword ptr [r11 + 0x10]
  3a7067:      	mov	rsi, qword ptr [rsi + 0x10]
  3a706b:      	mov	edi, dword ptr [rax + 0x14]
  3a706e:      	mov	eax, dword ptr [rax + 0x1c]
  3a7071:      	shl	rax, 0x20
  3a7075:      	or	rax, rdi
  3a7078:      	movlps	qword ptr [rsp + 0x18], xmm0
  3a707d:      	mov	qword ptr [rsp + 0x10], r10
  3a7082:      	mov	qword ptr [rsp + 0x8], rax
  3a7087:      	mov	qword ptr [rsp], r12
  3a708b:      	lea	rdi, [rbp - 0x130]
  3a7092:      	call	0x3caac0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a7097:      	mov	rax, qword ptr [rbp - 0x130]
  3a709e:      	mov	rdx, qword ptr [rbp - 0x128]
  3a70a5:      	movzx	r13d, byte ptr [rbp - 0x120]
  3a70ad:      	mov	rsi, qword ptr [rbp - 0x118]
  3a70b4:      	test	rsi, rsi
  3a70b7:      	je	0x3a7113 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b13>
  3a70b9:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a70be:      	mov	r11, qword ptr [rbp - 0x40]
  3a70c2:      	mov	r8d, 0xffffffff
  3a70c8:      	je	0x3a6b85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1585>
  3a70ce:      	jmp	0x3a732c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d2c>
  3a70d3:      	jmp	0x3a6b7b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157b>
  3a70d8:      	xor	edx, edx
  3a70da:      	div	ecx
  3a70dc:      	xor	edx, edx
  3a70de:      	mov	rsi, qword ptr [rbp - 0x48]
  3a70e2:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a70e7:      	mov	rax, qword ptr [rdi + 0x8]
  3a70eb:      	test	rdi, rdi
  3a70ee:      	jne	0x3a6f69 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1969>
  3a70f4:      	mov	edi, 0x1
  3a70f9:      	call	0x61eb00 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a70fe:      	mov	r8d, 0xffffffff
  3a7104:      	mov	r11, qword ptr [rbp - 0x40]
  3a7108:      	xor	edx, edx
  3a710a:      	mov	rsi, qword ptr [rbp - 0x48]
  3a710e:      	jmp	0x3a6fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a7113:      	jmp	0x3a6b7b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157b>
  3a7118:      	mov	rax, qword ptr [rbp - 0x70]
  3a711c:      	mov	dword ptr [rax - 0x38], r14d
  3a7120:      	mov	rdi, r11
  3a7123:      	mov	r14, r11
  3a7126:      	call	0x3a9a60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a712b:      	test	al, al
  3a712d:      	je	0x3a719f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b9f>
  3a712f:      	jmp	0x3a730e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d0e>
  3a7134:      	mov	rax, qword ptr [rbp - 0x70]
  3a7138:      	mov	dword ptr [rax - 0x38], r14d
  3a713c:      	mov	rdi, r11
  3a713f:      	mov	r14, r11
  3a7142:      	call	0x3a8740 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a7147:      	jmp	0x3a7192 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b92>
  3a7149:      	mov	rax, qword ptr [rbp - 0x70]
  3a714d:      	mov	dword ptr [rax - 0x38], r14d
  3a7151:      	mov	rdi, r11
  3a7154:      	mov	r14, r11
  3a7157:      	call	0x3a9540 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a715c:      	jmp	0x3a7192 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b92>
  3a715e:      	mov	rax, qword ptr [rbp - 0x70]
  3a7162:      	mov	dword ptr [rax - 0x38], r14d
  3a7166:      	mov	rdi, r11
  3a7169:      	mov	r14, r11
  3a716c:      	call	0x3a8240 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a7171:      	test	al, al
  3a7173:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a7178:      	je	0x3a719f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b9f>
  3a717a:      	jmp	0x3a7338 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d38>
  3a717f:      	mov	rax, qword ptr [rbp - 0x70]
  3a7183:      	mov	dword ptr [rax - 0x38], r14d
  3a7187:      	mov	rdi, r11
  3a718a:      	mov	r14, r11
  3a718d:      	call	0x3a9a60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a7192:      	test	al, al
  3a7194:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a7199:      	jne	0x3a730e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d0e>
  3a719f:      	mov	rcx, qword ptr [r14 + 0x48]
  3a71a3:      	test	rcx, rcx
  3a71a6:      	mov	r11, r14
  3a71a9:      	mov	r8d, 0xffffffff
  3a71af:      	mov	r14, qword ptr [rbp - 0x68]
  3a71b3:      	jg	0x3a5649 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x49>
  3a71b9:      	mov	edi, 0x333878
  3a71be:      	mov	esi, 0x299e80
  3a71c3:      	call	0x619ac0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a71c8:      	mov	edi, 0x1
  3a71cd:      	call	0x61eb00 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a71d2:      	jmp	0x3a730e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d0e>
  3a71d7:      	mov	edi, 0x28
  3a71dc:      	call	0x638250 <GC_malloc@plt>
  3a71e1:      	test	rax, rax
  3a71e4:      	je	0x3a746b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e6b>
  3a71ea:      	mov	r14, rax
  3a71ed:      	add	r13, r12
  3a71f0:      	mov	edi, 0x28
  3a71f5:      	call	0x619dc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a71fa:      	mov	rax, qword ptr [r13 + 0x8]
  3a71fe:      	movaps	xmm0, xmmword ptr [rip + 0x34ea5b] # 0x6f5c60 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a7205:      	movups	xmmword ptr [r14], xmm0
  3a7209:      	mov	qword ptr [r14 + 0x10], rax
  3a720d:      	mov	qword ptr [r14 + 0x18], 0x299e90
  3a7215:      	mov	qword ptr [r14 + 0x20], 0xe
  3a721d:      	mov	esi, 0x2ba890
  3a7222:      	jmp	0x3a7300 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d00>
  3a7227:      	mov	rdi, r11
  3a722a:      	mov	rsi, rax
  3a722d:      	jmp	0x3a7307 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d07>
  3a7232:      	mov	r13d, edx
  3a7235:      	jmp	0x3a729b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c9b>
  3a7237:      	mov	r14, qword ptr [rbp - 0x138]
  3a723e:      	mov	edi, 0x8
  3a7243:      	call	0x638250 <GC_malloc@plt>
  3a7248:      	test	rax, rax
  3a724b:      	je	0x3a746b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e6b>
  3a7251:      	mov	r15, rax
  3a7254:      	mov	edi, 0x8
  3a7259:      	call	0x619dc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a725e:      	mov	edx, 0x299e00
  3a7263:      	mov	rdi, r13
  3a7266:      	mov	rsi, r14
  3a7269:      	mov	rcx, r15
  3a726c:      	call	0x5eb600 <errors.As>
  3a7271:      	test	al, 0x1
  3a7273:      	je	0x3a7285 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c85>
  3a7275:      	mov	rax, qword ptr [r15]
  3a7278:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a727f:      	mov	ecx, dword ptr [rcx + 0x50]
  3a7282:      	mov	dword ptr [rax + 0x4], ecx
  3a7285:      	mov	rdi, qword ptr [rbp - 0x40]
  3a7289:      	mov	rsi, r13
  3a728c:      	jmp	0x3a7304 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d04>
  3a728e:      	mov	r13b, 0x3
  3a7291:      	jmp	0x3a729b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c9b>
  3a7293:      	mov	r13b, 0x1
  3a7296:      	jmp	0x3a729b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c9b>
  3a7298:      	mov	r13b, 0x9
  3a729b:      	mov	edi, 0x10
  3a72a0:      	call	0x638300 <GC_malloc_atomic@plt>
  3a72a5:      	test	rax, rax
  3a72a8:      	je	0x3a746b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e6b>
  3a72ae:      	mov	r14, rax
  3a72b1:      	mov	edi, 0x10
  3a72b6:      	call	0x619dc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a72bb:      	xorps	xmm0, xmm0
  3a72be:      	movups	xmmword ptr [r14], xmm0
  3a72c2:      	mov	rax, qword ptr [rbp - 0x58]
  3a72c6:      	mov	rsi, qword ptr [rax + 0x80]
  3a72cd:      	mov	edi, dword ptr [r15 + 0x4]
  3a72d1:      	cmp	rsi, rdi
  3a72d4:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d4a>
  3a72d6:      	mov	rax, qword ptr [rax + 0x78]
  3a72da:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a72e1:      	mov	ecx, dword ptr [rcx + 0x50]
  3a72e4:      	imul	rdx, rdi, 0xe0
  3a72eb:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a72f0:      	mov	byte ptr [r14], r13b
  3a72f3:      	mov	dword ptr [r14 + 0x4], ecx
  3a72f7:      	mov	qword ptr [r14 + 0x8], rax
  3a72fb:      	mov	esi, 0x29a218
  3a7300:      	mov	rdi, qword ptr [rbp - 0x40]
  3a7304:      	mov	rdx, r14
  3a7307:      	call	0x3a4f20 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a730c:      	mov	al, 0x3
  3a730e:      	mov	rdx, rbx
  3a7311:      	add	rsp, 0x2a8
  3a7318:      	pop	rbx
  3a7319:      	pop	r12
  3a731b:      	pop	r13
  3a731d:      	pop	r14
  3a731f:      	pop	r15
  3a7321:      	pop	rbp
  3a7322:      	ret
  3a7323:      	mov	rdx, qword ptr [rbp - 0xe8]
  3a732a:      	jmp	0x3a7333 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d33>
  3a732c:      	mov	rdx, qword ptr [rbp - 0x110]
  3a7333:      	mov	rdi, r11
  3a7336:      	jmp	0x3a7307 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d07>
  3a7338:      	mov	al, 0x1
  3a733a:      	jmp	0x3a730e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d0e>
  3a733c:      	mov	rsi, r9
  3a733f:      	call	0x619eb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a7344:      	mov	rdi, r14
  3a7347:      	mov	rsi, r13
  3a734a:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a734f:      	mov	rdi, r15
  3a7352:      	mov	rsi, r13
  3a7355:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a735a:      	mov	rdi, rax
  3a735d:      	mov	rsi, r9
  3a7360:      	call	0x619eb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a7365:      	mov	rsi, qword ptr [rbp - 0x38]
  3a7369:      	call	0x619eb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a736e:      	mov	rdi, r9
  3a7371:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7376:      	mov	rsi, qword ptr [rbp - 0x38]
  3a737a:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a737f:      	mov	rdi, rax
  3a7382:      	mov	esi, 0x1
  3a7387:      	mov	ecx, 0x2
  3a738c:      	call	0x61ae80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a7391:      	mov	esi, 0x1
  3a7396:      	mov	rdx, rax
  3a7399:      	mov	ecx, 0x3
  3a739e:      	call	0x61ae80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a73a3:      	mov	rdi, rax
  3a73a6:      	mov	rsi, r9
  3a73a9:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a73ae:      	mov	rsi, rcx
  3a73b1:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a73b6:      	mov	rdi, rsi
  3a73b9:      	mov	rsi, rcx
  3a73bc:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a73c1:      	mov	rdi, rcx
  3a73c4:      	mov	rsi, r9
  3a73c7:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a73cc:      	mov	edi, 0x1
  3a73d1:      	mov	esi, 0x1
  3a73d6:      	call	0x619eb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a73db:      	xor	edi, edi
  3a73dd:      	xor	esi, esi
  3a73df:      	call	0x619eb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a73e4:      	xor	esi, esi
  3a73e6:      	mov	rdx, rcx
  3a73e9:      	mov	ecx, 0x3
  3a73ee:      	call	0x61ae80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a73f3:      	mov	rdi, rcx
  3a73f6:      	xor	esi, esi
  3a73f8:      	mov	ecx, 0x2
  3a73fd:      	call	0x61ae80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a7402:      	mov	rdi, rax
  3a7405:      	mov	rsi, qword ptr [rbp - 0x38]
  3a7409:      	call	0x61b720 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a740e:      	mov	edi, 0x3
  3a7413:      	mov	esi, 0x3
  3a7418:      	call	0x619eb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a741d:      	mov	edi, 0x2
  3a7422:      	mov	esi, 0x2
  3a7427:      	call	0x619eb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a742c:      	mov	rdi, r15
  3a742f:      	xor	esi, esi
  3a7431:      	mov	ecx, 0x3
  3a7436:      	call	0x61ae80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a743b:      	mov	rdi, rdx
  3a743e:      	xor	esi, esi
  3a7440:      	mov	rdx, rax
  3a7443:      	mov	ecx, 0x2
  3a7448:      	call	0x61ae80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a744d:      	mov	edi, 0x3293c8
  3a7452:      	mov	esi, 0x3296f8
  3a7457:      	call	0x619ac0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a745c:      	mov	edi, 0x333878
  3a7461:      	mov	esi, 0x299ea0
  3a7466:      	call	0x619ac0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a746b:      	mov	edi, 0x333878
  3a7470:      	mov	esi, 0x3293a0
  3a7475:      	call	0x619ac0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a747a:      	int3
  3a747b:      	int3
  3a747c:      	int3
  3a747d:      	int3
  3a747e:      	int3
  3a747f:      	int3
