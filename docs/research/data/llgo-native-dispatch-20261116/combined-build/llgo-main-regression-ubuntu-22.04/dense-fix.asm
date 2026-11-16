
/home/runner/work/wasmvm/wasmvm/main-regression/dense-fix:	file format elf64-x86-64

Disassembly of section .text:

00000000003a5780 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a5780:      	push	rbp
  3a5781:      	mov	rbp, rsp
  3a5784:      	push	r15
  3a5786:      	push	r14
  3a5788:      	push	r13
  3a578a:      	push	r12
  3a578c:      	push	rbx
  3a578d:      	sub	rsp, 0x2b8
  3a5794:      	mov	rcx, qword ptr [rdi + 0x48]
  3a5798:      	nop	dword ptr [rax + rax]
  3a57a0:      	test	rcx, rcx
  3a57a3:      	jle	0x3a754f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dcf>
  3a57a9:      	mov	rbx, rsi
  3a57ac:      	mov	r11, rdi
  3a57af:      	movzx	r15d, byte ptr [rdi + 0x20]
  3a57b4:      	mov	r8d, 0xffffffff
  3a57ba:      	lea	rax, [rbp - 0x98]
  3a57c1:      	mov	qword ptr [rbp - 0x78], rax
  3a57c5:      	mov	qword ptr [rbp - 0x50], rdi
  3a57c9:      	mov	byte ptr [rbp - 0x29], r15b
  3a57cd:      	mov	r10, qword ptr [r11 + 0x40]
  3a57d1:      	shl	rcx, 0x6
  3a57d5:      	mov	byte ptr [r10 + rcx - 0x30], 0x0
  3a57db:      	mov	rsi, qword ptr [r10 + rcx - 0x40]
  3a57e0:      	mov	rax, qword ptr [rsi + 0x48]
  3a57e4:      	mov	qword ptr [rbp - 0x68], rax
  3a57e8:      	mov	r9, qword ptr [rax + 0x98]
  3a57ef:      	mov	r12, qword ptr [r10 + rcx - 0x28]
  3a57f4:      	mov	rdx, qword ptr [r11 + 0x88]
  3a57fb:      	lea	rax, [r9 + r12]
  3a57ff:      	nop
  3a5800:      	cmp	rax, rdx
  3a5803:      	ja	0x3a7473 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cf3>
  3a5809:      	cmp	r12, rax
  3a580c:      	ja	0x3a7485 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d05>
  3a5812:      	mov	rdi, qword ptr [rbp - 0x68]
  3a5816:      	mov	r14, qword ptr [rdi + 0xb8]
  3a581d:      	mov	r13d, dword ptr [r10 + rcx - 0x38]
  3a5822:      	sub	rdx, r12
  3a5825:      	shl	r12, 0x4
  3a5829:      	test	rdx, rdx
  3a582c:      	mov	eax, 0x0
  3a5831:      	cmovle	r12, rax
  3a5835:      	mov	qword ptr [rbp - 0x40], r14
  3a5839:      	nop	dword ptr [rax]
  3a5840:      	cmp	r14, r13
  3a5843:      	jbe	0x3a749a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d1a>
  3a5849:      	add	r10, rcx
  3a584c:      	mov	qword ptr [rbp - 0x70], r10
  3a5850:      	mov	qword ptr [rbp - 0xb8], rsi
  3a5857:      	mov	rax, qword ptr [rsi + 0x40]
  3a585b:      	mov	qword ptr [rbp - 0x60], rax
  3a585f:      	mov	rsi, qword ptr [rdi + 0xb0]
  3a5866:      	add	r12, qword ptr [r11 + 0x78]
  3a586a:      	mov	r14d, r13d
  3a586d:      	mov	qword ptr [rbp - 0x38], r9
  3a5871:      	mov	qword ptr [rbp - 0x58], r12
  3a5875:      	mov	rcx, qword ptr [rbp - 0x40]
  3a5879:      	mov	qword ptr [rbp - 0x48], rsi
  3a587d:      	jmp	0x3a58a6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x126>
  3a587f:      	nop
  3a5880:      	shl	rdi, 0x4
  3a5884:      	mov	qword ptr [r12 + rdi], rax
  3a5888:      	mov	qword ptr [r12 + rdi + 0x8], rdx
  3a588d:      	inc	r14d
  3a5890:      	mov	r12d, r14d
  3a5893:      	mov	r13d, r12d
  3a5896:      	mov	r14d, r12d
  3a5899:      	cmp	rcx, r13
  3a589c:      	mov	r12, qword ptr [rbp - 0x58]
  3a58a0:      	jbe	0x3a749e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d1e>
  3a58a6:      	shl	r13, 0x5
  3a58aa:      	add	r13, rsi
  3a58ad:      	test	r15b, 0x1
  3a58b1:      	je	0x3a5913 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193>
  3a58b3:      	mov	rax, qword ptr [rbp - 0x68]
  3a58b7:      	mov	rsi, qword ptr [rax + 0x80]
  3a58be:      	mov	edi, dword ptr [r13 + 0x4]
  3a58c2:      	cmp	rsi, rdi
  3a58c5:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a58cb:      	mov	r12, qword ptr [rax + 0x78]
  3a58cf:      	imul	r15, rdi, 0xe0
  3a58d6:      	mov	eax, dword ptr [r12 + r15]
  3a58da:      	cmp	eax, 0x5
  3a58dd:      	setne	cl
  3a58e0:      	cmp	eax, 0xb
  3a58e3:      	setne	al
  3a58e6:      	and	al, cl
  3a58e8:      	movzx	ecx, al
  3a58eb:      	mov	rax, qword ptr [r11 + 0x150]
  3a58f2:      	sub	rax, rcx
  3a58f5:      	jb	0x3a7273 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af3>
  3a58fb:      	mov	qword ptr [r11 + 0x150], rax
  3a5902:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a5907:      	mov	r12, qword ptr [rbp - 0x58]
  3a590b:      	mov	rcx, qword ptr [rbp - 0x40]
  3a590f:      	mov	rsi, qword ptr [rbp - 0x48]
  3a5913:      	inc	rbx
  3a5916:      	test	bl, 0x3f
  3a5919:      	jne	0x3a599d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21d>
  3a591f:      	mov	rax, qword ptr [r11 + 0x160]
  3a5926:      	test	rax, rax
  3a5929:      	je	0x3a593e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1be>
  3a592b:      	test	r11, r11
  3a592e:      	je	0x3a5a73 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2f3>
  3a5934:      	mov	eax, dword ptr [rax]
  3a5936:      	test	eax, eax
  3a5938:      	je	0x3a5ec8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x748>
  3a593e:      	mov	rax, qword ptr [r11 + 0x8]
  3a5942:      	test	rax, rax
  3a5945:      	je	0x3a7508 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d88>
  3a594b:      	mov	rdi, qword ptr [r11 + 0x10]
  3a594f:      	mov	rcx, qword ptr [rax + 0x8]
  3a5953:      	test	byte ptr [rcx + 0x17], 0x20
  3a5957:      	je	0x3a5970 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f0>
  3a5959:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a595d:      	and	cl, 0x1f
  3a5960:      	cmp	cl, 0x16
  3a5963:      	je	0x3a5970 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f0>
  3a5965:      	mov	qword ptr [rbp - 0x98], rdi
  3a596c:      	mov	rdi, qword ptr [rbp - 0x78]
  3a5970:      	call	qword ptr [rax + 0x28]
  3a5973:      	mov	r8d, 0xffffffff
  3a5979:      	mov	r9, qword ptr [rbp - 0x38]
  3a597d:      	mov	rcx, qword ptr [rbp - 0x40]
  3a5981:      	mov	rsi, qword ptr [rbp - 0x48]
  3a5985:      	test	rax, rax
  3a5988:      	je	0x3a5a98 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x318>
  3a598e:      	cmp	qword ptr [rax + 0x8], 0x0
  3a5993:      	mov	r11, qword ptr [rbp - 0x50]
  3a5997:      	jne	0x3a72d6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a599d:      	cmp	dword ptr [r13 + 0x1c], 0x0
  3a59a2:      	je	0x3a5ac0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a59a8:      	test	r15b, 0x1
  3a59ac:      	je	0x3a5a00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x280>
  3a59ae:      	mov	r10d, dword ptr [r13]
  3a59b2:      	cmp	r10d, 0x20
  3a59b6:      	je	0x3a5ac0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a59bc:      	nop	dword ptr [rax]
  3a59c0:      	cmp	r10d, 0x41
  3a59c4:      	je	0x3a5ac0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a59ca:      	jmp	0x3a5ac4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x344>
  3a59cf:      	nop	word ptr cs:[rax + rax]
  3a59de:      	nop	word ptr cs:[rax + rax]
  3a59ed:      	nop	word ptr cs:[rax + rax]
  3a59fc:      	nop	dword ptr [rax]
  3a5a00:      	mov	r10d, dword ptr [r13]
  3a5a04:      	cmp	r10d, 0x20
  3a5a08:      	je	0x3a5a14 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x294>
  3a5a0a:      	cmp	r10d, 0x41
  3a5a0e:      	jne	0x3a5ac4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x344>
  3a5a14:      	mov	eax, ebx
  3a5a16:      	not	eax
  3a5a18:      	nop	dword ptr [rax + rax]
  3a5a20:      	test	al, 0x3f
  3a5a22:      	je	0x3a5ac0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a5a28:      	cmp	r10d, 0x41
  3a5a2c:      	jne	0x3a5a50 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d0>
  3a5a2e:      	mov	edi, dword ptr [r13 + 0x8]
  3a5a32:      	cmp	r9, rdi
  3a5a35:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a5a3b:      	mov	eax, dword ptr [r13 + 0x18]
  3a5a3f:      	shl	rdi, 0x4
  3a5a43:      	mov	qword ptr [r12 + rdi], rax
  3a5a47:      	mov	qword ptr [r12 + rdi + 0x8], 0x0
  3a5a50:      	inc	r14d
  3a5a53:      	cmp	rcx, r14
  3a5a56:      	jbe	0x3a734a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bca>
  3a5a5c:      	mov	rax, r14
  3a5a5f:      	shl	rax, 0x5
  3a5a63:      	lea	r13, [rsi + rax]
  3a5a67:      	inc	rbx
  3a5a6a:      	cmp	dword ptr [rsi + rax + 0x1c], 0x0
  3a5a6f:      	jne	0x3a5a00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x280>
  3a5a71:      	jmp	0x3a5ac0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x340>
  3a5a73:      	mov	edi, 0x1
  3a5a78:      	call	0x61ec00 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a5a7d:      	mov	rsi, qword ptr [rbp - 0x48]
  3a5a81:      	mov	rcx, qword ptr [rbp - 0x40]
  3a5a85:      	mov	r9, qword ptr [rbp - 0x38]
  3a5a89:      	mov	r8d, 0xffffffff
  3a5a8f:      	test	rax, rax
  3a5a92:      	jne	0x3a598e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20e>
  3a5a98:      	mov	r11, qword ptr [rbp - 0x50]
  3a5a9c:      	cmp	dword ptr [r13 + 0x1c], 0x0
  3a5aa1:      	jne	0x3a59a8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x228>
  3a5aa7:      	nop	word ptr cs:[rax + rax]
  3a5ab6:      	nop	word ptr cs:[rax + rax]
  3a5ac0:      	mov	r10d, dword ptr [r13]
  3a5ac4:      	lea	eax, [r10 - 0xe0]
  3a5acb:      	cmp	eax, 0xffff20
  3a5ad0:      	jae	0x3a5e00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x680>
  3a5ad6:      	mov	eax, r10d
  3a5ad9:      	shr	eax, 0x10
  3a5adc:      	nop	dword ptr [rax]
  3a5ae0:      	cmp	eax, 0xfb
  3a5ae5:      	je	0x3a5cd5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x555>
  3a5aeb:      	cmp	eax, 0xfc
  3a5af0:      	je	0x3a5b9b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x41b>
  3a5af6:      	nop	word ptr cs:[rax + rax]
  3a5b00:      	cmp	eax, 0xfd
  3a5b05:      	jne	0x3a5d72 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5f2>
  3a5b0b:      	mov	edi, dword ptr [r13 + 0xc]
  3a5b0f:      	cmp	rdi, r8
  3a5b12:      	je	0x3a5eda <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x75a>
  3a5b18:      	cmp	r9, rdi
  3a5b1b:      	mov	r15, qword ptr [rbp - 0x68]
  3a5b1f:      	nop
  3a5b20:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a5b26:      	shl	rdi, 0x4
  3a5b2a:      	mov	rcx, qword ptr [r12 + rdi]
  3a5b2e:      	mov	rdx, qword ptr [r12 + rdi + 0x8]
  3a5b33:      	mov	edi, dword ptr [r13 + 0x10]
  3a5b37:      	mov	esi, 0xffffffff
  3a5b3c:      	nop	dword ptr [rax]
  3a5b40:      	cmp	rdi, r8
  3a5b43:      	je	0x3a5ef4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x774>
  3a5b49:      	cmp	r9, rdi
  3a5b4c:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a5b52:      	shl	rdi, 0x4
  3a5b56:      	mov	rax, qword ptr [r12 + rdi]
  3a5b5a:      	mov	r8, qword ptr [r12 + rdi + 0x8]
  3a5b5f:      	mov	edi, dword ptr [r13 + 0x14]
  3a5b63:      	cmp	rdi, rsi
  3a5b66:      	je	0x3a5f09 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x789>
  3a5b6c:      	cmp	r9, rdi
  3a5b6f:      	jbe	0x3a735d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bdd>
  3a5b75:      	shl	rdi, 0x4
  3a5b79:      	mov	r9, qword ptr [r12 + rdi]
  3a5b7d:      	mov	r11, qword ptr [r12 + rdi + 0x8]
  3a5b82:      	mov	rsi, qword ptr [r15 + 0x80]
  3a5b89:      	mov	edi, dword ptr [r13 + 0x4]
  3a5b8d:      	cmp	rsi, rdi
  3a5b90:      	ja	0x3a5f29 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7a9>
  3a5b96:      	jmp	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a5b9b:      	add	r10d, 0xff03ffed
  3a5ba2:      	cmp	r10d, 0x4
  3a5ba6:      	jae	0x3a5dc5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x645>
  3a5bac:      	mov	rax, qword ptr [rbp - 0x68]
  3a5bb0:      	mov	rdx, qword ptr [rax + 0xd8]
  3a5bb7:      	mov	edi, dword ptr [r13 + 0x18]
  3a5bbb:      	mov	eax, dword ptr [r13 + 0x1c]
  3a5bbf:      	lea	ecx, [rdi + rax]
  3a5bc2:      	add	ecx, 0x2
  3a5bc5:      	cmp	rdx, rcx
  3a5bc8:      	jb	0x3a7414 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c94>
  3a5bce:      	cmp	edi, ecx
  3a5bd0:      	ja	0x3a7464 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ce4>
  3a5bd6:      	sub	rdx, rdi
  3a5bd9:      	lea	r8, [4*rdi]
  3a5be1:      	test	rdx, rdx
  3a5be4:      	mov	eax, 0x0
  3a5be9:      	cmovle	r8, rax
  3a5bed:      	cmp	ecx, edi
  3a5bef:      	je	0x3a745b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cdb>
  3a5bf5:      	mov	rax, qword ptr [rbp - 0x68]
  3a5bf9:      	add	r8, qword ptr [rax + 0xc8]
  3a5c00:      	mov	eax, dword ptr [r8]
  3a5c03:      	cmp	r9, rax
  3a5c06:      	jbe	0x3a7423 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ca3>
  3a5c0c:      	sub	rcx, rdi
  3a5c0f:      	cmp	ecx, 0x1
  3a5c12:      	je	0x3a744c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ccc>
  3a5c18:      	mov	edi, dword ptr [r8 + 0x4]
  3a5c1c:      	nop	dword ptr [rax]
  3a5c20:      	cmp	r9, rdi
  3a5c23:      	jbe	0x3a740b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c8b>
  3a5c29:      	shl	rax, 0x4
  3a5c2d:      	mov	r9, qword ptr [r12 + rax]
  3a5c31:      	shl	rdi, 0x4
  3a5c35:      	mov	rsi, qword ptr [r12 + rdi]
  3a5c39:      	mov	r10d, dword ptr [r13]
  3a5c3d:      	lea	eax, [r10 - 0xfc0013]
  3a5c44:      	cmp	eax, 0x2
  3a5c47:      	jae	0x3a6005 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x885>
  3a5c4d:      	cmp	ecx, 0x3
  3a5c50:      	jb	0x3a74c4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d44>
  3a5c56:      	mov	edi, dword ptr [r8 + 0x8]
  3a5c5a:      	nop	word ptr [rax + rax]
  3a5c60:      	cmp	qword ptr [rbp - 0x38], rdi
  3a5c64:      	jbe	0x3a740b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c8b>
  3a5c6a:      	cmp	ecx, 0x3
  3a5c6d:      	je	0x3a74b5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d35>
  3a5c73:      	mov	eax, dword ptr [r8 + 0xc]
  3a5c77:      	nop	word ptr [rax + rax]
  3a5c80:      	cmp	qword ptr [rbp - 0x38], rax
  3a5c84:      	jbe	0x3a74a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d29>
  3a5c8a:      	shl	rdi, 0x4
  3a5c8e:      	mov	rdi, qword ptr [r12 + rdi]
  3a5c92:      	shl	rax, 0x4
  3a5c96:      	mov	rdx, qword ptr [r12 + rax]
  3a5c9a:      	nop	word ptr [rax + rax]
  3a5ca0:      	cmp	r10d, 0xfc0013
  3a5ca7:      	jne	0x3a613e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9be>
  3a5cad:      	lea	rax, [rdi + r9]
  3a5cb1:      	mov	r10, rdi
  3a5cb4:      	and	r10, r9
  3a5cb7:      	or	rdi, r9
  3a5cba:      	mov	r9, rax
  3a5cbd:      	not	r9
  3a5cc0:      	and	r9, rdi
  3a5cc3:      	or	r9, r10
  3a5cc6:      	shr	r9, 0x3f
  3a5cca:      	add	rdx, rsi
  3a5ccd:      	add	rdx, r9
  3a5cd0:      	jmp	0x3a6163 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9e3>
  3a5cd5:      	add	r10d, -0x2
  3a5cd9:      	nop	dword ptr [rax]
  3a5ce0:      	cmp	r10w, 0x21
  3a5ce5:      	jae	0x3a5d85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x605>
  3a5ceb:      	lea	rax, [r8 + 0x1c003e10]
  3a5cf2:      	movzx	ecx, r10w
  3a5cf6:      	bt	rax, rcx
  3a5cfa:      	nop	word ptr [rax + rax]
  3a5d00:      	jae	0x3a5d85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x605>
  3a5d06:      	mov	edi, dword ptr [r13 + 0xc]
  3a5d0a:      	mov	eax, 0xffffffff
  3a5d0f:      	cmp	rdi, r8
  3a5d12:      	je	0x3a652c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdac>
  3a5d18:      	nop	dword ptr [rax + rax]
  3a5d20:      	cmp	r9, rdi
  3a5d23:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a5d29:      	shl	rdi, 0x4
  3a5d2d:      	mov	r8, qword ptr [r12 + rdi]
  3a5d31:      	mov	edi, dword ptr [r13 + 0x10]
  3a5d35:      	cmp	rdi, rax
  3a5d38:      	je	0x3a653c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdbc>
  3a5d3e:      	nop
  3a5d40:      	cmp	r9, rdi
  3a5d43:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a5d49:      	shl	rdi, 0x4
  3a5d4d:      	mov	rax, qword ptr [r12 + rdi]
  3a5d51:      	mov	rdx, qword ptr [r12 + rdi + 0x8]
  3a5d56:      	mov	r9d, 0x0
  3a5d5c:      	mov	r11d, 0x0
  3a5d62:      	cmp	r10w, 0xc
  3a5d67:      	je	0x3a6553 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd3>
  3a5d6d:      	jmp	0x3a6597 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe17>
  3a5d72:      	cmp	r10d, 0xe7
  3a5d79:      	jae	0x3a5e00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x680>
  3a5d7f:      	nop
  3a5d80:      	jmp	0x3a71f3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a73>
  3a5d85:      	mov	r12, qword ptr [rbp - 0x70]
  3a5d89:      	mov	dword ptr [r12 - 0x38], r14d
  3a5d8e:      	mov	rdi, r11
  3a5d91:      	call	0x3a7800 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a5d96:      	mov	rsi, qword ptr [rbp - 0x48]
  3a5d9a:      	mov	r9, qword ptr [rbp - 0x38]
  3a5d9e:      	mov	r8d, 0xffffffff
  3a5da4:      	mov	r11, qword ptr [rbp - 0x50]
  3a5da8:      	test	al, al
  3a5daa:      	mov	rcx, qword ptr [rbp - 0x40]
  3a5dae:      	jne	0x3a73e9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c69>
  3a5db4:      	inc	r14d
  3a5db7:      	mov	dword ptr [r12 - 0x38], r14d
  3a5dbc:      	nop	dword ptr [rax]
  3a5dc0:      	jmp	0x3a5890 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x110>
  3a5dc5:      	mov	r10d, dword ptr [r13]
  3a5dc9:      	cmp	r10d, 0xfc0008
  3a5dd0:      	jae	0x3a5e00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x680>
  3a5dd2:      	mov	edi, dword ptr [r13 + 0xc]
  3a5dd6:      	cmp	r9, rdi
  3a5dd9:      	jbe	0x3a735d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bdd>
  3a5ddf:      	shl	rdi, 0x4
  3a5de3:      	mov	rsi, qword ptr [r12 + rdi]
  3a5de7:      	mov	edi, r10d
  3a5dea:      	xor	edx, edx
  3a5dec:      	call	0x3d7380 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a5df1:      	mov	edi, edx
  3a5df3:      	xor	edx, edx
  3a5df5:      	jmp	0x3a6722 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfa2>
  3a5dfa:      	nop	word ptr [rax + rax]
  3a5e00:      	mov	edi, dword ptr [r13 + 0xc]
  3a5e04:      	cmp	rdi, r8
  3a5e07:      	je	0x3a5e80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x700>
  3a5e09:      	cmp	r9, rdi
  3a5e0c:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a5e12:      	shl	rdi, 0x4
  3a5e16:      	mov	rax, qword ptr [r12 + rdi]
  3a5e1a:      	mov	rdx, qword ptr [r12 + rdi + 0x8]
  3a5e1f:      	mov	edi, dword ptr [r13 + 0x10]
  3a5e23:      	cmp	rdi, r8
  3a5e26:      	je	0x3a5e8d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x70d>
  3a5e28:      	cmp	r9, rdi
  3a5e2b:      	jbe	0x3a735d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bdd>
  3a5e31:      	shl	rdi, 0x4
  3a5e35:      	mov	rcx, qword ptr [r12 + rdi]
  3a5e39:      	mov	r9, qword ptr [r12 + rdi + 0x8]
  3a5e3e:      	nop
  3a5e40:      	jmp	0x3a5e92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x712>
  3a5e42:      	nop	word ptr cs:[rax + rax]
  3a5e51:      	nop	word ptr cs:[rax + rax]
  3a5e60:      	nop	word ptr cs:[rax + rax]
  3a5e6f:      	nop	word ptr cs:[rax + rax]
  3a5e7e:      	nop
  3a5e80:      	xor	eax, eax
  3a5e82:      	xor	edx, edx
  3a5e84:      	mov	edi, dword ptr [r13 + 0x10]
  3a5e88:      	cmp	rdi, r8
  3a5e8b:      	jne	0x3a5e28 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6a8>
  3a5e8d:      	xor	ecx, ecx
  3a5e8f:      	xor	r9d, r9d
  3a5e92:      	mov	esi, r10d
  3a5e95:      	and	esi, 0xffffff
  3a5e9b:      	mov	edi, r10d
  3a5e9e:      	shr	edi, 0x10
  3a5ea1:      	and	edi, 0xff00
  3a5ea7:      	or	edi, esi
  3a5ea9:      	lea	esi, [rdi - 0x45]
  3a5eac:      	cmp	esi, 0x39
  3a5eaf:      	ja	0x3a61df <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa5f>
  3a5eb5:      	jmp	qword ptr [8*rsi + 0x2033d8]
  3a5ebc:      	xor	ecx, ecx
  3a5ebe:      	test	eax, eax
  3a5ec0:      	sete	cl
  3a5ec3:      	jmp	0x3a650d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd8d>
  3a5ec8:      	xor	eax, eax
  3a5eca:      	xor	edx, edx
  3a5ecc:      	test	rax, rax
  3a5ecf:      	jne	0x3a598e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20e>
  3a5ed5:      	jmp	0x3a5a98 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x318>
  3a5eda:      	xor	edx, edx
  3a5edc:      	xor	ecx, ecx
  3a5ede:      	mov	r15, qword ptr [rbp - 0x68]
  3a5ee2:      	mov	edi, dword ptr [r13 + 0x10]
  3a5ee6:      	mov	esi, 0xffffffff
  3a5eeb:      	cmp	rdi, r8
  3a5eee:      	jne	0x3a5b49 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3c9>
  3a5ef4:      	xor	eax, eax
  3a5ef6:      	xor	r8d, r8d
  3a5ef9:      	mov	edi, dword ptr [r13 + 0x14]
  3a5efd:      	nop	dword ptr [rax]
  3a5f00:      	cmp	rdi, rsi
  3a5f03:      	jne	0x3a5b6c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3ec>
  3a5f09:      	xor	r9d, r9d
  3a5f0c:      	xor	r11d, r11d
  3a5f0f:      	mov	rsi, qword ptr [r15 + 0x80]
  3a5f16:      	mov	edi, dword ptr [r13 + 0x4]
  3a5f1a:      	nop	word ptr [rax + rax]
  3a5f20:      	cmp	rsi, rdi
  3a5f23:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a5f29:      	imul	rdi, rdi, 0xe0
  3a5f30:      	add	rdi, qword ptr [r15 + 0x78]
  3a5f34:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a5f3a:      	movzx	esi, r10w
  3a5f3e:      	cmp	esi, 0xb
  3a5f41:      	setb	r15b
  3a5f45:      	mov	esi, r10d
  3a5f48:      	and	esi, 0xfffc
  3a5f4e:      	cmp	esi, 0x54
  3a5f51:      	sete	r12b
  3a5f55:      	or	r12b, r15b
  3a5f58:      	jne	0x3a5f6c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7ec>
  3a5f5a:      	movzx	r15d, r10w
  3a5f5e:      	lea	r12d, [r15 - 0x5c]
  3a5f62:      	cmp	r12d, 0x2
  3a5f66:      	jae	0x3a6035 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8b5>
  3a5f6c:      	mov	r9, qword ptr [rbp - 0x60]
  3a5f70:      	mov	rsi, qword ptr [r9 + 0x78]
  3a5f74:      	mov	edx, dword ptr [rdi + 0x18]
  3a5f77:      	nop	word ptr [rax + rax]
  3a5f80:      	cmp	rsi, rdx
  3a5f83:      	jbe	0x3a7403 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c83>
  3a5f89:      	shl	rdx, 0x3
  3a5f8d:      	add	rdx, qword ptr [r9 + 0x70]
  3a5f91:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a5f96:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a5f9c:      	mov	rdx, qword ptr [rdx]
  3a5f9f:      	nop
  3a5fa0:      	test	rdx, rdx
  3a5fa3:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a5fa9:      	mov	rsi, qword ptr [rdx + 0x10]
  3a5fad:      	mov	edx, dword ptr [rdi + 0x14]
  3a5fb0:      	mov	r11d, dword ptr [rdi + 0x1c]
  3a5fb4:      	shl	r11, 0x20
  3a5fb8:      	or	r11, rdx
  3a5fbb:      	movzx	edx, byte ptr [rdi + 0x40]
  3a5fbf:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a5fc4:      	je	0x3a5fd1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x851>
  3a5fc6:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a5fcb:      	je	0x3a60e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x965>
  3a5fd1:      	mov	qword ptr [rbp - 0xb0], rax
  3a5fd8:      	mov	qword ptr [rbp - 0xa8], r8
  3a5fdf:      	movups	xmm0, xmmword ptr [rbp - 0xb0]
  3a5fe6:      	movups	xmmword ptr [rsp], xmm0
  3a5fea:      	movzx	r9d, dl
  3a5fee:      	mov	r12, qword ptr [rbp - 0x78]
  3a5ff2:      	mov	rdi, r12
  3a5ff5:      	mov	edx, r10d
  3a5ff8:      	mov	r8, r11
  3a5ffb:      	call	0x3cc840 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a6000:      	jmp	0x3a6114 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x994>
  3a6005:      	mov	rax, rsi
  3a6008:      	mul	r9
  3a600b:      	cmp	dword ptr [r13], 0xfc0015
  3a6013:      	jne	0x3a6163 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9e3>
  3a6019:      	mov	rdi, r9
  3a601c:      	sar	rdi, 0x3f
  3a6020:      	and	rdi, rsi
  3a6023:      	sar	rsi, 0x3f
  3a6027:      	and	rsi, r9
  3a602a:      	add	rsi, rdi
  3a602d:      	sub	rdx, rsi
  3a6030:      	jmp	0x3a6163 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9e3>
  3a6035:      	cmp	r15d, 0xb
  3a6039:      	je	0x3a6049 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8c9>
  3a603b:      	nop	dword ptr [rax + rax]
  3a6040:      	cmp	esi, 0x58
  3a6043:      	jne	0x3a668b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0b>
  3a6049:      	mov	r9, qword ptr [rbp - 0x60]
  3a604d:      	mov	rsi, qword ptr [r9 + 0x78]
  3a6051:      	mov	edx, dword ptr [rdi + 0x18]
  3a6054:      	cmp	rsi, rdx
  3a6057:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a605c:      	mov	r12, qword ptr [rbp - 0x58]
  3a6060:      	jbe	0x3a7403 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c83>
  3a6066:      	shl	rdx, 0x3
  3a606a:      	add	rdx, qword ptr [r9 + 0x70]
  3a606e:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6074:      	mov	rdx, qword ptr [rdx]
  3a6077:      	nop	word ptr [rax + rax]
  3a6080:      	test	rdx, rdx
  3a6083:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6089:      	mov	rdx, qword ptr [rdx + 0x10]
  3a608d:      	mov	esi, dword ptr [rdi + 0x14]
  3a6090:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a6094:      	shl	r9, 0x20
  3a6098:      	or	r9, rsi
  3a609b:      	movzx	esi, byte ptr [rdi + 0x40]
  3a609f:      	cmp	qword ptr [rdx + 0x40], 0x0
  3a60a4:      	je	0x3a60b1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x931>
  3a60a6:      	cmp	qword ptr [rdx + 0x60], 0x0
  3a60ab:      	je	0x3a66ee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf6e>
  3a60b1:      	mov	qword ptr [rbp - 0x98], rax
  3a60b8:      	mov	qword ptr [rbp - 0x90], r8
  3a60bf:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a60c6:      	movups	xmmword ptr [rsp], xmm0
  3a60ca:      	movzx	r8d, sil
  3a60ce:      	mov	rdi, rdx
  3a60d1:      	mov	esi, r10d
  3a60d4:      	mov	rdx, rcx
  3a60d7:      	mov	rcx, r9
  3a60da:      	call	0x3cdea0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a60df:      	nop
  3a60e0:      	jmp	0x3a671c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf9c>
  3a60e5:      	mov	qword ptr [rbp - 0xb0], rax
  3a60ec:      	mov	qword ptr [rbp - 0xa8], r8
  3a60f3:      	movups	xmm0, xmmword ptr [rbp - 0xb0]
  3a60fa:      	movups	xmmword ptr [rsp], xmm0
  3a60fe:      	movzx	r9d, dl
  3a6102:      	mov	r12, qword ptr [rbp - 0x78]
  3a6106:      	mov	rdi, r12
  3a6109:      	mov	edx, r10d
  3a610c:      	mov	r8, r11
  3a610f:      	call	0x3cc9a0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a6114:      	mov	rax, qword ptr [r12]
  3a6118:      	mov	rdx, qword ptr [r12 + 0x8]
  3a611d:      	mov	qword ptr [rbp - 0x78], r12
  3a6121:      	movzx	edi, byte ptr [r12 + 0x10]
  3a6127:      	mov	r11, qword ptr [rbp - 0x50]
  3a612b:      	mov	r8d, 0xffffffff
  3a6131:      	mov	r9, qword ptr [rbp - 0x38]
  3a6135:      	mov	r12, qword ptr [rbp - 0x58]
  3a6139:      	jmp	0x3a6730 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfb0>
  3a613e:      	mov	rax, r9
  3a6141:      	sub	rax, rdi
  3a6144:      	not	r9
  3a6147:      	mov	r10, rdi
  3a614a:      	and	r10, r9
  3a614d:      	xor	r9, rdi
  3a6150:      	and	r9, rax
  3a6153:      	or	r9, r10
  3a6156:      	shr	r9, 0x3f
  3a615a:      	add	rdx, r9
  3a615d:      	sub	rsi, rdx
  3a6160:      	mov	rdx, rsi
  3a6163:      	mov	r9, qword ptr [rbp - 0x38]
  3a6167:      	mov	edi, dword ptr [r13 + 0x1c]
  3a616b:      	cmp	ecx, edi
  3a616d:      	jbe	0x3a742e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cae>
  3a6173:      	lea	esi, [rdi + 0x1]
  3a6176:      	cmp	ecx, esi
  3a6178:      	mov	r10, qword ptr [rbp - 0x70]
  3a617c:      	nop	dword ptr [rax]
  3a6180:      	jbe	0x3a7436 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cb6>
  3a6186:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a618a:      	cmp	r9, rcx
  3a618d:      	jbe	0x3a7441 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cc1>
  3a6193:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a6197:      	shl	rcx, 0x4
  3a619b:      	mov	qword ptr [r12 + rcx], rax
  3a619f:      	mov	qword ptr [r12 + rcx + 0x8], 0x0
  3a61a8:      	cmp	r9, rdi
  3a61ab:      	jbe	0x3a7444 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cc4>
  3a61b1:      	shl	rdi, 0x4
  3a61b5:      	mov	qword ptr [r12 + rdi], rdx
  3a61b9:      	mov	qword ptr [r12 + rdi + 0x8], 0x0
  3a61c2:      	inc	r14d
  3a61c5:      	mov	dword ptr [r10 - 0x38], r14d
  3a61c9:      	mov	r12d, r14d
  3a61cc:      	mov	r8d, 0xffffffff
  3a61d2:      	mov	rcx, qword ptr [rbp - 0x40]
  3a61d6:      	mov	rsi, qword ptr [rbp - 0x48]
  3a61da:      	jmp	0x3a5893 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x113>
  3a61df:      	lea	esi, [rdi - 0x5b]
  3a61e2:      	cmp	esi, 0x69
  3a61e5:      	ja	0x3a67cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x104d>
  3a61eb:      	jmp	qword ptr [8*rsi + 0x2035a8]
  3a61f2:      	mov	rsi, rax
  3a61f5:      	mov	rdx, rcx
  3a61f8:      	call	0x3d7380 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a61fd:      	mov	r8d, 0xffffffff
  3a6203:      	mov	r11, qword ptr [rbp - 0x50]
  3a6207:      	mov	edi, edx
  3a6209:      	xor	edx, edx
  3a620b:      	jmp	0x3a672c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfac>
  3a6210:      	add	rax, rcx
  3a6213:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6218:      	mov	rcx, rax
  3a621b:      	shr	rcx
  3a621e:      	movabs	rdx, 0x5555555555555555
  3a6228:      	and	rcx, rdx
  3a622b:      	and	rax, rdx
  3a622e:      	add	rax, rcx
  3a6231:      	mov	rcx, rax
  3a6234:      	shr	rcx, 0x2
  3a6238:      	movabs	rdx, 0x3333333333333333
  3a6242:      	and	rcx, rdx
  3a6245:      	and	rax, rdx
  3a6248:      	add	rax, rcx
  3a624b:      	mov	rcx, rax
  3a624e:      	shr	rcx, 0x4
  3a6252:      	add	rcx, rax
  3a6255:      	movabs	rax, 0xf0f0f0f0f0f0f0f
  3a625f:      	and	rax, rcx
  3a6262:      	mov	rcx, rax
  3a6265:      	shr	rcx, 0x8
  3a6269:      	add	rcx, rax
  3a626c:      	mov	rdx, rcx
  3a626f:      	shr	rdx, 0x10
  3a6273:      	add	rdx, rcx
  3a6276:      	mov	rax, rdx
  3a6279:      	shr	rax, 0x20
  3a627d:      	add	eax, edx
  3a627f:      	and	eax, 0x7f
  3a6282:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6287:      	xor	ecx, ecx
  3a6289:      	mov	rdx, rax
  3a628c:      	shr	rdx, 0x20
  3a6290:      	cmove	rdx, rax
  3a6294:      	setne	cl
  3a6297:      	shl	ecx, 0x5
  3a629a:      	mov	rax, rdx
  3a629d:      	shr	rax, 0x10
  3a62a1:      	lea	rsi, [rcx + 0x10]
  3a62a5:      	cmp	rdx, 0x10000
  3a62ac:      	cmovb	rax, rdx
  3a62b0:      	cmovb	rsi, rcx
  3a62b4:      	mov	ecx, eax
  3a62b6:      	shr	ecx, 0x8
  3a62b9:      	lea	rdx, [rsi + 0x8]
  3a62bd:      	cmp	rax, 0x100
  3a62c3:      	cmovb	rcx, rax
  3a62c7:      	cmovb	rdx, rsi
  3a62cb:      	movzx	eax, cl
  3a62ce:      	movzx	ecx, byte ptr [rax + 0x328461]
  3a62d5:      	add	rcx, rdx
  3a62d8:      	mov	eax, 0x40
  3a62dd:      	sub	rax, rcx
  3a62e0:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a62e5:      	xor	esi, esi
  3a62e7:      	cmp	rax, rcx
  3a62ea:      	setle	sil
  3a62ee:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a62f3:      	xor	esi, esi
  3a62f5:      	cmp	rax, rcx
  3a62f8:      	setae	sil
  3a62fc:      	nop	dword ptr [rax]
  3a6300:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a6305:      	add	ecx, eax
  3a6307:      	jmp	0x3a650d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd8d>
  3a630c:      	test	eax, eax
  3a630e:      	je	0x3a66b3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf33>
  3a6314:      	tzcnt	eax, eax
  3a6318:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a631d:      	shr	eax, cl
  3a631f:      	nop
  3a6320:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6325:      	and	eax, ecx
  3a6327:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a632c:      	xor	esi, esi
  3a632e:      	cmp	rax, rcx
  3a6331:      	sete	sil
  3a6335:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a633a:      	sub	rax, rcx
  3a633d:      	nop	dword ptr [rax]
  3a6340:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6345:      	xor	ecx, ecx
  3a6347:      	test	rax, rax
  3a634a:      	sete	cl
  3a634d:      	jmp	0x3a650d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd8d>
  3a6352:      	mov	ecx, eax
  3a6354:      	shr	ecx, 0x10
  3a6357:      	xor	edx, edx
  3a6359:      	cmp	eax, 0x10000
  3a635e:      	setb	dl
  3a6361:      	cmovb	ecx, eax
  3a6364:      	shl	edx, 0x4
  3a6367:      	mov	esi, ecx
  3a6369:      	shr	esi, 0x8
  3a636c:      	xor	eax, eax
  3a636e:      	cmp	ecx, 0x100
  3a6374:      	setb	al
  3a6377:      	lea	eax, [rdx + 8*rax]
  3a637a:      	cmovb	esi, ecx
  3a637d:      	movzx	ecx, sil
  3a6381:      	movzx	ecx, byte ptr [rcx + 0x328461]
  3a6388:      	sub	rax, rcx
  3a638b:      	add	rax, 0x8
  3a638f:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6394:      	xor	esi, esi
  3a6396:      	cmp	rax, rcx
  3a6399:      	setbe	sil
  3a639d:      	nop	dword ptr [rax]
  3a63a0:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a63a5:      	or	ecx, eax
  3a63a7:      	jmp	0x3a650d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd8d>
  3a63ac:      	shl	eax, cl
  3a63ae:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a63b3:      	xor	esi, esi
  3a63b5:      	cmp	eax, ecx
  3a63b7:      	seta	sil
  3a63bb:      	nop	dword ptr [rax + rax]
  3a63c0:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a63c5:      	xor	esi, esi
  3a63c7:      	cmp	eax, ecx
  3a63c9:      	setle	sil
  3a63cd:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a63d2:      	xor	esi, esi
  3a63d4:      	cmp	eax, ecx
  3a63d6:      	setb	sil
  3a63da:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a63df:      	mov	ecx, eax
  3a63e1:      	shr	ecx, 0x10
  3a63e4:      	movzx	ecx, cl
  3a63e7:      	movzx	ecx, byte ptr [rcx + 0x328561]
  3a63ee:      	movzx	edx, ah
  3a63f1:      	movzx	esi, al
  3a63f4:      	shr	eax, 0x18
  3a63f7:      	add	cl, byte ptr [rax + 0x328561]
  3a63fd:      	add	cl, byte ptr [rdx + 0x328561]
  3a6403:      	add	cl, byte ptr [rsi + 0x328561]
  3a6409:      	movzx	eax, cl
  3a640c:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6411:      	imul	rax, rcx
  3a6415:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a641a:      	xor	esi, esi
  3a641c:      	cmp	eax, ecx
  3a641e:      	setne	sil
  3a6422:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a6427:      	xor	esi, esi
  3a6429:      	cmp	eax, ecx
  3a642b:      	sete	sil
  3a642f:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a6434:      	xor	esi, esi
  3a6436:      	cmp	eax, ecx
  3a6438:      	setge	sil
  3a643c:      	nop	dword ptr [rax]
  3a6440:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a6445:      	xor	esi, esi
  3a6447:      	cmp	eax, ecx
  3a6449:      	setbe	sil
  3a644d:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a6452:      	xor	esi, esi
  3a6454:      	cmp	rax, rcx
  3a6457:      	setl	sil
  3a645b:      	nop	dword ptr [rax + rax]
  3a6460:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a6465:      	xor	esi, esi
  3a6467:      	cmp	rax, rcx
  3a646a:      	setb	sil
  3a646e:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a6473:      	xor	ecx, eax
  3a6475:      	jmp	0x3a650d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd8d>
  3a647a:      	xor	esi, esi
  3a647c:      	cmp	eax, ecx
  3a647e:      	setg	sil
  3a6482:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a6487:      	mov	ecx, 0x40
  3a648c:      	tzcnt	rcx, rax
  3a6491:      	jmp	0x3a650d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd8d>
  3a6496:      	xor	esi, esi
  3a6498:      	cmp	eax, ecx
  3a649a:      	setl	sil
  3a649e:      	nop
  3a64a0:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a64a2:      	xor	esi, esi
  3a64a4:      	cmp	rax, rcx
  3a64a7:      	seta	sil
  3a64ab:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a64ad:      	rol	eax, cl
  3a64af:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a64b4:      	xor	esi, esi
  3a64b6:      	cmp	eax, ecx
  3a64b8:      	setae	sil
  3a64bc:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a64be:      	xor	esi, esi
  3a64c0:      	cmp	rax, rcx
  3a64c3:      	setg	sil
  3a64c7:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a64c9:      	xor	esi, esi
  3a64cb:      	cmp	rax, rcx
  3a64ce:      	setne	sil
  3a64d2:      	jmp	0x3a6500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd80>
  3a64d4:      	xor	esi, esi
  3a64d6:      	cmp	rax, rcx
  3a64d9:      	setge	sil
  3a64dd:      	nop	word ptr cs:[rax + rax]
  3a64ec:      	nop	word ptr cs:[rax + rax]
  3a64fb:      	nop	dword ptr [rax + rax]
  3a6500:      	xor	edx, edx
  3a6502:      	mov	rax, rsi
  3a6505:      	jmp	0x3a66c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf42>
  3a650a:      	imul	ecx, eax
  3a650d:      	xor	edx, edx
  3a650f:      	mov	rax, rcx
  3a6512:      	jmp	0x3a66c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf42>
  3a6517:      	sub	eax, ecx
  3a6519:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a651e:      	ror	eax, cl
  3a6520:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6525:      	sar	eax, cl
  3a6527:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a652c:      	xor	r8d, r8d
  3a652f:      	mov	edi, dword ptr [r13 + 0x10]
  3a6533:      	cmp	rdi, rax
  3a6536:      	jne	0x3a5d3e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5be>
  3a653c:      	xor	eax, eax
  3a653e:      	xor	edx, edx
  3a6540:      	mov	r9d, 0x0
  3a6546:      	mov	r11d, 0x0
  3a654c:      	cmp	r10w, 0xc
  3a6551:      	jne	0x3a6597 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe17>
  3a6553:      	mov	rcx, qword ptr [rbp - 0x68]
  3a6557:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a655e:      	mov	edi, dword ptr [r13 + 0x18]
  3a6562:      	add	edi, 0x2
  3a6565:      	cmp	rsi, rdi
  3a6568:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a656e:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a6575:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a6578:      	nop	dword ptr [rax + rax]
  3a6580:      	cmp	qword ptr [rbp - 0x38], rdi
  3a6584:      	jbe	0x3a735d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bdd>
  3a658a:      	shl	rdi, 0x4
  3a658e:      	mov	r9, qword ptr [r12 + rdi]
  3a6592:      	mov	r11, qword ptr [r12 + rdi + 0x8]
  3a6597:      	cmp	qword ptr [rbp - 0x50], 0x0
  3a659c:      	nop	dword ptr [rax]
  3a65a0:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a65a6:      	mov	r10, qword ptr [rbp - 0x68]
  3a65aa:      	mov	rsi, qword ptr [r10 + 0x80]
  3a65b1:      	mov	edi, dword ptr [r13 + 0x4]
  3a65b5:      	cmp	rsi, rdi
  3a65b8:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a65be:      	imul	rcx, rdi, 0xe0
  3a65c5:      	add	rcx, qword ptr [r10 + 0x78]
  3a65c9:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a65cf:      	mov	rsi, qword ptr [rbp - 0x50]
  3a65d3:      	mov	rsi, qword ptr [rsi]
  3a65d6:      	mov	qword ptr [rbp - 0xd8], rax
  3a65dd:      	mov	qword ptr [rbp - 0xd0], rdx
  3a65e4:      	mov	qword ptr [rbp - 0xc8], r9
  3a65eb:      	mov	qword ptr [rbp - 0xc0], r11
  3a65f2:      	movups	xmm0, xmmword ptr [rbp - 0xc8]
  3a65f9:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a65fe:      	movups	xmm0, xmmword ptr [rbp - 0xd8]
  3a6605:      	movups	xmmword ptr [rsp], xmm0
  3a6609:      	lea	rdi, [rbp - 0x160]
  3a6610:      	mov	rdx, qword ptr [rbp - 0x60]
  3a6614:      	call	0x3bb200 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a6619:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a6620:      	mov	rax, qword ptr [rbp - 0x150]
  3a6627:      	test	rax, rax
  3a662a:      	je	0x3a664a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xeca>
  3a662c:      	cmp	qword ptr [rax + 0x8], 0x0
  3a6631:      	mov	r8d, 0xffffffff
  3a6637:      	mov	r9, qword ptr [rbp - 0x38]
  3a663b:      	mov	rcx, qword ptr [rbp - 0x40]
  3a663f:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6643:      	je	0x3a665c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xedc>
  3a6645:      	jmp	0x3a72e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b65>
  3a664a:      	mov	r8d, 0xffffffff
  3a6650:      	mov	r9, qword ptr [rbp - 0x38]
  3a6654:      	mov	rcx, qword ptr [rbp - 0x40]
  3a6658:      	mov	rsi, qword ptr [rbp - 0x48]
  3a665c:      	mov	edi, dword ptr [r13 + 0x8]
  3a6660:      	cmp	rdi, r8
  3a6663:      	mov	r11, qword ptr [rbp - 0x50]
  3a6667:      	mov	r12, qword ptr [rbp - 0x70]
  3a666b:      	je	0x3a5db4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x634>
  3a6671:      	cmp	r9, rdi
  3a6674:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a667a:      	shl	rdi, 0x4
  3a667e:      	mov	rax, qword ptr [rbp - 0x58]
  3a6682:      	movups	xmmword ptr [rax + rdi], xmm0
  3a6686:      	jmp	0x3a5db4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x634>
  3a668b:      	mov	qword ptr [rsp], r11
  3a668f:      	mov	rsi, rcx
  3a6692:      	mov	rcx, rax
  3a6695:      	call	0x3d5340 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a669a:      	mov	r11, qword ptr [rbp - 0x50]
  3a669e:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a66a3:      	mov	r8d, 0xffffffff
  3a66a9:      	mov	r9, qword ptr [rbp - 0x38]
  3a66ad:      	mov	r12, qword ptr [rbp - 0x58]
  3a66b1:      	jmp	0x3a66c6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf46>
  3a66b3:      	mov	eax, 0x20
  3a66b8:      	nop	dword ptr [rax + rax]
  3a66c0:      	xor	edx, edx
  3a66c2:      	mov	r9, qword ptr [rbp - 0x38]
  3a66c6:      	mov	rcx, qword ptr [rbp - 0x40]
  3a66ca:      	mov	rsi, qword ptr [rbp - 0x48]
  3a66ce:      	mov	edi, dword ptr [r13 + 0x8]
  3a66d2:      	cmp	rdi, r8
  3a66d5:      	je	0x3a588d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10d>
  3a66db:      	nop	dword ptr [rax + rax]
  3a66e0:      	cmp	r9, rdi
  3a66e3:      	ja	0x3a5880 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a66e9:      	jmp	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a66ee:      	mov	qword ptr [rbp - 0x98], rax
  3a66f5:      	mov	qword ptr [rbp - 0x90], r8
  3a66fc:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a6703:      	movups	xmmword ptr [rsp], xmm0
  3a6707:      	movzx	r8d, sil
  3a670b:      	mov	rdi, rdx
  3a670e:      	mov	esi, r10d
  3a6711:      	mov	rdx, rcx
  3a6714:      	mov	rcx, r9
  3a6717:      	call	0x3ce1a0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a671c:      	mov	edi, eax
  3a671e:      	xor	edx, edx
  3a6720:      	xor	eax, eax
  3a6722:      	mov	r11, qword ptr [rbp - 0x50]
  3a6726:      	mov	r8d, 0xffffffff
  3a672c:      	mov	r9, qword ptr [rbp - 0x38]
  3a6730:      	mov	rcx, qword ptr [rbp - 0x40]
  3a6734:      	test	dil, dil
  3a6737:      	mov	rsi, qword ptr [rbp - 0x48]
  3a673b:      	je	0x3a66ce <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf4e>
  3a673d:      	nop	dword ptr [rax]
  3a6740:      	jmp	0x3a736e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bee>
  3a6745:      	or	rax, rcx
  3a6748:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a674d:      	mov	eax, eax
  3a674f:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6754:      	shr	rax, cl
  3a6757:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a675c:      	and	rax, rcx
  3a675f:      	nop
  3a6760:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6765:      	rol	rax, cl
  3a6768:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a676d:      	ror	rax, cl
  3a6770:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6775:      	mov	eax, eax
  3a6777:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a677c:      	movsx	rax, ax
  3a6780:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6785:      	cdqe
  3a6787:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a678c:      	cdqe
  3a678e:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6793:      	sar	rax, cl
  3a6796:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a679b:      	movsx	rax, ax
  3a679f:      	mov	eax, eax
  3a67a1:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a67a6:      	shl	rax, cl
  3a67a9:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a67ae:      	movsx	rax, al
  3a67b2:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a67b7:      	movsx	rax, al
  3a67bb:      	mov	eax, eax
  3a67bd:      	nop	dword ptr [rax]
  3a67c0:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a67c5:      	xor	rax, rcx
  3a67c8:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a67cd:      	lea	esi, [rdi - 0x1]
  3a67d0:      	cmp	esi, 0x108
  3a67d6:      	ja	0x3a69d2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1252>
  3a67dc:      	nop	dword ptr [rax]
  3a67e0:      	jmp	qword ptr [8*rsi + 0x2038f8]
  3a67e7:      	mov	rdx, qword ptr [rbp - 0x68]
  3a67eb:      	mov	rsi, qword ptr [rdx + 0x80]
  3a67f2:      	mov	ecx, dword ptr [r13 + 0x4]
  3a67f6:      	cmp	rsi, rcx
  3a67f9:      	jbe	0x3a7517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d97>
  3a67ff:      	imul	r8, rcx, 0xe0
  3a6806:      	add	r8, qword ptr [rdx + 0x78]
  3a680a:      	mov	r10, qword ptr [rbp - 0x78]
  3a680e:      	mov	rdx, qword ptr [rbp - 0x60]
  3a6812:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6818:      	mov	rsi, qword ptr [rdx + 0x78]
  3a681c:      	mov	ecx, dword ptr [r8 + 0x18]
  3a6820:      	cmp	rsi, rcx
  3a6823:      	jbe	0x3a7517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d97>
  3a6829:      	shl	rcx, 0x3
  3a682d:      	add	rcx, qword ptr [rdx + 0x70]
  3a6831:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6837:      	mov	rcx, qword ptr [rcx]
  3a683a:      	nop	word ptr [rax + rax]
  3a6840:      	test	rcx, rcx
  3a6843:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6849:      	mov	rdx, rdi
  3a684c:      	mov	rsi, qword ptr [rcx + 0x10]
  3a6850:      	mov	ecx, dword ptr [r8 + 0x14]
  3a6854:      	mov	r8d, dword ptr [r8 + 0x1c]
  3a6858:      	shl	r8, 0x20
  3a685c:      	or	r8, rcx
  3a685f:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a6864:      	je	0x3a686d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10ed>
  3a6866:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a686b:      	je	0x3a68bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x113d>
  3a686d:      	xorps	xmm0, xmm0
  3a6870:      	movaps	xmmword ptr [rbp - 0xb0], xmm0
  3a6877:      	movups	xmmword ptr [rsp], xmm0
  3a687b:      	mov	rdi, r10
  3a687e:      	mov	rcx, rax
  3a6881:      	xor	r9d, r9d
  3a6884:      	mov	r12, r10
  3a6887:      	call	0x3cc840 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a688c:      	jmp	0x3a6114 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x994>
  3a6891:      	movzx	ecx, r10b
  3a6895:      	xor	edi, edi
  3a6897:      	add	ecx, -0x2
  3a689a:      	nop	word ptr [rax + rax]
  3a68a0:      	cmp	ecx, 0x7
  3a68a3:      	ja	0x3a6943 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11c3>
  3a68a9:      	jmp	qword ptr [8*rcx + 0x204140]
  3a68b0:      	xor	edi, edi
  3a68b2:      	test	eax, eax
  3a68b4:      	sete	dil
  3a68b8:      	jmp	0x3a6943 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11c3>
  3a68bd:      	xorps	xmm0, xmm0
  3a68c0:      	movaps	xmmword ptr [rbp - 0xb0], xmm0
  3a68c7:      	movups	xmmword ptr [rsp], xmm0
  3a68cb:      	mov	rdi, r10
  3a68ce:      	mov	rcx, rax
  3a68d1:      	xor	r9d, r9d
  3a68d4:      	mov	r12, r10
  3a68d7:      	jmp	0x3a610f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x98f>
  3a68dc:      	mov	edi, dword ptr [r13 + 0x18]
  3a68e0:      	mov	r9, qword ptr [rbp - 0x38]
  3a68e4:      	cmp	r9, rdi
  3a68e7:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a68ed:      	shl	rdi, 0x4
  3a68f1:      	mov	qword ptr [r12 + rdi], rax
  3a68f5:      	mov	qword ptr [r12 + rdi + 0x8], rdx
  3a68fa:      	jmp	0x3a66c6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf46>
  3a68ff:      	mov	edi, dword ptr [r13 + 0x14]
  3a6903:      	cmp	rdi, r8
  3a6906:      	je	0x3a6c1a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x149a>
  3a690c:      	cmp	qword ptr [rbp - 0x38], rdi
  3a6910:      	jbe	0x3a735d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bdd>
  3a6916:      	shl	rdi, 0x4
  3a691a:      	mov	rsi, qword ptr [r12 + rdi]
  3a691e:      	test	esi, esi
  3a6920:      	cmove	rax, rcx
  3a6924:      	cmove	rdx, r9
  3a6928:      	jmp	0x3a66c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf42>
  3a692d:      	mov	edi, dword ptr [r13 + 0x1c]
  3a6931:      	dec	edi
  3a6933:      	cmp	edi, eax
  3a6935:      	cmovae	edi, eax
  3a6938:      	jmp	0x3a6943 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11c3>
  3a693a:      	xor	edi, edi
  3a693c:      	test	rax, rax
  3a693f:      	setne	dil
  3a6943:      	mov	rax, qword ptr [rbp - 0x68]
  3a6947:      	mov	rsi, qword ptr [rax + 0xe8]
  3a694e:      	add	edi, dword ptr [r13 + 0x18]
  3a6952:      	cmp	rsi, rdi
  3a6955:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a695b:      	mov	rax, qword ptr [rax + 0xe0]
  3a6962:      	lea	rcx, [rdi + 2*rdi]
  3a6966:      	mov	r12d, dword ptr [rax + 4*rcx]
  3a696a:      	mov	r13d, dword ptr [rax + 4*rcx + 0x4]
  3a696f:      	mov	esi, dword ptr [rax + 4*rcx + 0x8]
  3a6973:      	cmp	r12d, r14d
  3a6976:      	mov	r9, qword ptr [rbp - 0x38]
  3a697a:      	nop	word ptr [rax + rax]
  3a6980:      	ja	0x3a6c87 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a6986:      	mov	rax, qword ptr [r11 + 0x160]
  3a698d:      	test	rax, rax
  3a6990:      	je	0x3a69a8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1228>
  3a6992:      	test	r11, r11
  3a6995:      	je	0x3a6c25 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14a5>
  3a699b:      	mov	eax, dword ptr [rax]
  3a699d:      	nop	dword ptr [rax]
  3a69a0:      	test	eax, eax
  3a69a2:      	je	0x3a6c5d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dd>
  3a69a8:      	mov	rax, qword ptr [r11 + 0x8]
  3a69ac:      	test	rax, rax
  3a69af:      	je	0x3a7508 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d88>
  3a69b5:      	mov	rdi, qword ptr [r11 + 0x10]
  3a69b9:      	mov	rcx, qword ptr [rax + 0x8]
  3a69bd:      	test	byte ptr [rcx + 0x17], 0x20
  3a69c1:      	jne	0x3a6bfb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x147b>
  3a69c7:      	mov	r14, rsi
  3a69ca:      	call	qword ptr [rax + 0x28]
  3a69cd:      	jmp	0x3a6c64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14e4>
  3a69d2:      	lea	esi, [rdi - 0x10]
  3a69d5:      	cmp	esi, 0x34
  3a69d8:      	ja	0x3a6c34 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14b4>
  3a69de:      	nop
  3a69e0:      	jmp	qword ptr [8*rsi + 0x204180]
  3a69e7:      	mov	rax, qword ptr [r13 + 0x18]
  3a69eb:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a69f0:      	mov	r10, qword ptr [rbp - 0x68]
  3a69f4:      	mov	rsi, qword ptr [r10 + 0x80]
  3a69fb:      	mov	edx, dword ptr [r13 + 0x4]
  3a69ff:      	nop
  3a6a00:      	cmp	rsi, rdx
  3a6a03:      	jbe	0x3a7403 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c83>
  3a6a09:      	imul	r8, rdx, 0xe0
  3a6a10:      	add	r8, qword ptr [r10 + 0x78]
  3a6a14:      	mov	r10, qword ptr [rbp - 0x60]
  3a6a18:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6a1e:      	mov	rsi, qword ptr [r10 + 0x78]
  3a6a22:      	mov	edx, dword ptr [r8 + 0x18]
  3a6a26:      	cmp	rsi, rdx
  3a6a29:      	jbe	0x3a7403 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c83>
  3a6a2f:      	shl	rdx, 0x3
  3a6a33:      	add	rdx, qword ptr [r10 + 0x70]
  3a6a37:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6a3d:      	mov	rdx, qword ptr [rdx]
  3a6a40:      	test	rdx, rdx
  3a6a43:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6a49:      	mov	rsi, rdi
  3a6a4c:      	mov	rdi, qword ptr [rdx + 0x10]
  3a6a50:      	mov	edx, dword ptr [r8 + 0x14]
  3a6a54:      	mov	r8d, dword ptr [r8 + 0x1c]
  3a6a58:      	shl	r8, 0x20
  3a6a5c:      	or	r8, rdx
  3a6a5f:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a6a64:      	je	0x3a6a6d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12ed>
  3a6a66:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a6a6b:      	je	0x3a6a99 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1319>
  3a6a6d:      	mov	qword ptr [rbp - 0x98], rcx
  3a6a74:      	mov	qword ptr [rbp - 0x90], r9
  3a6a7b:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a6a82:      	movups	xmmword ptr [rsp], xmm0
  3a6a86:      	mov	rdx, rax
  3a6a89:      	mov	rcx, r8
  3a6a8c:      	xor	r8d, r8d
  3a6a8f:      	call	0x3cdea0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a6a94:      	jmp	0x3a671c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf9c>
  3a6a99:      	mov	qword ptr [rbp - 0x98], rcx
  3a6aa0:      	mov	qword ptr [rbp - 0x90], r9
  3a6aa7:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a6aae:      	movups	xmmword ptr [rsp], xmm0
  3a6ab2:      	mov	rdx, rax
  3a6ab5:      	mov	rcx, r8
  3a6ab8:      	xor	r8d, r8d
  3a6abb:      	nop	dword ptr [rax + rax]
  3a6ac0:      	jmp	0x3a6717 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf97>
  3a6ac5:      	mov	rax, qword ptr [rbp - 0x60]
  3a6ac9:      	mov	rsi, qword ptr [rax + 0x90]
  3a6ad0:      	mov	edi, dword ptr [r13 + 0x18]
  3a6ad4:      	cmp	rsi, rdi
  3a6ad7:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a6add:      	mov	rax, qword ptr [rax + 0x88]
  3a6ae4:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a6ae8:      	mov	rax, qword ptr [rcx + 0x20]
  3a6aec:      	mov	rdx, qword ptr [rcx + 0x28]
  3a6af0:      	jmp	0x3a66c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf42>
  3a6af5:      	mov	rcx, qword ptr [rbp - 0x60]
  3a6af9:      	mov	rsi, qword ptr [rcx + 0x90]
  3a6b00:      	mov	edi, dword ptr [r13 + 0x18]
  3a6b04:      	cmp	rsi, rdi
  3a6b07:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a6b0d:      	mov	rcx, qword ptr [rcx + 0x88]
  3a6b14:      	mov	rcx, qword ptr [rcx + 8*rdi]
  3a6b18:      	mov	qword ptr [rcx + 0x20], rax
  3a6b1c:      	mov	qword ptr [rcx + 0x28], rdx
  3a6b20:      	xor	eax, eax
  3a6b22:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6b27:      	mov	edi, dword ptr [r13 + 0x18]
  3a6b2b:      	mov	r9, qword ptr [rbp - 0x38]
  3a6b2f:      	cmp	r9, rdi
  3a6b32:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a6b38:      	shl	rdi, 0x4
  3a6b3c:      	mov	rax, qword ptr [r12 + rdi]
  3a6b40:      	mov	rdx, qword ptr [r12 + rdi + 0x8]
  3a6b45:      	jmp	0x3a66c6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf46>
  3a6b4a:      	mov	rcx, qword ptr [rbp - 0x60]
  3a6b4e:      	mov	rsi, qword ptr [rcx + 0x60]
  3a6b52:      	mov	edi, dword ptr [r13 + 0x18]
  3a6b56:      	cmp	rsi, rdi
  3a6b59:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a6b5f:      	mov	rcx, qword ptr [rcx + 0x58]
  3a6b63:      	lea	rdx, [rcx + 8*rdi]
  3a6b67:      	test	rdx, rdx
  3a6b6a:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6b70:      	mov	rcx, qword ptr [rcx + 8*rdi]
  3a6b74:      	test	rcx, rcx
  3a6b77:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6b7d:      	cmp	byte ptr [rcx + 0x29], 0x0
  3a6b81:      	mov	edx, eax
  3a6b83:      	cmovne	rdx, rax
  3a6b87:      	cmp	rdx, qword ptr [rcx + 0x40]
  3a6b8b:      	jae	0x3a7366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1be6>
  3a6b91:      	mov	rax, qword ptr [rcx + 0x38]
  3a6b95:      	mov	rax, qword ptr [rax + 8*rdx]
  3a6b99:      	jmp	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a6b9b:      	mov	rdx, qword ptr [rbp - 0x60]
  3a6b9f:      	mov	rsi, qword ptr [rdx + 0x60]
  3a6ba3:      	mov	edi, dword ptr [r13 + 0x18]
  3a6ba7:      	cmp	rsi, rdi
  3a6baa:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a6bb0:      	mov	rdx, qword ptr [rdx + 0x58]
  3a6bb4:      	lea	rsi, [rdx + 8*rdi]
  3a6bb8:      	nop	dword ptr [rax + rax]
  3a6bc0:      	test	rsi, rsi
  3a6bc3:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6bc9:      	mov	rdx, qword ptr [rdx + 8*rdi]
  3a6bcd:      	test	rdx, rdx
  3a6bd0:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6bd6:      	cmp	byte ptr [rdx + 0x29], 0x0
  3a6bda:      	mov	esi, eax
  3a6bdc:      	cmovne	rsi, rax
  3a6be0:      	cmp	rsi, qword ptr [rdx + 0x40]
  3a6be4:      	jae	0x3a7366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1be6>
  3a6bea:      	mov	rax, qword ptr [rdx + 0x38]
  3a6bee:      	mov	qword ptr [rax + 8*rsi], rcx
  3a6bf2:      	xor	eax, eax
  3a6bf4:      	xor	edx, edx
  3a6bf6:      	jmp	0x3a669a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf1a>
  3a6bfb:      	mov	r14, rsi
  3a6bfe:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a6c02:      	and	cl, 0x1f
  3a6c05:      	cmp	cl, 0x16
  3a6c08:      	je	0x3a6c15 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1495>
  3a6c0a:      	mov	qword ptr [rbp - 0x98], rdi
  3a6c11:      	mov	rdi, qword ptr [rbp - 0x78]
  3a6c15:      	call	qword ptr [rax + 0x28]
  3a6c18:      	jmp	0x3a6c64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14e4>
  3a6c1a:      	mov	rax, rcx
  3a6c1d:      	mov	rdx, r9
  3a6c20:      	jmp	0x3a66c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf42>
  3a6c25:      	mov	r14, rsi
  3a6c28:      	mov	edi, 0x1
  3a6c2d:      	call	0x61ec00 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a6c32:      	jmp	0x3a6c64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14e4>
  3a6c34:      	lea	esi, [rdi - 0xd0]
  3a6c3a:      	nop	word ptr [rax + rax]
  3a6c40:      	cmp	esi, 0x4
  3a6c43:      	ja	0x3a6d1b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x159b>
  3a6c49:      	jmp	qword ptr [8*rsi + 0x204328]
  3a6c50:      	xor	ecx, ecx
  3a6c52:      	test	rax, rax
  3a6c55:      	sete	cl
  3a6c58:      	mov	rax, rcx
  3a6c5b:      	jmp	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a6c5d:      	mov	r14, rsi
  3a6c60:      	xor	eax, eax
  3a6c62:      	xor	edx, edx
  3a6c64:      	test	rax, rax
  3a6c67:      	mov	r11, qword ptr [rbp - 0x50]
  3a6c6b:      	mov	r8d, 0xffffffff
  3a6c71:      	mov	r9, qword ptr [rbp - 0x38]
  3a6c75:      	mov	rsi, r14
  3a6c78:      	je	0x3a6c87 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a6c7a:      	cmp	qword ptr [rax + 0x8], 0x0
  3a6c7f:      	nop
  3a6c80:      	je	0x3a6c87 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1507>
  3a6c82:      	jmp	0x3a72d6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a6c87:      	mov	rax, qword ptr [rbp - 0x68]
  3a6c8b:      	mov	rdx, qword ptr [rax + 0x108]
  3a6c92:      	lea	eax, [rsi + r13]
  3a6c96:      	cmp	rdx, rax
  3a6c99:      	jb	0x3a751f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d9f>
  3a6c9f:      	nop
  3a6ca0:      	cmp	r13d, eax
  3a6ca3:      	ja	0x3a752e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dae>
  3a6ca9:      	sub	rdx, r13
  3a6cac:      	lea	rcx, [8*r13]
  3a6cb4:      	test	rdx, rdx
  3a6cb7:      	mov	edx, 0x0
  3a6cbc:      	cmovle	rcx, rdx
  3a6cc0:      	test	esi, esi
  3a6cc2:      	je	0x3a61d2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa52>
  3a6cc8:      	mov	rdx, qword ptr [rbp - 0x68]
  3a6ccc:      	mov	rdx, qword ptr [rdx + 0xf8]
  3a6cd3:      	sub	rax, r13
  3a6cd6:      	add	rdx, rcx
  3a6cd9:      	add	rdx, 0x4
  3a6cdd:      	xor	esi, esi
  3a6cdf:      	mov	edi, dword ptr [rdx + 8*rsi]
  3a6ce2:      	cmp	r9, rdi
  3a6ce5:      	jbe	0x3a7355 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a6ceb:      	mov	ecx, dword ptr [rdx + 8*rsi - 0x4]
  3a6cef:      	cmp	r9, rcx
  3a6cf2:      	jbe	0x3a74fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d7d>
  3a6cf8:      	shl	rdi, 0x4
  3a6cfc:      	shl	rcx, 0x4
  3a6d00:      	mov	r10, qword ptr [rbp - 0x58]
  3a6d04:      	movups	xmm0, xmmword ptr [r10 + rdi]
  3a6d09:      	movups	xmmword ptr [r10 + rcx], xmm0
  3a6d0e:      	inc	rsi
  3a6d11:      	cmp	rax, rsi
  3a6d14:      	jne	0x3a6cdf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x155f>
  3a6d16:      	jmp	0x3a61d2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa52>
  3a6d1b:      	mov	rdx, rdi
  3a6d1e:      	mov	edi, dword ptr [r13 + 0x14]
  3a6d22:      	cmp	rdi, r8
  3a6d25:      	je	0x3a6eb5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1735>
  3a6d2b:      	cmp	qword ptr [rbp - 0x38], rdi
  3a6d2f:      	jbe	0x3a735d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bdd>
  3a6d35:      	shl	rdi, 0x4
  3a6d39:      	mov	rsi, qword ptr [rbp - 0x58]
  3a6d3d:      	movups	xmm0, xmmword ptr [rsi + rdi]
  3a6d41:      	jmp	0x3a6eb8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1738>
  3a6d46:      	mov	eax, dword ptr [r13 + 0x18]
  3a6d4a:      	jmp	0x3a66c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf40>
  3a6d4f:      	mov	rax, qword ptr [rbp - 0x60]
  3a6d53:      	mov	rsi, qword ptr [rax + 0x78]
  3a6d57:      	mov	edi, dword ptr [r13 + 0x18]
  3a6d5b:      	nop	dword ptr [rax + rax]
  3a6d60:      	cmp	rsi, rdi
  3a6d63:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a6d69:      	shl	rdi, 0x3
  3a6d6d:      	mov	rax, qword ptr [rbp - 0x60]
  3a6d71:      	add	rdi, qword ptr [rax + 0x70]
  3a6d75:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6d7b:      	mov	rax, qword ptr [rdi]
  3a6d7e:      	nop
  3a6d80:      	test	rax, rax
  3a6d83:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6d89:      	mov	rdi, qword ptr [rax + 0x10]
  3a6d8d:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a6d92:      	mov	rax, qword ptr [rdi + 0x60]
  3a6d96:      	je	0x3a7085 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1905>
  3a6d9c:      	nop	dword ptr [rax]
  3a6da0:      	test	rax, rax
  3a6da3:      	jne	0x3a708e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x190e>
  3a6da9:      	call	0x3ca280 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a6dae:      	jmp	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a6db3:      	mov	rcx, qword ptr [rbp - 0x60]
  3a6db7:      	mov	rsi, qword ptr [rcx + 0x78]
  3a6dbb:      	mov	edi, dword ptr [r13 + 0x18]
  3a6dbf:      	nop
  3a6dc0:      	cmp	rsi, rdi
  3a6dc3:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a6dc9:      	shl	rdi, 0x3
  3a6dcd:      	mov	rcx, qword ptr [rbp - 0x60]
  3a6dd1:      	add	rdi, qword ptr [rcx + 0x70]
  3a6dd5:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6ddb:      	cmp	qword ptr [rbp - 0x50], 0x0
  3a6de0:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6de6:      	mov	r15, qword ptr [rdi]
  3a6de9:      	test	r15, r15
  3a6dec:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6df2:      	mov	rcx, qword ptr [rbp - 0x50]
  3a6df6:      	mov	rdi, qword ptr [rcx]
  3a6df9:      	mov	rcx, qword ptr [r15 + 0x10]
  3a6dfd:      	movzx	ecx, byte ptr [rcx + 0x31]
  3a6e01:      	mov	byte ptr [rbp - 0x79], cl
  3a6e04:      	mov	edx, eax
  3a6e06:      	test	cl, 0x1
  3a6e09:      	cmovne	rdx, rax
  3a6e0d:      	mov	rsi, r15
  3a6e10:      	call	0x3c6400 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a6e15:      	test	dl, 0x1
  3a6e18:      	jne	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a6e1e:      	mov	rax, qword ptr [r15 + 0x10]
  3a6e22:      	movzx	eax, byte ptr [rax + 0x31]
  3a6e26:      	mov	byte ptr [rbp - 0x7a], al
  3a6e29:      	and	al, 0x1
  3a6e2b:      	neg	al
  3a6e2d:      	mov	eax, 0x0
  3a6e32:      	sbb	rax, rax
  3a6e35:      	mov	r8d, 0xffffffff
  3a6e3b:      	or	rax, r8
  3a6e3e:      	xor	edx, edx
  3a6e40:      	mov	r11, qword ptr [rbp - 0x50]
  3a6e44:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a6e49:      	jmp	0x3a66a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf29>
  3a6e4e:      	test	rax, rax
  3a6e51:      	mov	r11, qword ptr [rbp - 0x50]
  3a6e55:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a6e5a:      	mov	r8d, 0xffffffff
  3a6e60:      	mov	r9, qword ptr [rbp - 0x38]
  3a6e64:      	mov	r12, qword ptr [rbp - 0x58]
  3a6e68:      	mov	rcx, qword ptr [rbp - 0x40]
  3a6e6c:      	mov	rsi, qword ptr [rbp - 0x48]
  3a6e70:      	jne	0x3a66ce <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf4e>
  3a6e76:      	jmp	0x3a74d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d53>
  3a6e7b:      	mov	rax, qword ptr [rbp - 0x60]
  3a6e7f:      	mov	rsi, qword ptr [rax + 0x48]
  3a6e83:      	mov	edi, dword ptr [r13 + 0x18]
  3a6e87:      	cmp	rsi, rdi
  3a6e8a:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a6e90:      	mov	rax, qword ptr [rbp - 0x60]
  3a6e94:      	mov	rax, qword ptr [rax + 0x40]
  3a6e98:      	mov	rax, qword ptr [rax + 8*rdi]
  3a6e9c:      	nop	dword ptr [rax]
  3a6ea0:      	jmp	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a6ea5:      	xor	edx, edx
  3a6ea7:      	cmp	rax, rcx
  3a6eaa:      	sete	dl
  3a6ead:      	mov	rax, rdx
  3a6eb0:      	jmp	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a6eb5:      	xorps	xmm0, xmm0
  3a6eb8:      	and	r10d, 0xff0000
  3a6ebf:      	nop
  3a6ec0:      	cmp	r10d, 0xfe0000
  3a6ec7:      	jne	0x3a6f05 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1785>
  3a6ec9:      	movzx	r9d, dx
  3a6ecd:      	cmp	r9d, 0x3
  3a6ed1:      	jne	0x3a70c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1945>
  3a6ed7:      	cmp	qword ptr [rbp - 0x50], 0x0
  3a6edc:      	nop	dword ptr [rax]
  3a6ee0:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6ee6:      	mov	rax, qword ptr [rbp - 0x50]
  3a6eea:      	mov	rax, qword ptr [rax]
  3a6eed:      	test	rax, rax
  3a6ef0:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6ef6:      	mov	rdi, qword ptr [rax + 0x8]
  3a6efa:      	call	0x3c9700 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a6eff:      	nop
  3a6f00:      	jmp	0x3a6bf2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1472>
  3a6f05:      	add	edx, 0xff03fff8
  3a6f0b:      	cmp	edx, 0xa
  3a6f0e:      	jae	0x3a7540 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc0>
  3a6f14:      	mov	rdx, qword ptr [rbp - 0x68]
  3a6f18:      	mov	rsi, qword ptr [rdx + 0x80]
  3a6f1f:      	mov	edi, dword ptr [r13 + 0x4]
  3a6f23:      	cmp	rsi, rdi
  3a6f26:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a6f2c:      	imul	rdx, rdi, 0xe0
  3a6f33:      	mov	rsi, qword ptr [rbp - 0x68]
  3a6f37:      	nop	word ptr [rax + rax]
  3a6f40:      	add	rdx, qword ptr [rsi + 0x78]
  3a6f44:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a6f4a:      	movups	xmm1, xmmword ptr [rdx + 0x98]
  3a6f51:      	movups	xmm2, xmmword ptr [rdx + 0x88]
  3a6f58:      	movups	xmm3, xmmword ptr [rdx + 0x78]
  3a6f5c:      	movups	xmm4, xmmword ptr [rdx + 0x68]
  3a6f60:      	movups	xmm5, xmmword ptr [rdx + 0x58]
  3a6f64:      	movups	xmm6, xmmword ptr [rdx + 0x48]
  3a6f68:      	movzx	esi, byte ptr [rdx + 0x40]
  3a6f6c:      	movups	xmm7, xmmword ptr [rdx + 0x30]
  3a6f70:      	movups	xmm8, xmmword ptr [rdx + 0x20]
  3a6f75:      	movups	xmm9, xmmword ptr [rdx + 0x10]
  3a6f7a:      	mov	rdi, qword ptr [rdx + 0x8]
  3a6f7e:      	mov	edx, dword ptr [rdx]
  3a6f80:      	mov	dword ptr [rbp - 0x218], edx
  3a6f86:      	mov	qword ptr [rbp - 0x210], rdi
  3a6f8d:      	movups	xmmword ptr [rbp - 0x208], xmm9
  3a6f95:      	movups	xmmword ptr [rbp - 0x1f8], xmm8
  3a6f9d:      	movups	xmmword ptr [rbp - 0x1e8], xmm7
  3a6fa4:      	mov	byte ptr [rbp - 0x1d8], sil
  3a6fab:      	movups	xmmword ptr [rbp - 0x1d0], xmm6
  3a6fb2:      	movups	xmmword ptr [rbp - 0x1c0], xmm5
  3a6fb9:      	movups	xmmword ptr [rbp - 0x1b0], xmm4
  3a6fc0:      	movups	xmmword ptr [rbp - 0x1a0], xmm3
  3a6fc7:      	movups	xmmword ptr [rbp - 0x190], xmm2
  3a6fce:      	movups	xmmword ptr [rbp - 0x180], xmm1
  3a6fd5:      	mov	qword ptr [rbp - 0xe8], rcx
  3a6fdc:      	mov	qword ptr [rbp - 0xe0], r9
  3a6fe3:      	movaps	xmmword ptr [rbp - 0x170], xmm0
  3a6fea:      	mov	ecx, 0x15
  3a6fef:      	mov	rdi, rsp
  3a6ff2:      	lea	rsi, [rbp - 0x218]
  3a6ff9:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a6ffc:      	movaps	xmm0, xmmword ptr [rbp - 0x170]
  3a7003:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a700b:      	movups	xmm0, xmmword ptr [rbp - 0xe8]
  3a7012:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a701a:      	lea	rdi, [rbp - 0x110]
  3a7021:      	mov	rsi, qword ptr [rbp - 0x50]
  3a7025:      	mov	rcx, rax
  3a7028:      	mov	rdx, qword ptr [rbp - 0x60]
  3a702c:      	call	0x3ac900 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a7031:      	mov	rax, qword ptr [rbp - 0x110]
  3a7038:      	mov	rdx, qword ptr [rbp - 0x108]
  3a703f:      	movzx	edi, byte ptr [rbp - 0x100]
  3a7046:      	mov	rsi, qword ptr [rbp - 0xf8]
  3a704d:      	test	rsi, rsi
  3a7050:      	je	0x3a71bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a3d>
  3a7056:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a705b:      	mov	r11, qword ptr [rbp - 0x50]
  3a705f:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a7064:      	mov	r8d, 0xffffffff
  3a706a:      	mov	r9, qword ptr [rbp - 0x38]
  3a706e:      	mov	r12, qword ptr [rbp - 0x58]
  3a7072:      	mov	rcx, qword ptr [rbp - 0x40]
  3a7076:      	je	0x3a6734 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfb4>
  3a707c:      	nop	dword ptr [rax]
  3a7080:      	jmp	0x3a74e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d65>
  3a7085:      	test	rax, rax
  3a7088:      	je	0x3a71c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a48>
  3a708e:      	mov	rax, qword ptr [rax]
  3a7091:      	test	rdi, rdi
  3a7094:      	je	0x3a71d5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a55>
  3a709a:      	mov	rcx, qword ptr [rdi + 0x38]
  3a709e:      	test	rcx, rcx
  3a70a1:      	mov	edx, 0x10000
  3a70a6:      	cmove	rcx, rdx
  3a70aa:      	mov	rdx, rax
  3a70ad:      	or	rdx, rcx
  3a70b0:      	shr	rdx, 0x20
  3a70b4:      	je	0x3a71bf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a3f>
  3a70ba:      	xor	edx, edx
  3a70bc:      	div	rcx
  3a70bf:      	nop
  3a70c0:      	jmp	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a70c5:      	mov	rdx, qword ptr [rbp - 0x68]
  3a70c9:      	mov	rsi, qword ptr [rdx + 0x80]
  3a70d0:      	mov	edi, dword ptr [r13 + 0x4]
  3a70d4:      	cmp	rsi, rdi
  3a70d7:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a70dd:      	imul	r11, rdi, 0xe0
  3a70e4:      	mov	rdx, qword ptr [rbp - 0x68]
  3a70e8:      	add	r11, qword ptr [rdx + 0x78]
  3a70ec:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a70f2:      	mov	rdx, qword ptr [rbp - 0x60]
  3a70f6:      	mov	rsi, qword ptr [rdx + 0x78]
  3a70fa:      	mov	edi, dword ptr [r11 + 0x18]
  3a70fe:      	nop
  3a7100:      	cmp	rsi, rdi
  3a7103:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a7109:      	shl	rdi, 0x3
  3a710d:      	mov	rdx, qword ptr [rbp - 0x60]
  3a7111:      	add	rdi, qword ptr [rdx + 0x70]
  3a7115:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a711b:      	mov	rsi, qword ptr [rdi]
  3a711e:      	nop
  3a7120:      	test	rsi, rsi
  3a7123:      	je	0x3a72c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a7129:      	mov	r10, qword ptr [rbp - 0x50]
  3a712d:      	mov	rdi, qword ptr [r10]
  3a7130:      	mov	rdx, qword ptr [r10 + 0x8]
  3a7134:      	mov	r8, qword ptr [rdi + 0x8]
  3a7138:      	mov	r10, qword ptr [r10 + 0x10]
  3a713c:      	mov	rsi, qword ptr [rsi + 0x10]
  3a7140:      	mov	edi, dword ptr [r11 + 0x14]
  3a7144:      	mov	r11d, dword ptr [r11 + 0x1c]
  3a7148:      	shl	r11, 0x20
  3a714c:      	or	r11, rdi
  3a714f:      	movlps	qword ptr [rsp + 0x18], xmm0
  3a7154:      	mov	qword ptr [rsp + 0x10], rcx
  3a7159:      	mov	qword ptr [rsp + 0x8], r11
  3a715e:      	mov	qword ptr [rsp], rax
  3a7162:      	lea	rdi, [rbp - 0x138]
  3a7169:      	mov	rcx, r10
  3a716c:      	call	0x3cabc0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a7171:      	mov	rax, qword ptr [rbp - 0x138]
  3a7178:      	mov	rdx, qword ptr [rbp - 0x130]
  3a717f:      	movzx	edi, byte ptr [rbp - 0x128]
  3a7186:      	mov	rsi, qword ptr [rbp - 0x120]
  3a718d:      	test	rsi, rsi
  3a7190:      	je	0x3a71e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a65>
  3a7192:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a7197:      	mov	r11, qword ptr [rbp - 0x50]
  3a719b:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a71a0:      	mov	r8d, 0xffffffff
  3a71a6:      	mov	r9, qword ptr [rbp - 0x38]
  3a71aa:      	mov	r12, qword ptr [rbp - 0x58]
  3a71ae:      	mov	rcx, qword ptr [rbp - 0x40]
  3a71b2:      	je	0x3a6734 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfb4>
  3a71b8:      	jmp	0x3a74ee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d6e>
  3a71bd:      	jmp	0x3a71e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a65>
  3a71bf:      	xor	edx, edx
  3a71c1:      	div	ecx
  3a71c3:      	jmp	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a71c8:      	mov	rax, qword ptr [rdi + 0x8]
  3a71cc:      	test	rdi, rdi
  3a71cf:      	jne	0x3a709a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x191a>
  3a71d5:      	mov	edi, 0x1
  3a71da:      	call	0x61ec00 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a71df:      	nop
  3a71e0:      	jmp	0x3a6bf4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a71e5:      	mov	r11, qword ptr [rbp - 0x50]
  3a71e9:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a71ee:      	jmp	0x3a612b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9ab>
  3a71f3:      	mov	rax, qword ptr [rbp - 0x70]
  3a71f7:      	mov	dword ptr [rax - 0x38], r14d
  3a71fb:      	mov	rdi, r11
  3a71fe:      	mov	r14, r11
  3a7201:      	call	0x3a9b60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a7206:      	test	al, al
  3a7208:      	jne	0x3a73e9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c69>
  3a720e:      	mov	rcx, qword ptr [r14 + 0x48]
  3a7212:      	test	rcx, rcx
  3a7215:      	mov	r11, r14
  3a7218:      	mov	r8d, 0xffffffff
  3a721e:      	nop
  3a7220:      	jg	0x3a57cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4d>
  3a7226:      	jmp	0x3a754f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dcf>
  3a722b:      	mov	rax, qword ptr [rbp - 0x70]
  3a722f:      	mov	dword ptr [rax - 0x38], r14d
  3a7233:      	mov	rdi, r11
  3a7236:      	mov	r14, r11
  3a7239:      	call	0x3a9640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a723e:      	nop
  3a7240:      	jmp	0x3a7206 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a86>
  3a7242:      	mov	rax, qword ptr [rbp - 0x70]
  3a7246:      	mov	dword ptr [rax - 0x38], r14d
  3a724a:      	mov	rdi, r11
  3a724d:      	mov	r14, r11
  3a7250:      	call	0x3a8340 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a7255:      	test	al, al
  3a7257:      	je	0x3a720e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8e>
  3a7259:      	jmp	0x3a74db <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d5b>
  3a725e:      	mov	rax, qword ptr [rbp - 0x70]
  3a7262:      	mov	dword ptr [rax - 0x38], r14d
  3a7266:      	mov	rdi, r11
  3a7269:      	mov	r14, r11
  3a726c:      	call	0x3a8840 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a7271:      	jmp	0x3a7206 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a86>
  3a7273:      	mov	edi, 0x28
  3a7278:      	call	0x638390 <GC_malloc@plt>
  3a727d:      	nop	dword ptr [rax]
  3a7280:      	test	rax, rax
  3a7283:      	je	0x3a755e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dde>
  3a7289:      	mov	r14, rax
  3a728c:      	add	r12, r15
  3a728f:      	mov	edi, 0x28
  3a7294:      	call	0x619ec0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a7299:      	mov	rax, qword ptr [r12 + 0x8]
  3a729e:      	movaps	xmm0, xmmword ptr [rip + 0x34ea9b] # 0x6f5d40 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a72a5:      	movups	xmmword ptr [r14], xmm0
  3a72a9:      	mov	qword ptr [r14 + 0x10], rax
  3a72ad:      	mov	qword ptr [r14 + 0x18], 0x29a008
  3a72b5:      	mov	qword ptr [r14 + 0x20], 0xe
  3a72bd:      	mov	esi, 0x2baa08
  3a72c2:      	jmp	0x3a73db <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5b>
  3a72c7:      	mov	edi, 0x1
  3a72cc:      	call	0x61ec00 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a72d1:      	jmp	0x3a73e9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c69>
  3a72d6:      	mov	rdi, r11
  3a72d9:      	mov	rsi, rax
  3a72dc:      	nop	dword ptr [rax]
  3a72e0:      	jmp	0x3a73e2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c62>
  3a72e5:      	mov	r15, rax
  3a72e8:      	mov	r14, qword ptr [rbp - 0x148]
  3a72ef:      	mov	edi, 0x8
  3a72f4:      	call	0x638390 <GC_malloc@plt>
  3a72f9:      	nop	dword ptr [rax]
  3a7300:      	test	rax, rax
  3a7303:      	je	0x3a755e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dde>
  3a7309:      	mov	r12, rax
  3a730c:      	mov	edi, 0x8
  3a7311:      	call	0x619ec0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a7316:      	mov	edx, 0x299f78
  3a731b:      	mov	rdi, r15
  3a731e:      	mov	rsi, r14
  3a7321:      	mov	rcx, r12
  3a7324:      	call	0x5eb700 <errors.As>
  3a7329:      	test	al, 0x1
  3a732b:      	je	0x3a733e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bbe>
  3a732d:      	mov	rax, qword ptr [r12]
  3a7331:      	mov	rcx, qword ptr [rbp - 0xb8]
  3a7338:      	mov	ecx, dword ptr [rcx + 0x50]
  3a733b:      	mov	dword ptr [rax + 0x4], ecx
  3a733e:      	mov	rdi, qword ptr [rbp - 0x50]
  3a7342:      	mov	rsi, r15
  3a7345:      	jmp	0x3a73df <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5f>
  3a734a:      	mov	rdi, r14
  3a734d:      	mov	rsi, rcx
  3a7350:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7355:      	mov	rsi, r9
  3a7358:      	call	0x619fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a735d:      	mov	rsi, qword ptr [rbp - 0x38]
  3a7361:      	call	0x619fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a7366:      	mov	dil, 0x3
  3a7369:      	jmp	0x3a736e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bee>
  3a736b:      	mov	dil, 0x1
  3a736e:      	mov	r15d, edi
  3a7371:      	mov	edi, 0x10
  3a7376:      	call	0x638320 <GC_malloc_atomic@plt>
  3a737b:      	nop	dword ptr [rax + rax]
  3a7380:      	test	rax, rax
  3a7383:      	je	0x3a755e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dde>
  3a7389:      	mov	r14, rax
  3a738c:      	mov	edi, 0x10
  3a7391:      	call	0x619ec0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a7396:      	xorps	xmm0, xmm0
  3a7399:      	movups	xmmword ptr [r14], xmm0
  3a739d:      	mov	rax, qword ptr [rbp - 0x68]
  3a73a1:      	mov	rsi, qword ptr [rax + 0x80]
  3a73a8:      	mov	edi, dword ptr [r13 + 0x4]
  3a73ac:      	cmp	rsi, rdi
  3a73af:      	jbe	0x3a73fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a73b1:      	mov	rax, qword ptr [rax + 0x78]
  3a73b5:      	mov	rcx, qword ptr [rbp - 0xb8]
  3a73bc:      	mov	ecx, dword ptr [rcx + 0x50]
  3a73bf:      	imul	rdx, rdi, 0xe0
  3a73c6:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a73cb:      	mov	byte ptr [r14], r15b
  3a73ce:      	mov	dword ptr [r14 + 0x4], ecx
  3a73d2:      	mov	qword ptr [r14 + 0x8], rax
  3a73d6:      	mov	esi, 0x29a390
  3a73db:      	mov	rdi, qword ptr [rbp - 0x50]
  3a73df:      	mov	rdx, r14
  3a73e2:      	call	0x3a50a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a73e7:      	mov	al, 0x3
  3a73e9:      	mov	rdx, rbx
  3a73ec:      	add	rsp, 0x2b8
  3a73f3:      	pop	rbx
  3a73f4:      	pop	r12
  3a73f6:      	pop	r13
  3a73f8:      	pop	r14
  3a73fa:      	pop	r15
  3a73fc:      	pop	rbp
  3a73fd:      	ret
  3a73fe:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7403:      	mov	rdi, rdx
  3a7406:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a740b:      	mov	rsi, qword ptr [rbp - 0x38]
  3a740f:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7414:      	mov	rdi, rcx
  3a7417:      	xor	esi, esi
  3a7419:      	mov	ecx, 0x2
  3a741e:      	call	0x61af80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a7423:      	mov	rdi, rax
  3a7426:      	mov	rsi, r9
  3a7429:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a742e:      	mov	rsi, rcx
  3a7431:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7436:      	mov	rdi, rsi
  3a7439:      	mov	rsi, rcx
  3a743c:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a7441:      	mov	rdi, rcx
  3a7444:      	mov	rsi, r9
  3a7447:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a744c:      	mov	edi, 0x1
  3a7451:      	mov	esi, 0x1
  3a7456:      	call	0x619fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a745b:      	xor	edi, edi
  3a745d:      	xor	esi, esi
  3a745f:      	call	0x619fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a7464:      	xor	esi, esi
  3a7466:      	mov	rdx, rcx
  3a7469:      	mov	ecx, 0x3
  3a746e:      	call	0x61af80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a7473:      	mov	rdi, rax
  3a7476:      	mov	esi, 0x1
  3a747b:      	mov	ecx, 0x2
  3a7480:      	call	0x61af80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a7485:      	mov	rdi, r12
  3a7488:      	mov	esi, 0x1
  3a748d:      	mov	rdx, rax
  3a7490:      	mov	ecx, 0x3
  3a7495:      	call	0x61af80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a749a:      	mov	rcx, qword ptr [rbp - 0x40]
  3a749e:      	mov	rdi, r13
  3a74a1:      	mov	rsi, rcx
  3a74a4:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a74a9:      	mov	rdi, rax
  3a74ac:      	mov	rsi, qword ptr [rbp - 0x38]
  3a74b0:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a74b5:      	mov	edi, 0x3
  3a74ba:      	mov	esi, 0x3
  3a74bf:      	call	0x619fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a74c4:      	mov	edi, 0x2
  3a74c9:      	mov	esi, 0x2
  3a74ce:      	call	0x619fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a74d3:      	mov	dil, 0x9
  3a74d6:      	jmp	0x3a736e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bee>
  3a74db:      	mov	al, 0x1
  3a74dd:      	nop	dword ptr [rax]
  3a74e0:      	jmp	0x3a73e9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c69>
  3a74e5:      	mov	rdx, qword ptr [rbp - 0xf0]
  3a74ec:      	jmp	0x3a74f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d75>
  3a74ee:      	mov	rdx, qword ptr [rbp - 0x118]
  3a74f5:      	mov	rdi, r11
  3a74f8:      	jmp	0x3a73e2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c62>
  3a74fd:      	mov	rdi, rcx
  3a7500:      	mov	rsi, r9
  3a7503:      	call	0x619fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a7508:      	mov	edi, 0x329540
  3a750d:      	mov	esi, 0x329870
  3a7512:      	call	0x619bc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a7517:      	mov	rdi, rcx
  3a751a:      	call	0x61b820 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a751f:      	mov	rdi, rax
  3a7522:      	xor	esi, esi
  3a7524:      	mov	ecx, 0x2
  3a7529:      	call	0x61af80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a752e:      	mov	rdi, r13
  3a7531:      	xor	esi, esi
  3a7533:      	mov	rdx, rax
  3a7536:      	mov	ecx, 0x3
  3a753b:      	call	0x61af80 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a7540:      	mov	edi, 0x3339f8
  3a7545:      	mov	esi, 0x29a018
  3a754a:      	call	0x619bc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a754f:      	mov	edi, 0x3339f8
  3a7554:      	mov	esi, 0x299ff8
  3a7559:      	call	0x619bc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a755e:      	mov	edi, 0x3339f8
  3a7563:      	mov	esi, 0x329518
  3a7568:      	call	0x619bc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a756d:      	int3
  3a756e:      	int3
  3a756f:      	int3
  3a7570:      	int3
  3a7571:      	int3
  3a7572:      	int3
  3a7573:      	int3
  3a7574:      	int3
  3a7575:      	int3
  3a7576:      	int3
  3a7577:      	int3
  3a7578:      	int3
  3a7579:      	int3
  3a757a:      	int3
  3a757b:      	int3
  3a757c:      	int3
  3a757d:      	int3
  3a757e:      	int3
  3a757f:      	int3
