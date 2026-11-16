
/home/runner/work/wasmvm/wasmvm/main-regression/counter-fix:	file format elf64-x86-64

Disassembly of section .text:

00000000003a2810 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a2810:      	push	rbp
  3a2811:      	mov	rbp, rsp
  3a2814:      	push	r15
  3a2816:      	push	r14
  3a2818:      	push	r13
  3a281a:      	push	r12
  3a281c:      	push	rbx
  3a281d:      	sub	rsp, 0x2a8
  3a2824:      	mov	rcx, qword ptr [rdi + 0x48]
  3a2828:      	test	rcx, rcx
  3a282b:      	jle	0x3a41ba <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19aa>
  3a2831:      	mov	rbx, rsi
  3a2834:      	mov	r11, rdi
  3a2837:      	movzx	r12d, byte ptr [rdi + 0x20]
  3a283c:      	mov	r8d, 0xffffffff
  3a2842:      	lea	r14, [rbp - 0x90]
  3a2849:      	mov	qword ptr [rbp - 0x40], rdi
  3a284d:      	mov	byte ptr [rbp - 0x29], r12b
  3a2851:      	mov	r10, qword ptr [r11 + 0x40]
  3a2855:      	shl	rcx, 0x6
  3a2859:      	mov	byte ptr [r10 + rcx - 0x30], 0x0
  3a285f:      	mov	rsi, qword ptr [r10 + rcx - 0x40]
  3a2864:      	mov	rax, qword ptr [rsi + 0x48]
  3a2868:      	mov	qword ptr [rbp - 0x58], rax
  3a286c:      	mov	r9, qword ptr [rax + 0x98]
  3a2873:      	mov	rdi, qword ptr [r10 + rcx - 0x28]
  3a2878:      	mov	rdx, qword ptr [r11 + 0x88]
  3a287f:      	lea	rax, [r9 + rdi]
  3a2883:      	cmp	rax, rdx
  3a2886:      	ja	0x3a4380 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b70>
  3a288c:      	mov	qword ptr [rbp - 0x68], r14
  3a2890:      	cmp	rdi, rax
  3a2893:      	ja	0x3a4392 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b82>
  3a2899:      	mov	r14, qword ptr [rbp - 0x58]
  3a289d:      	mov	r13, qword ptr [r14 + 0xb8]
  3a28a4:      	mov	r15d, dword ptr [r10 + rcx - 0x38]
  3a28a9:      	sub	rdx, rdi
  3a28ac:      	shl	rdi, 0x4
  3a28b0:      	test	rdx, rdx
  3a28b3:      	mov	eax, 0x0
  3a28b8:      	cmovle	rdi, rax
  3a28bc:      	cmp	r13, r15
  3a28bf:      	jbe	0x3a4350 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b40>
  3a28c5:      	add	r10, rcx
  3a28c8:      	mov	qword ptr [rbp - 0x70], r10
  3a28cc:      	mov	qword ptr [rbp - 0xa8], rsi
  3a28d3:      	mov	rax, qword ptr [rsi + 0x40]
  3a28d7:      	mov	qword ptr [rbp - 0x60], rax
  3a28db:      	mov	rsi, qword ptr [r14 + 0xb0]
  3a28e2:      	add	rdi, qword ptr [r11 + 0x78]
  3a28e6:      	mov	r14d, r15d
  3a28e9:      	mov	qword ptr [rbp - 0x38], r9
  3a28ed:      	mov	r10, rdi
  3a28f0:      	mov	qword ptr [rbp - 0x50], rdi
  3a28f4:      	mov	qword ptr [rbp - 0xb0], r13
  3a28fb:      	mov	qword ptr [rbp - 0x48], rsi
  3a28ff:      	jmp	0x3a2931 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x121>
  3a2901:      	nop	word ptr cs:[rax + rax]
  3a2910:      	inc	r14d
  3a2913:      	mov	r13d, r14d
  3a2916:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a291b:      	mov	r15d, r13d
  3a291e:      	mov	r14d, r13d
  3a2921:      	mov	r13, qword ptr [rbp - 0xb0]
  3a2928:      	cmp	r13, r15
  3a292b:      	jbe	0x3a4350 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b40>
  3a2931:      	shl	r15, 0x5
  3a2935:      	add	r15, rsi
  3a2938:      	test	r12b, 0x1
  3a293c:      	je	0x3a299e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18e>
  3a293e:      	mov	rax, qword ptr [rbp - 0x58]
  3a2942:      	mov	rsi, qword ptr [rax + 0x80]
  3a2949:      	mov	edi, dword ptr [r15 + 0x4]
  3a294d:      	cmp	rsi, rdi
  3a2950:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a2956:      	mov	r13, qword ptr [rax + 0x78]
  3a295a:      	imul	r12, rdi, 0xe0
  3a2961:      	mov	eax, dword ptr [r13 + r12]
  3a2966:      	cmp	eax, 0x5
  3a2969:      	setne	cl
  3a296c:      	cmp	eax, 0xb
  3a296f:      	setne	al
  3a2972:      	and	al, cl
  3a2974:      	movzx	ecx, al
  3a2977:      	mov	rax, qword ptr [r11 + 0x150]
  3a297e:      	sub	rax, rcx
  3a2981:      	jb	0x3a41d8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c8>
  3a2987:      	mov	qword ptr [r11 + 0x150], rax
  3a298e:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a2993:      	mov	r13, qword ptr [rbp - 0xb0]
  3a299a:      	mov	rsi, qword ptr [rbp - 0x48]
  3a299e:      	inc	rbx
  3a29a1:      	test	bl, 0x3f
  3a29a4:      	jne	0x3a2a24 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x214>
  3a29a6:      	mov	rax, qword ptr [r11 + 0x160]
  3a29ad:      	test	rax, rax
  3a29b0:      	je	0x3a29c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5>
  3a29b2:      	test	r11, r11
  3a29b5:      	je	0x3a2ab7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2a7>
  3a29bb:      	mov	eax, dword ptr [rax]
  3a29bd:      	test	eax, eax
  3a29bf:      	je	0x3a2f48 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x738>
  3a29c5:      	mov	rax, qword ptr [r11 + 0x8]
  3a29c9:      	test	rax, rax
  3a29cc:      	je	0x3a444e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3e>
  3a29d2:      	mov	rdi, qword ptr [r11 + 0x10]
  3a29d6:      	mov	rcx, qword ptr [rax + 0x8]
  3a29da:      	test	byte ptr [rcx + 0x17], 0x20
  3a29de:      	je	0x3a29f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e7>
  3a29e0:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a29e4:      	and	cl, 0x1f
  3a29e7:      	cmp	cl, 0x16
  3a29ea:      	je	0x3a29f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e7>
  3a29ec:      	mov	qword ptr [rbp - 0x90], rdi
  3a29f3:      	mov	rdi, qword ptr [rbp - 0x68]
  3a29f7:      	call	qword ptr [rax + 0x28]
  3a29fa:      	mov	r8d, 0xffffffff
  3a2a00:      	mov	r9, qword ptr [rbp - 0x38]
  3a2a04:      	mov	r10, qword ptr [rbp - 0x50]
  3a2a08:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2a0c:      	test	rax, rax
  3a2a0f:      	je	0x3a2adc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2cc>
  3a2a15:      	cmp	qword ptr [rax + 0x8], 0x0
  3a2a1a:      	mov	r11, qword ptr [rbp - 0x40]
  3a2a1e:      	jne	0x3a4228 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a18>
  3a2a24:      	cmp	dword ptr [r15 + 0x1c], 0x0
  3a2a29:      	je	0x3a2af0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2a2f:      	test	r12b, 0x1
  3a2a33:      	je	0x3a2a50 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2a35:      	mov	edx, dword ptr [r15]
  3a2a38:      	cmp	edx, 0x20
  3a2a3b:      	je	0x3a2af0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2a41:      	cmp	edx, 0x41
  3a2a44:      	je	0x3a2af0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2a4a:      	jmp	0x3a2af3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e3>
  3a2a4f:      	nop
  3a2a50:      	mov	edx, dword ptr [r15]
  3a2a53:      	cmp	edx, 0x20
  3a2a56:      	je	0x3a2a61 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x251>
  3a2a58:      	cmp	edx, 0x41
  3a2a5b:      	jne	0x3a2af3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e3>
  3a2a61:      	mov	eax, ebx
  3a2a63:      	not	eax
  3a2a65:      	test	al, 0x3f
  3a2a67:      	je	0x3a2af0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2a6d:      	cmp	edx, 0x41
  3a2a70:      	jne	0x3a2a94 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x284>
  3a2a72:      	mov	edi, dword ptr [r15 + 0x8]
  3a2a76:      	cmp	r9, rdi
  3a2a79:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2a7f:      	mov	eax, dword ptr [r15 + 0x18]
  3a2a83:      	shl	rdi, 0x4
  3a2a87:      	mov	qword ptr [r10 + rdi], rax
  3a2a8b:      	mov	qword ptr [r10 + rdi + 0x8], 0x0
  3a2a94:      	inc	r14d
  3a2a97:      	cmp	r13, r14
  3a2a9a:      	jbe	0x3a4345 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b35>
  3a2aa0:      	mov	rax, r14
  3a2aa3:      	shl	rax, 0x5
  3a2aa7:      	lea	r15, [rsi + rax]
  3a2aab:      	inc	rbx
  3a2aae:      	cmp	dword ptr [rsi + rax + 0x1c], 0x0
  3a2ab3:      	jne	0x3a2a50 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2ab5:      	jmp	0x3a2af0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2ab7:      	mov	edi, 0x1
  3a2abc:      	call	0x5e7240 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a2ac1:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2ac5:      	mov	r10, qword ptr [rbp - 0x50]
  3a2ac9:      	mov	r9, qword ptr [rbp - 0x38]
  3a2acd:      	mov	r8d, 0xffffffff
  3a2ad3:      	test	rax, rax
  3a2ad6:      	jne	0x3a2a15 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x205>
  3a2adc:      	mov	r11, qword ptr [rbp - 0x40]
  3a2ae0:      	cmp	dword ptr [r15 + 0x1c], 0x0
  3a2ae5:      	jne	0x3a2a2f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21f>
  3a2aeb:      	nop	dword ptr [rax + rax]
  3a2af0:      	mov	edx, dword ptr [r15]
  3a2af3:      	lea	eax, [rdx - 0xe0]
  3a2af9:      	cmp	eax, 0xffff20
  3a2afe:      	jae	0x3a2d70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x560>
  3a2b04:      	mov	eax, edx
  3a2b06:      	shr	eax, 0x10
  3a2b09:      	cmp	eax, 0xfb
  3a2b0e:      	je	0x3a2ce0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4d0>
  3a2b14:      	cmp	eax, 0xfc
  3a2b19:      	je	0x3a2bbe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3ae>
  3a2b1f:      	cmp	eax, 0xfd
  3a2b24:      	jne	0x3a2d62 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x552>
  3a2b2a:      	mov	edi, dword ptr [r15 + 0xc]
  3a2b2e:      	cmp	rdi, r8
  3a2b31:      	je	0x3a3000 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7f0>
  3a2b37:      	cmp	r9, rdi
  3a2b3a:      	mov	r12, qword ptr [rbp - 0x58]
  3a2b3e:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2b44:      	shl	rdi, 0x4
  3a2b48:      	mov	rcx, qword ptr [r10 + rdi]
  3a2b4c:      	mov	r10, qword ptr [r10 + rdi + 0x8]
  3a2b51:      	mov	edi, dword ptr [r15 + 0x10]
  3a2b55:      	mov	esi, 0xffffffff
  3a2b5a:      	cmp	rdi, r8
  3a2b5d:      	je	0x3a301b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x80b>
  3a2b63:      	cmp	r9, rdi
  3a2b66:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2b6c:      	shl	rdi, 0x4
  3a2b70:      	mov	r8, qword ptr [rbp - 0x50]
  3a2b74:      	mov	rax, qword ptr [r8 + rdi]
  3a2b78:      	mov	r8, qword ptr [r8 + rdi + 0x8]
  3a2b7d:      	mov	edi, dword ptr [r15 + 0x14]
  3a2b81:      	cmp	rdi, rsi
  3a2b84:      	je	0x3a302d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x81d>
  3a2b8a:      	cmp	r9, rdi
  3a2b8d:      	jbe	0x3a4366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a2b93:      	shl	rdi, 0x4
  3a2b97:      	mov	rsi, qword ptr [rbp - 0x50]
  3a2b9b:      	mov	r9, qword ptr [rsi + rdi]
  3a2b9f:      	mov	r11, qword ptr [rsi + rdi + 0x8]
  3a2ba4:      	mov	rsi, qword ptr [r12 + 0x80]
  3a2bac:      	mov	edi, dword ptr [r15 + 0x4]
  3a2bb0:      	cmp	rsi, rdi
  3a2bb3:      	ja	0x3a3048 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x838>
  3a2bb9:      	jmp	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a2bbe:      	add	edx, 0xff03ffed
  3a2bc4:      	cmp	edx, 0x4
  3a2bc7:      	jae	0x3a2d70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x560>
  3a2bcd:      	mov	rax, qword ptr [rbp - 0x58]
  3a2bd1:      	mov	rdx, qword ptr [rax + 0xd8]
  3a2bd8:      	mov	edi, dword ptr [r15 + 0x18]
  3a2bdc:      	mov	eax, dword ptr [r15 + 0x1c]
  3a2be0:      	lea	ecx, [rdi + rax]
  3a2be3:      	add	ecx, 0x2
  3a2be6:      	cmp	rdx, rcx
  3a2be9:      	jb	0x3a43f4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1be4>
  3a2bef:      	cmp	edi, ecx
  3a2bf1:      	ja	0x3a43e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a2bf7:      	sub	rdx, rdi
  3a2bfa:      	lea	r8, [4*rdi]
  3a2c02:      	test	rdx, rdx
  3a2c05:      	mov	eax, 0x0
  3a2c0a:      	cmovle	r8, rax
  3a2c0e:      	cmp	ecx, edi
  3a2c10:      	je	0x3a43dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bcc>
  3a2c16:      	mov	rax, qword ptr [rbp - 0x58]
  3a2c1a:      	add	r8, qword ptr [rax + 0xc8]
  3a2c21:      	mov	eax, dword ptr [r8]
  3a2c24:      	cmp	r9, rax
  3a2c27:      	jbe	0x3a43a4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b94>
  3a2c2d:      	sub	rcx, rdi
  3a2c30:      	cmp	ecx, 0x1
  3a2c33:      	je	0x3a43cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bbd>
  3a2c39:      	mov	edi, dword ptr [r8 + 0x4]
  3a2c3d:      	cmp	r9, rdi
  3a2c40:      	jbe	0x3a4377 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b67>
  3a2c46:      	shl	rax, 0x4
  3a2c4a:      	mov	r9, qword ptr [r10 + rax]
  3a2c4e:      	shl	rdi, 0x4
  3a2c52:      	mov	rsi, qword ptr [r10 + rdi]
  3a2c56:      	mov	r10d, dword ptr [r15]
  3a2c59:      	lea	eax, [r10 - 0xfc0013]
  3a2c60:      	cmp	eax, 0x2
  3a2c63:      	jae	0x3a3161 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x951>
  3a2c69:      	cmp	ecx, 0x3
  3a2c6c:      	jb	0x3a441e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c0e>
  3a2c72:      	mov	edi, dword ptr [r8 + 0x8]
  3a2c76:      	cmp	qword ptr [rbp - 0x38], rdi
  3a2c7a:      	jbe	0x3a4377 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b67>
  3a2c80:      	cmp	ecx, 0x3
  3a2c83:      	je	0x3a440f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bff>
  3a2c89:      	mov	eax, dword ptr [r8 + 0xc]
  3a2c8d:      	cmp	qword ptr [rbp - 0x38], rax
  3a2c91:      	jbe	0x3a4403 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bf3>
  3a2c97:      	shl	rdi, 0x4
  3a2c9b:      	mov	rdx, qword ptr [rbp - 0x50]
  3a2c9f:      	mov	rdi, qword ptr [rdx + rdi]
  3a2ca3:      	shl	rax, 0x4
  3a2ca7:      	mov	rdx, qword ptr [rdx + rax]
  3a2cab:      	cmp	r10d, 0xfc0013
  3a2cb2:      	jne	0x3a3294 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa84>
  3a2cb8:      	lea	rax, [rdi + r9]
  3a2cbc:      	mov	r10, rdi
  3a2cbf:      	and	r10, r9
  3a2cc2:      	or	rdi, r9
  3a2cc5:      	mov	r9, rax
  3a2cc8:      	not	r9
  3a2ccb:      	and	r9, rdi
  3a2cce:      	or	r9, r10
  3a2cd1:      	shr	r9, 0x3f
  3a2cd5:      	add	rdx, rsi
  3a2cd8:      	add	rdx, r9
  3a2cdb:      	jmp	0x3a32b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa9>
  3a2ce0:      	add	edx, -0x2
  3a2ce3:      	cmp	dx, 0x21
  3a2ce7:      	jae	0x3a2ed9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6c9>
  3a2ced:      	lea	rax, [r8 + 0x1c003e10]
  3a2cf4:      	movzx	ecx, dx
  3a2cf7:      	bt	rax, rcx
  3a2cfb:      	jae	0x3a2ed9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6c9>
  3a2d01:      	mov	edi, dword ptr [r15 + 0xc]
  3a2d05:      	mov	eax, 0xffffffff
  3a2d0a:      	cmp	rdi, r8
  3a2d0d:      	je	0x3a3334 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb24>
  3a2d13:      	cmp	r9, rdi
  3a2d16:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2d1c:      	shl	rdi, 0x4
  3a2d20:      	mov	r8, qword ptr [r10 + rdi]
  3a2d24:      	mov	edi, dword ptr [r15 + 0x10]
  3a2d28:      	cmp	rdi, rax
  3a2d2b:      	je	0x3a3344 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb34>
  3a2d31:      	cmp	r9, rdi
  3a2d34:      	jbe	0x3a4366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a2d3a:      	shl	rdi, 0x4
  3a2d3e:      	mov	rax, qword ptr [r10 + rdi]
  3a2d42:      	mov	r9, qword ptr [r10 + rdi + 0x8]
  3a2d47:      	mov	r10d, 0x0
  3a2d4d:      	mov	r11d, 0x0
  3a2d53:      	cmp	dx, 0xc
  3a2d57:      	je	0x3a335b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb4b>
  3a2d5d:      	jmp	0x3a339b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb8b>
  3a2d62:      	cmp	edx, 0xe7
  3a2d68:      	jb	0x3a4119 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1909>
  3a2d6e:      	nop
  3a2d70:      	mov	edi, dword ptr [r15 + 0xc]
  3a2d74:      	cmp	rdi, r8
  3a2d77:      	je	0x3a2e00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5f0>
  3a2d7d:      	cmp	r9, rdi
  3a2d80:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2d86:      	shl	rdi, 0x4
  3a2d8a:      	mov	r12, qword ptr [r10 + rdi]
  3a2d8e:      	mov	rdx, qword ptr [r10 + rdi + 0x8]
  3a2d93:      	mov	edi, dword ptr [r15 + 0x10]
  3a2d97:      	cmp	rdi, r8
  3a2d9a:      	je	0x3a2e0e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5fe>
  3a2d9c:      	cmp	r9, rdi
  3a2d9f:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2da5:      	shl	rdi, 0x4
  3a2da9:      	mov	rax, r10
  3a2dac:      	mov	r10, qword ptr [r10 + rdi]
  3a2db0:      	mov	rax, qword ptr [rax + rdi + 0x8]
  3a2db5:      	mov	r9d, dword ptr [r15]
  3a2db8:      	lea	ecx, [r9 - 0x45]
  3a2dbc:      	cmp	ecx, 0x7f
  3a2dbf:      	ja	0x3a2e1f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x60f>
  3a2dc1:      	ja	0x3a3b74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1364>
  3a2dc7:      	jmp	qword ptr [8*rcx + 0x203ab8]
  3a2dce:      	mov	edi, r9d
  3a2dd1:      	mov	rsi, r12
  3a2dd4:      	mov	rdx, r10
  3a2dd7:      	call	0x3d0d10 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a2ddc:      	test	dl, dl
  3a2dde:      	jne	0x3a4233 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a23>
  3a2de4:      	xor	edx, edx
  3a2de6:      	mov	r11, qword ptr [rbp - 0x40]
  3a2dea:      	mov	r8d, 0xffffffff
  3a2df0:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2df4:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a2df9:      	nop	dword ptr [rax]
  3a2e00:      	xor	r12d, r12d
  3a2e03:      	xor	edx, edx
  3a2e05:      	mov	edi, dword ptr [r15 + 0x10]
  3a2e09:      	cmp	rdi, r8
  3a2e0c:      	jne	0x3a2d9c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x58c>
  3a2e0e:      	xor	r10d, r10d
  3a2e11:      	xor	eax, eax
  3a2e13:      	mov	r9d, dword ptr [r15]
  3a2e16:      	lea	ecx, [r9 - 0x45]
  3a2e1a:      	cmp	ecx, 0x7f
  3a2e1d:      	jbe	0x3a2dc1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5b1>
  3a2e1f:      	lea	ecx, [r9 - 0x1]
  3a2e23:      	cmp	ecx, 0xd3
  3a2e29:      	ja	0x3a2f16 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x706>
  3a2e2f:      	jmp	qword ptr [8*rcx + 0x2033d8]
  3a2e36:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2e3a:      	mov	rsi, qword ptr [rcx + 0x80]
  3a2e41:      	mov	edi, dword ptr [r15 + 0x4]
  3a2e45:      	cmp	rsi, rdi
  3a2e48:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a2e4e:      	imul	rax, rdi, 0xe0
  3a2e55:      	add	rax, qword ptr [rcx + 0x78]
  3a2e59:      	mov	rcx, qword ptr [rbp - 0x60]
  3a2e5d:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a2e63:      	mov	rsi, qword ptr [rcx + 0x78]
  3a2e67:      	mov	edi, dword ptr [rax + 0x18]
  3a2e6a:      	cmp	rsi, rdi
  3a2e6d:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a2e73:      	shl	rdi, 0x3
  3a2e77:      	add	rdi, qword ptr [rcx + 0x70]
  3a2e7b:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a2e81:      	mov	rcx, qword ptr [rdi]
  3a2e84:      	test	rcx, rcx
  3a2e87:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a2e8d:      	mov	rsi, qword ptr [rcx + 0x10]
  3a2e91:      	mov	ecx, dword ptr [rax + 0x14]
  3a2e94:      	mov	r8d, dword ptr [rax + 0x1c]
  3a2e98:      	shl	r8, 0x20
  3a2e9c:      	or	r8, rcx
  3a2e9f:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a2ea4:      	je	0x3a2eb1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6a1>
  3a2ea6:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a2eab:      	je	0x3a325b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa4b>
  3a2eb1:      	xorps	xmm0, xmm0
  3a2eb4:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a2ebb:      	movups	xmmword ptr [rsp], xmm0
  3a2ebf:      	mov	r13, qword ptr [rbp - 0x68]
  3a2ec3:      	mov	rdi, r13
  3a2ec6:      	mov	edx, r9d
  3a2ec9:      	mov	rcx, r12
  3a2ecc:      	xor	r9d, r9d
  3a2ecf:      	call	0x3c6e70 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a2ed4:      	jmp	0x3a327e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa6e>
  3a2ed9:      	mov	r15, qword ptr [rbp - 0x70]
  3a2edd:      	mov	dword ptr [r15 - 0x38], r14d
  3a2ee1:      	mov	rdi, r11
  3a2ee4:      	call	0x3a46c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a2ee9:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2eed:      	mov	r10, qword ptr [rbp - 0x50]
  3a2ef1:      	mov	r9, qword ptr [rbp - 0x38]
  3a2ef5:      	mov	r8d, 0xffffffff
  3a2efb:      	mov	r11, qword ptr [rbp - 0x40]
  3a2eff:      	test	al, al
  3a2f01:      	jne	0x3a430f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a2f07:      	inc	r14d
  3a2f0a:      	mov	dword ptr [r15 - 0x38], r14d
  3a2f0e:      	mov	r13d, r14d
  3a2f11:      	jmp	0x3a291b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10b>
  3a2f16:      	lea	ecx, [r9 - 0xfc0000]
  3a2f1d:      	cmp	ecx, 0x8
  3a2f20:      	jae	0x3a3114 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x904>
  3a2f26:      	mov	edi, r9d
  3a2f29:      	mov	rsi, r12
  3a2f2c:      	mov	rdx, r10
  3a2f2f:      	call	0x3d0d10 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a2f34:      	mov	r8d, 0xffffffff
  3a2f3a:      	mov	r11, qword ptr [rbp - 0x40]
  3a2f3e:      	mov	r13d, edx
  3a2f41:      	xor	edx, edx
  3a2f43:      	jmp	0x3a3bcd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13bd>
  3a2f48:      	xor	eax, eax
  3a2f4a:      	xor	edx, edx
  3a2f4c:      	test	rax, rax
  3a2f4f:      	jne	0x3a2a15 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x205>
  3a2f55:      	jmp	0x3a2adc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2cc>
  3a2f5a:      	mov	rdx, qword ptr [rbp - 0x58]
  3a2f5e:      	mov	rsi, qword ptr [rdx + 0x80]
  3a2f65:      	mov	edi, dword ptr [r15 + 0x4]
  3a2f69:      	cmp	rsi, rdi
  3a2f6c:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a2f72:      	imul	rcx, rdi, 0xe0
  3a2f79:      	add	rcx, qword ptr [rdx + 0x78]
  3a2f7d:      	mov	rdx, qword ptr [rbp - 0x60]
  3a2f81:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a2f87:      	mov	rsi, qword ptr [rdx + 0x78]
  3a2f8b:      	mov	edi, dword ptr [rcx + 0x18]
  3a2f8e:      	cmp	rsi, rdi
  3a2f91:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a2f97:      	shl	rdi, 0x3
  3a2f9b:      	add	rdi, qword ptr [rdx + 0x70]
  3a2f9f:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a2fa5:      	mov	rdx, qword ptr [rdi]
  3a2fa8:      	test	rdx, rdx
  3a2fab:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a2fb1:      	mov	rdi, qword ptr [rdx + 0x10]
  3a2fb5:      	mov	edx, dword ptr [rcx + 0x14]
  3a2fb8:      	mov	ecx, dword ptr [rcx + 0x1c]
  3a2fbb:      	shl	rcx, 0x20
  3a2fbf:      	or	rcx, rdx
  3a2fc2:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a2fc7:      	je	0x3a2fd4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7c4>
  3a2fc9:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a2fce:      	je	0x3a34ae <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc9e>
  3a2fd4:      	mov	qword ptr [rbp - 0x90], r10
  3a2fdb:      	mov	qword ptr [rbp - 0x88], rax
  3a2fe2:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a2fe9:      	movups	xmmword ptr [rsp], xmm0
  3a2fed:      	mov	esi, r9d
  3a2ff0:      	mov	rdx, r12
  3a2ff3:      	xor	r8d, r8d
  3a2ff6:      	call	0x3c8360 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a2ffb:      	jmp	0x3a3bbc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13ac>
  3a3000:      	xor	r10d, r10d
  3a3003:      	xor	ecx, ecx
  3a3005:      	mov	r12, qword ptr [rbp - 0x58]
  3a3009:      	mov	edi, dword ptr [r15 + 0x10]
  3a300d:      	mov	esi, 0xffffffff
  3a3012:      	cmp	rdi, r8
  3a3015:      	jne	0x3a2b63 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x353>
  3a301b:      	xor	eax, eax
  3a301d:      	xor	r8d, r8d
  3a3020:      	mov	edi, dword ptr [r15 + 0x14]
  3a3024:      	cmp	rdi, rsi
  3a3027:      	jne	0x3a2b8a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x37a>
  3a302d:      	xor	r9d, r9d
  3a3030:      	xor	r11d, r11d
  3a3033:      	mov	rsi, qword ptr [r12 + 0x80]
  3a303b:      	mov	edi, dword ptr [r15 + 0x4]
  3a303f:      	cmp	rsi, rdi
  3a3042:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3048:      	imul	rdi, rdi, 0xe0
  3a304f:      	add	rdi, qword ptr [r12 + 0x78]
  3a3054:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a305a:      	movzx	esi, dx
  3a305d:      	cmp	esi, 0xb
  3a3060:      	setb	r12b
  3a3064:      	mov	esi, edx
  3a3066:      	and	esi, 0xfffc
  3a306c:      	cmp	esi, 0x54
  3a306f:      	sete	r13b
  3a3073:      	or	r13b, r12b
  3a3076:      	jne	0x3a308b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x87b>
  3a3078:      	movzx	r12d, dx
  3a307c:      	lea	r13d, [r12 - 0x5c]
  3a3081:      	cmp	r13d, 0x2
  3a3085:      	jae	0x3a3190 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x980>
  3a308b:      	mov	r10, qword ptr [rbp - 0x60]
  3a308f:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3093:      	mov	r9d, dword ptr [rdi + 0x18]
  3a3097:      	cmp	rsi, r9
  3a309a:      	jbe	0x3a436f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5f>
  3a30a0:      	shl	r9, 0x3
  3a30a4:      	add	r9, qword ptr [r10 + 0x70]
  3a30a8:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a30ae:      	mov	rsi, qword ptr [r9]
  3a30b1:      	test	rsi, rsi
  3a30b4:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a30ba:      	mov	rsi, qword ptr [rsi + 0x10]
  3a30be:      	mov	r9d, dword ptr [rdi + 0x14]
  3a30c2:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a30c6:      	shl	r10, 0x20
  3a30ca:      	or	r10, r9
  3a30cd:      	movzx	edi, byte ptr [rdi + 0x40]
  3a30d1:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a30d6:      	je	0x3a30e3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8d3>
  3a30d8:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a30dd:      	je	0x3a3229 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa19>
  3a30e3:      	mov	qword ptr [rbp - 0xa0], rax
  3a30ea:      	mov	qword ptr [rbp - 0x98], r8
  3a30f1:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a30f8:      	movups	xmmword ptr [rsp], xmm0
  3a30fc:      	movzx	r9d, dil
  3a3100:      	mov	r13, qword ptr [rbp - 0x68]
  3a3104:      	mov	rdi, r13
  3a3107:      	mov	r8, r10
  3a310a:      	call	0x3c6e70 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a310f:      	jmp	0x3a327e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa6e>
  3a3114:      	lea	ecx, [r9 - 0x1000001]
  3a311b:      	cmp	ecx, 0x8
  3a311e:      	ja	0x3a3d7e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x156e>
  3a3124:      	mov	edx, 0x117
  3a3129:      	bt	edx, ecx
  3a312c:      	jae	0x3a3d6c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x155c>
  3a3132:      	movzx	eax, r9b
  3a3136:      	xor	edi, edi
  3a3138:      	add	eax, -0x2
  3a313b:      	cmp	eax, 0x7
  3a313e:      	ja	0x3a3bea <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13da>
  3a3144:      	mov	r9, qword ptr [rbp - 0x38]
  3a3148:      	mov	r10, qword ptr [rbp - 0x50]
  3a314c:      	jmp	qword ptr [8*rax + 0x203a78]
  3a3153:      	xor	edi, edi
  3a3155:      	test	r12d, r12d
  3a3158:      	sete	dil
  3a315c:      	jmp	0x3a3bf2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13e2>
  3a3161:      	mov	rax, rsi
  3a3164:      	mul	r9
  3a3167:      	cmp	dword ptr [r15], 0xfc0015
  3a316e:      	jne	0x3a32b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa9>
  3a3174:      	mov	rdi, r9
  3a3177:      	sar	rdi, 0x3f
  3a317b:      	and	rdi, rsi
  3a317e:      	sar	rsi, 0x3f
  3a3182:      	and	rsi, r9
  3a3185:      	add	rsi, rdi
  3a3188:      	sub	rdx, rsi
  3a318b:      	jmp	0x3a32b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa9>
  3a3190:      	cmp	r12d, 0xb
  3a3194:      	je	0x3a319f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x98f>
  3a3196:      	cmp	esi, 0x58
  3a3199:      	jne	0x3a3513 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd03>
  3a319f:      	mov	r10, qword ptr [rbp - 0x60]
  3a31a3:      	mov	rsi, qword ptr [r10 + 0x78]
  3a31a7:      	mov	r9d, dword ptr [rdi + 0x18]
  3a31ab:      	cmp	rsi, r9
  3a31ae:      	jbe	0x3a436f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5f>
  3a31b4:      	shl	r9, 0x3
  3a31b8:      	add	r9, qword ptr [r10 + 0x70]
  3a31bc:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a31c2:      	mov	rsi, qword ptr [r9]
  3a31c5:      	test	rsi, rsi
  3a31c8:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a31ce:      	mov	rsi, qword ptr [rsi + 0x10]
  3a31d2:      	mov	r10d, dword ptr [rdi + 0x14]
  3a31d6:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a31da:      	shl	r9, 0x20
  3a31de:      	or	r9, r10
  3a31e1:      	movzx	edi, byte ptr [rdi + 0x40]
  3a31e5:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a31ea:      	je	0x3a31f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9e7>
  3a31ec:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a31f1:      	je	0x3a3b8f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x137f>
  3a31f7:      	mov	qword ptr [rbp - 0x90], rax
  3a31fe:      	mov	qword ptr [rbp - 0x88], r8
  3a3205:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a320c:      	movups	xmmword ptr [rsp], xmm0
  3a3210:      	movzx	r8d, dil
  3a3214:      	mov	rdi, rsi
  3a3217:      	mov	esi, edx
  3a3219:      	mov	rdx, rcx
  3a321c:      	mov	rcx, r9
  3a321f:      	call	0x3c8360 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a3224:      	jmp	0x3a3bbc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13ac>
  3a3229:      	mov	qword ptr [rbp - 0xa0], rax
  3a3230:      	mov	qword ptr [rbp - 0x98], r8
  3a3237:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a323e:      	movups	xmmword ptr [rsp], xmm0
  3a3242:      	movzx	r9d, dil
  3a3246:      	mov	r13, qword ptr [rbp - 0x68]
  3a324a:      	mov	rdi, r13
  3a324d:      	mov	r8, r10
  3a3250:      	jmp	0x3a3279 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa69>
  3a3252:      	mov	rax, qword ptr [r15 + 0x18]
  3a3256:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a325b:      	xorps	xmm0, xmm0
  3a325e:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a3265:      	movups	xmmword ptr [rsp], xmm0
  3a3269:      	mov	r13, qword ptr [rbp - 0x68]
  3a326d:      	mov	rdi, r13
  3a3270:      	mov	edx, r9d
  3a3273:      	mov	rcx, r12
  3a3276:      	xor	r9d, r9d
  3a3279:      	call	0x3c6fb0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a327e:      	mov	rax, qword ptr [r13]
  3a3282:      	mov	rdx, qword ptr [r13 + 0x8]
  3a3286:      	mov	qword ptr [rbp - 0x68], r13
  3a328a:      	movzx	r13d, byte ptr [r13 + 0x10]
  3a328f:      	jmp	0x3a3bc3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13b3>
  3a3294:      	mov	rax, r9
  3a3297:      	sub	rax, rdi
  3a329a:      	not	r9
  3a329d:      	mov	r10, rdi
  3a32a0:      	and	r10, r9
  3a32a3:      	xor	r9, rdi
  3a32a6:      	and	r9, rax
  3a32a9:      	or	r9, r10
  3a32ac:      	shr	r9, 0x3f
  3a32b0:      	add	rdx, r9
  3a32b3:      	sub	rsi, rdx
  3a32b6:      	mov	rdx, rsi
  3a32b9:      	mov	edi, dword ptr [r15 + 0x1c]
  3a32bd:      	cmp	ecx, edi
  3a32bf:      	mov	r9, qword ptr [rbp - 0x38]
  3a32c3:      	jbe	0x3a43af <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b9f>
  3a32c9:      	lea	esi, [rdi + 0x1]
  3a32cc:      	cmp	ecx, esi
  3a32ce:      	mov	r10, qword ptr [rbp - 0x70]
  3a32d2:      	jbe	0x3a43b7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ba7>
  3a32d8:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a32dc:      	cmp	r9, rcx
  3a32df:      	jbe	0x3a43c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bb2>
  3a32e5:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a32e9:      	shl	rcx, 0x4
  3a32ed:      	mov	rsi, qword ptr [rbp - 0x50]
  3a32f1:      	mov	qword ptr [rsi + rcx], rax
  3a32f5:      	mov	qword ptr [rsi + rcx + 0x8], 0x0
  3a32fe:      	cmp	r9, rdi
  3a3301:      	jbe	0x3a43c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bb5>
  3a3307:      	shl	rdi, 0x4
  3a330b:      	mov	qword ptr [rsi + rdi], rdx
  3a330f:      	mov	qword ptr [rsi + rdi + 0x8], 0x0
  3a3318:      	inc	r14d
  3a331b:      	mov	dword ptr [r10 - 0x38], r14d
  3a331f:      	mov	r13d, r14d
  3a3322:      	mov	r8d, 0xffffffff
  3a3328:      	mov	r10, rsi
  3a332b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a332f:      	jmp	0x3a291b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10b>
  3a3334:      	xor	r8d, r8d
  3a3337:      	mov	edi, dword ptr [r15 + 0x10]
  3a333b:      	cmp	rdi, rax
  3a333e:      	jne	0x3a2d31 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x521>
  3a3344:      	xor	eax, eax
  3a3346:      	xor	r9d, r9d
  3a3349:      	mov	r10d, 0x0
  3a334f:      	mov	r11d, 0x0
  3a3355:      	cmp	dx, 0xc
  3a3359:      	jne	0x3a339b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb8b>
  3a335b:      	mov	rcx, qword ptr [rbp - 0x58]
  3a335f:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a3366:      	mov	edi, dword ptr [r15 + 0x18]
  3a336a:      	add	edi, 0x2
  3a336d:      	cmp	rsi, rdi
  3a3370:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3376:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a337d:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a3380:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3384:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3388:      	jbe	0x3a4366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a338e:      	shl	rdi, 0x4
  3a3392:      	mov	r10, qword ptr [rcx + rdi]
  3a3396:      	mov	r11, qword ptr [rcx + rdi + 0x8]
  3a339b:      	cmp	qword ptr [rbp - 0x40], 0x0
  3a33a0:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a33a6:      	mov	rdx, qword ptr [rbp - 0x58]
  3a33aa:      	mov	rsi, qword ptr [rdx + 0x80]
  3a33b1:      	mov	edi, dword ptr [r15 + 0x4]
  3a33b5:      	cmp	rsi, rdi
  3a33b8:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a33be:      	imul	rcx, rdi, 0xe0
  3a33c5:      	add	rcx, qword ptr [rdx + 0x78]
  3a33c9:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a33cf:      	mov	rdx, qword ptr [rbp - 0x40]
  3a33d3:      	mov	rsi, qword ptr [rdx]
  3a33d6:      	mov	qword ptr [rbp - 0xd0], rax
  3a33dd:      	mov	qword ptr [rbp - 0xc8], r9
  3a33e4:      	mov	qword ptr [rbp - 0xc0], r10
  3a33eb:      	mov	qword ptr [rbp - 0xb8], r11
  3a33f2:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a33f9:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a33fe:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a3405:      	movups	xmmword ptr [rsp], xmm0
  3a3409:      	lea	rdi, [rbp - 0x150]
  3a3410:      	mov	rdx, qword ptr [rbp - 0x60]
  3a3414:      	call	0x3b6880 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a3419:      	movaps	xmm0, xmmword ptr [rbp - 0x150]
  3a3420:      	mov	r13, qword ptr [rbp - 0x140]
  3a3427:      	test	r13, r13
  3a342a:      	je	0x3a34d5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcc5>
  3a3430:      	cmp	qword ptr [r13 + 0x8], 0x0
  3a3435:      	mov	r8d, 0xffffffff
  3a343b:      	mov	r9, qword ptr [rbp - 0x38]
  3a343f:      	mov	r10, qword ptr [rbp - 0x50]
  3a3443:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3447:      	je	0x3a34e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcd7>
  3a344d:      	jmp	0x3a4238 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a28>
  3a3452:      	mov	edi, dword ptr [r15 + 0x18]
  3a3456:      	cmp	qword ptr [rbp - 0x38], rdi
  3a345a:      	jbe	0x3a4366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a3460:      	shl	rdi, 0x4
  3a3464:      	mov	rax, qword ptr [rbp - 0x50]
  3a3468:      	mov	qword ptr [rax + rdi], r12
  3a346c:      	mov	qword ptr [rax + rdi + 0x8], rdx
  3a3471:      	mov	rax, r12
  3a3474:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3479:      	mov	edi, dword ptr [r15 + 0x14]
  3a347d:      	cmp	rdi, r8
  3a3480:      	je	0x3a3bdf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13cf>
  3a3486:      	cmp	qword ptr [rbp - 0x38], rdi
  3a348a:      	jbe	0x3a4366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a3490:      	shl	rdi, 0x4
  3a3494:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3498:      	mov	rcx, qword ptr [rcx + rdi]
  3a349c:      	test	ecx, ecx
  3a349e:      	cmove	rdx, rax
  3a34a2:      	cmove	r12, r10
  3a34a6:      	mov	rax, r12
  3a34a9:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a34ae:      	mov	qword ptr [rbp - 0x90], r10
  3a34b5:      	mov	qword ptr [rbp - 0x88], rax
  3a34bc:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a34c3:      	movups	xmmword ptr [rsp], xmm0
  3a34c7:      	mov	esi, r9d
  3a34ca:      	mov	rdx, r12
  3a34cd:      	xor	r8d, r8d
  3a34d0:      	jmp	0x3a3bb7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13a7>
  3a34d5:      	mov	r8d, 0xffffffff
  3a34db:      	mov	r9, qword ptr [rbp - 0x38]
  3a34df:      	mov	r10, qword ptr [rbp - 0x50]
  3a34e3:      	mov	rsi, qword ptr [rbp - 0x48]
  3a34e7:      	mov	edi, dword ptr [r15 + 0x8]
  3a34eb:      	cmp	rdi, r8
  3a34ee:      	mov	r11, qword ptr [rbp - 0x40]
  3a34f2:      	mov	r15, qword ptr [rbp - 0x70]
  3a34f6:      	je	0x3a2f07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6f7>
  3a34fc:      	cmp	r9, rdi
  3a34ff:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a3505:      	shl	rdi, 0x4
  3a3509:      	movups	xmmword ptr [r10 + rdi], xmm0
  3a350e:      	jmp	0x3a2f07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6f7>
  3a3513:      	mov	qword ptr [rsp], r11
  3a3517:      	mov	rsi, rcx
  3a351a:      	mov	rdx, r10
  3a351d:      	mov	rcx, rax
  3a3520:      	call	0x3cef10 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a3525:      	jmp	0x3a2de6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d6>
  3a352a:      	mov	edi, dword ptr [r15 + 0x1c]
  3a352e:      	dec	edi
  3a3530:      	cmp	edi, r12d
  3a3533:      	cmovae	edi, r12d
  3a3537:      	jmp	0x3a3bf2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13e2>
  3a353c:      	xor	edi, edi
  3a353e:      	test	r12, r12
  3a3541:      	setne	dil
  3a3545:      	jmp	0x3a3bf2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13e2>
  3a354a:      	xor	eax, eax
  3a354c:      	cmp	r12, r10
  3a354f:      	sete	al
  3a3552:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3557:      	mov	rax, qword ptr [rbp - 0x60]
  3a355b:      	mov	rsi, qword ptr [rax + 0x78]
  3a355f:      	mov	edi, dword ptr [r15 + 0x18]
  3a3563:      	cmp	rsi, rdi
  3a3566:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a356c:      	shl	rdi, 0x3
  3a3570:      	add	rdi, qword ptr [rax + 0x70]
  3a3574:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a357a:      	test	r11, r11
  3a357d:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3583:      	mov	r13, qword ptr [rdi]
  3a3586:      	test	r13, r13
  3a3589:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a358f:      	mov	rdi, qword ptr [r11]
  3a3592:      	mov	rax, qword ptr [r13 + 0x10]
  3a3596:      	movzx	eax, byte ptr [rax + 0x31]
  3a359a:      	mov	byte ptr [rbp - 0x71], al
  3a359d:      	mov	edx, r12d
  3a35a0:      	test	al, 0x1
  3a35a2:      	cmovne	rdx, r12
  3a35a6:      	mov	rsi, r13
  3a35a9:      	call	0x3c0fe0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a35ae:      	test	dl, 0x1
  3a35b1:      	jne	0x3a2de4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d4>
  3a35b7:      	mov	rax, qword ptr [r13 + 0x10]
  3a35bb:      	movzx	eax, byte ptr [rax + 0x31]
  3a35bf:      	mov	byte ptr [rbp - 0x72], al
  3a35c2:      	and	al, 0x1
  3a35c4:      	neg	al
  3a35c6:      	mov	eax, 0x0
  3a35cb:      	sbb	rax, rax
  3a35ce:      	mov	r8d, 0xffffffff
  3a35d4:      	or	rax, r8
  3a35d7:      	xor	edx, edx
  3a35d9:      	mov	r11, qword ptr [rbp - 0x40]
  3a35dd:      	mov	rsi, qword ptr [rbp - 0x48]
  3a35e1:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a35e6:      	mov	rax, qword ptr [rbp - 0x60]
  3a35ea:      	mov	rsi, qword ptr [rax + 0x60]
  3a35ee:      	mov	edi, dword ptr [r15 + 0x18]
  3a35f2:      	cmp	rsi, rdi
  3a35f5:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a35fb:      	mov	rax, qword ptr [rax + 0x58]
  3a35ff:      	lea	rcx, [rax + 8*rdi]
  3a3603:      	test	rcx, rcx
  3a3606:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a360c:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3610:      	test	rax, rax
  3a3613:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3619:      	cmp	byte ptr [rax + 0x29], 0x0
  3a361d:      	mov	ecx, r12d
  3a3620:      	cmovne	rcx, r12
  3a3624:      	cmp	rcx, qword ptr [rax + 0x40]
  3a3628:      	jae	0x3a428f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7f>
  3a362e:      	mov	rax, qword ptr [rax + 0x38]
  3a3632:      	mov	rax, qword ptr [rax + 8*rcx]
  3a3636:      	jmp	0x3a40dd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cd>
  3a363b:      	mov	rax, qword ptr [rbp - 0x60]
  3a363f:      	mov	rsi, qword ptr [rax + 0x48]
  3a3643:      	mov	edi, dword ptr [r15 + 0x18]
  3a3647:      	cmp	rsi, rdi
  3a364a:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3650:      	mov	rax, qword ptr [rax + 0x40]
  3a3654:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3658:      	jmp	0x3a40dd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cd>
  3a365d:      	mov	rax, r12
  3a3660:      	test	r12, r12
  3a3663:      	jne	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3669:      	jmp	0x3a4299 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a89>
  3a366e:      	mov	rax, qword ptr [rbp - 0x60]
  3a3672:      	mov	rsi, qword ptr [rax + 0x90]
  3a3679:      	mov	edi, dword ptr [r15 + 0x18]
  3a367d:      	cmp	rsi, rdi
  3a3680:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3686:      	mov	rax, qword ptr [rax + 0x88]
  3a368d:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3691:      	mov	qword ptr [rax + 0x20], r12
  3a3695:      	mov	qword ptr [rax + 0x28], rdx
  3a3699:      	jmp	0x3a3722 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf12>
  3a369e:      	mov	rax, qword ptr [rbp - 0x60]
  3a36a2:      	mov	rsi, qword ptr [rax + 0x90]
  3a36a9:      	mov	edi, dword ptr [r15 + 0x18]
  3a36ad:      	cmp	rsi, rdi
  3a36b0:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a36b6:      	mov	rax, qword ptr [rax + 0x88]
  3a36bd:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a36c1:      	mov	rax, qword ptr [rcx + 0x20]
  3a36c5:      	mov	rdx, qword ptr [rcx + 0x28]
  3a36c9:      	mov	rsi, qword ptr [rbp - 0x48]
  3a36cd:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a36d2:      	mov	rax, qword ptr [rbp - 0x60]
  3a36d6:      	mov	rsi, qword ptr [rax + 0x60]
  3a36da:      	mov	edi, dword ptr [r15 + 0x18]
  3a36de:      	cmp	rsi, rdi
  3a36e1:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a36e7:      	mov	rax, qword ptr [rax + 0x58]
  3a36eb:      	lea	rcx, [rax + 8*rdi]
  3a36ef:      	test	rcx, rcx
  3a36f2:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a36f8:      	mov	rax, qword ptr [rax + 8*rdi]
  3a36fc:      	test	rax, rax
  3a36ff:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3705:      	cmp	byte ptr [rax + 0x29], 0x0
  3a3709:      	mov	ecx, r12d
  3a370c:      	cmovne	rcx, r12
  3a3710:      	cmp	rcx, qword ptr [rax + 0x40]
  3a3714:      	jae	0x3a428f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7f>
  3a371a:      	mov	rax, qword ptr [rax + 0x38]
  3a371e:      	mov	qword ptr [rax + 8*rcx], r10
  3a3722:      	xor	edx, edx
  3a3724:      	xor	eax, eax
  3a3726:      	mov	rsi, qword ptr [rbp - 0x48]
  3a372a:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a372f:      	mov	edi, dword ptr [r15 + 0x18]
  3a3733:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3737:      	jbe	0x3a4366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a373d:      	shl	rdi, 0x4
  3a3741:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3745:      	mov	rax, qword ptr [rcx + rdi]
  3a3749:      	mov	rdx, qword ptr [rcx + rdi + 0x8]
  3a374e:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3753:      	xor	eax, eax
  3a3755:      	test	r12, r12
  3a3758:      	sete	al
  3a375b:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3760:      	mov	rax, qword ptr [rbp - 0x60]
  3a3764:      	mov	rsi, qword ptr [rax + 0x78]
  3a3768:      	mov	edi, dword ptr [r15 + 0x18]
  3a376c:      	cmp	rsi, rdi
  3a376f:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3775:      	shl	rdi, 0x3
  3a3779:      	add	rdi, qword ptr [rax + 0x70]
  3a377d:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3783:      	mov	rax, qword ptr [rdi]
  3a3786:      	test	rax, rax
  3a3789:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a378f:      	mov	rdi, qword ptr [rax + 0x10]
  3a3793:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a3798:      	mov	rax, qword ptr [rdi + 0x60]
  3a379c:      	je	0x3a3f71 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1761>
  3a37a2:      	test	rax, rax
  3a37a5:      	jne	0x3a3f7a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x176a>
  3a37ab:      	call	0x3c4a70 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a37b0:      	mov	r8d, 0xffffffff
  3a37b6:      	mov	r11, qword ptr [rbp - 0x40]
  3a37ba:      	jmp	0x3a40dd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cd>
  3a37bf:      	mov	eax, dword ptr [r15 + 0x18]
  3a37c3:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a37c8:      	xor	r10, r12
  3a37cb:      	jmp	0x3a3b2c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a37d0:      	mov	eax, r12d
  3a37d3:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a37d8:      	mov	ecx, r10d
  3a37db:      	shr	r12, cl
  3a37de:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a37e3:      	mov	ecx, r10d
  3a37e6:      	shr	r12d, cl
  3a37e9:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a37ee:      	or	r10, r12
  3a37f1:      	jmp	0x3a3b2c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a37f6:      	mov	ecx, r10d
  3a37f9:      	sar	r12d, cl
  3a37fc:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3801:      	xor	r10d, r12d
  3a3804:      	jmp	0x3a3b2c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a3809:      	mov	rax, r12
  3a380c:      	shr	rax
  3a380f:      	movabs	rcx, 0x5555555555555555
  3a3819:      	and	rax, rcx
  3a381c:      	and	r12, rcx
  3a381f:      	add	r12, rax
  3a3822:      	mov	rax, r12
  3a3825:      	shr	rax, 0x2
  3a3829:      	movabs	rcx, 0x3333333333333333
  3a3833:      	and	rax, rcx
  3a3836:      	and	r12, rcx
  3a3839:      	add	r12, rax
  3a383c:      	mov	rax, r12
  3a383f:      	shr	rax, 0x4
  3a3843:      	add	rax, r12
  3a3846:      	movabs	rcx, 0xf0f0f0f0f0f0f0f
  3a3850:      	and	rcx, rax
  3a3853:      	mov	rax, rcx
  3a3856:      	shr	rax, 0x8
  3a385a:      	add	rax, rcx
  3a385d:      	mov	rcx, rax
  3a3860:      	shr	rcx, 0x10
  3a3864:      	add	rcx, rax
  3a3867:      	mov	rax, rcx
  3a386a:      	shr	rax, 0x20
  3a386e:      	add	eax, ecx
  3a3870:      	and	eax, 0x7f
  3a3873:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3878:      	xor	eax, eax
  3a387a:      	cmp	r12, r10
  3a387d:      	setb	al
  3a3880:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3885:      	xor	eax, eax
  3a3887:      	cmp	r12, r10
  3a388a:      	setbe	al
  3a388d:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3892:      	xor	eax, eax
  3a3894:      	cmp	r12, r10
  3a3897:      	seta	al
  3a389a:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a389f:      	imul	r10d, r12d
  3a38a3:      	jmp	0x3a3b2c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a38a8:      	mov	ecx, r12d
  3a38ab:      	shr	ecx, 0x10
  3a38ae:      	xor	eax, eax
  3a38b0:      	cmp	r12d, 0x10000
  3a38b7:      	setb	al
  3a38ba:      	cmovb	ecx, r12d
  3a38be:      	shl	eax, 0x4
  3a38c1:      	mov	edx, ecx
  3a38c3:      	shr	edx, 0x8
  3a38c6:      	xor	esi, esi
  3a38c8:      	cmp	ecx, 0x100
  3a38ce:      	setb	sil
  3a38d2:      	lea	eax, [rax + 8*rsi]
  3a38d5:      	mov	rsi, qword ptr [rbp - 0x48]
  3a38d9:      	cmovb	edx, ecx
  3a38dc:      	movzx	ecx, dl
  3a38df:      	movzx	ecx, byte ptr [rcx + 0x3282e9]
  3a38e6:      	sub	rax, rcx
  3a38e9:      	add	rax, 0x8
  3a38ed:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a38f2:      	mov	ecx, r10d
  3a38f5:      	rol	r12d, cl
  3a38f8:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a38fd:      	xor	eax, eax
  3a38ff:      	cmp	r12, r10
  3a3902:      	setg	al
  3a3905:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a390a:      	xor	eax, eax
  3a390c:      	cmp	r12d, r10d
  3a390f:      	setl	al
  3a3912:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3917:      	xor	eax, eax
  3a3919:      	cmp	r12, r10
  3a391c:      	setne	al
  3a391f:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3924:      	test	r12d, r12d
  3a3927:      	je	0x3a3fc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17b2>
  3a392d:      	tzcnt	eax, r12d
  3a3932:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3937:      	add	r10d, r12d
  3a393a:      	jmp	0x3a3b2c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a393f:      	xor	eax, eax
  3a3941:      	cmp	r12d, r10d
  3a3944:      	setne	al
  3a3947:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a394c:      	xor	eax, eax
  3a394e:      	cmp	r12, r10
  3a3951:      	setle	al
  3a3954:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3959:      	mov	ecx, r10d
  3a395c:      	ror	r12d, cl
  3a395f:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3964:      	xor	eax, eax
  3a3966:      	cmp	r12d, r10d
  3a3969:      	setbe	al
  3a396c:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3971:      	xor	eax, eax
  3a3973:      	cmp	r12d, r10d
  3a3976:      	setge	al
  3a3979:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a397e:      	xor	eax, eax
  3a3980:      	test	r12d, r12d
  3a3983:      	sete	al
  3a3986:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a398b:      	xor	eax, eax
  3a398d:      	cmp	r12d, r10d
  3a3990:      	setb	al
  3a3993:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3998:      	mov	rdx, r12
  3a399b:      	mov	eax, edx
  3a399d:      	shr	eax, 0x18
  3a39a0:      	mov	ecx, edx
  3a39a2:      	shr	ecx, 0x10
  3a39a5:      	movzx	ecx, cl
  3a39a8:      	movzx	ecx, byte ptr [rcx + 0x3283e9]
  3a39af:      	add	cl, byte ptr [rax + 0x3283e9]
  3a39b5:      	movzx	eax, dh
  3a39b8:      	add	cl, byte ptr [rax + 0x3283e9]
  3a39be:      	movzx	eax, dl
  3a39c1:      	add	cl, byte ptr [rax + 0x3283e9]
  3a39c7:      	movzx	eax, cl
  3a39ca:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a39cf:      	add	r10, r12
  3a39d2:      	jmp	0x3a3b2c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a39d7:      	mov	ecx, r10d
  3a39da:      	shl	r12d, cl
  3a39dd:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a39e2:      	xor	eax, eax
  3a39e4:      	test	r12, r12
  3a39e7:      	sete	al
  3a39ea:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a39ef:      	and	r10, r12
  3a39f2:      	jmp	0x3a3b2c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a39f7:      	sub	r12, r10
  3a39fa:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a39ff:      	xor	eax, eax
  3a3a01:      	cmp	r12d, r10d
  3a3a04:      	sete	al
  3a3a07:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3a0c:      	mov	ecx, r10d
  3a3a0f:      	rol	r12, cl
  3a3a12:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3a17:      	xor	eax, eax
  3a3a19:      	cmp	r12d, r10d
  3a3a1c:      	setae	al
  3a3a1f:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3a24:      	and	r12d, r10d
  3a3a27:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3a2c:      	xor	eax, eax
  3a3a2e:      	cmp	r12d, r10d
  3a3a31:      	setle	al
  3a3a34:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3a39:      	xor	eax, eax
  3a3a3b:      	cmp	r12d, r10d
  3a3a3e:      	setg	al
  3a3a41:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3a46:      	xor	eax, eax
  3a3a48:      	mov	rcx, r12
  3a3a4b:      	shr	rcx, 0x20
  3a3a4f:      	cmove	rcx, r12
  3a3a53:      	setne	al
  3a3a56:      	shl	eax, 0x5
  3a3a59:      	mov	rdx, rcx
  3a3a5c:      	shr	rdx, 0x10
  3a3a60:      	lea	rsi, [rax + 0x10]
  3a3a64:      	cmp	rcx, 0x10000
  3a3a6b:      	cmovb	rdx, rcx
  3a3a6f:      	cmovb	rsi, rax
  3a3a73:      	mov	eax, edx
  3a3a75:      	shr	eax, 0x8
  3a3a78:      	lea	rcx, [rsi + 0x8]
  3a3a7c:      	cmp	rdx, 0x100
  3a3a83:      	cmovb	rax, rdx
  3a3a87:      	cmovb	rcx, rsi
  3a3a8b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3a8f:      	movzx	eax, al
  3a3a92:      	movzx	edx, byte ptr [rax + 0x3282e9]
  3a3a99:      	add	rdx, rcx
  3a3a9c:      	mov	eax, 0x40
  3a3aa1:      	sub	rax, rdx
  3a3aa4:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3aa9:      	xor	eax, eax
  3a3aab:      	cmp	r12, r10
  3a3aae:      	setl	al
  3a3ab1:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3ab6:      	mov	ecx, r10d
  3a3ab9:      	ror	r12, cl
  3a3abc:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3ac1:      	xor	eax, eax
  3a3ac3:      	cmp	r12d, r10d
  3a3ac6:      	seta	al
  3a3ac9:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3ace:      	mov	eax, r12d
  3a3ad1:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3ad6:      	movsx	rax, r12w
  3a3ada:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3adf:      	movsxd	rax, r12d
  3a3ae2:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3ae7:      	xor	eax, eax
  3a3ae9:      	cmp	r12, r10
  3a3aec:      	sete	al
  3a3aef:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3af4:      	movsxd	rax, r12d
  3a3af7:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3afc:      	xor	eax, eax
  3a3afe:      	cmp	r12, r10
  3a3b01:      	setae	al
  3a3b04:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b09:      	xor	eax, eax
  3a3b0b:      	cmp	r12, r10
  3a3b0e:      	setge	al
  3a3b11:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b16:      	mov	ecx, r10d
  3a3b19:      	sar	r12, cl
  3a3b1c:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3b1e:      	or	r10d, r12d
  3a3b21:      	jmp	0x3a3b2c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a3b23:      	sub	r12d, r10d
  3a3b26:      	jmp	0x3a3b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3b28:      	imul	r10, r12
  3a3b2c:      	xor	edx, edx
  3a3b2e:      	mov	rax, r10
  3a3b31:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3b36:      	movsx	rax, r12w
  3a3b3a:      	mov	eax, eax
  3a3b3c:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b41:      	mov	eax, 0x40
  3a3b46:      	tzcnt	rax, r12
  3a3b4b:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b50:      	mov	ecx, r10d
  3a3b53:      	shl	r12, cl
  3a3b56:      	xor	edx, edx
  3a3b58:      	mov	rax, r12
  3a3b5b:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3b60:      	movsx	rax, r12b
  3a3b64:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b69:      	movsx	rax, r12b
  3a3b6d:      	mov	eax, eax
  3a3b6f:      	jmp	0x3a3fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b74:      	movzx	ecx, r9b
  3a3b78:      	add	ecx, -0x61
  3a3b7b:      	cmp	ecx, 0x5e
  3a3b7e:      	ja	0x3a3df5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15e5>
  3a3b84:      	xor	edx, edx
  3a3b86:      	xor	eax, eax
  3a3b88:      	jmp	qword ptr [8*rcx + 0x203eb8]
  3a3b8f:      	mov	qword ptr [rbp - 0x90], rax
  3a3b96:      	mov	qword ptr [rbp - 0x88], r8
  3a3b9d:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a3ba4:      	movups	xmmword ptr [rsp], xmm0
  3a3ba8:      	movzx	r8d, dil
  3a3bac:      	mov	rdi, rsi
  3a3baf:      	mov	esi, edx
  3a3bb1:      	mov	rdx, rcx
  3a3bb4:      	mov	rcx, r9
  3a3bb7:      	call	0x3c8640 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a3bbc:      	mov	r13d, eax
  3a3bbf:      	xor	edx, edx
  3a3bc1:      	xor	eax, eax
  3a3bc3:      	mov	r11, qword ptr [rbp - 0x40]
  3a3bc7:      	mov	r8d, 0xffffffff
  3a3bcd:      	test	r13b, r13b
  3a3bd0:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3bd4:      	je	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3bda:      	jmp	0x3a429c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8c>
  3a3bdf:      	mov	rdx, rax
  3a3be2:      	mov	rax, r10
  3a3be5:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3bea:      	mov	r9, qword ptr [rbp - 0x38]
  3a3bee:      	mov	r10, qword ptr [rbp - 0x50]
  3a3bf2:      	mov	rax, qword ptr [rbp - 0x58]
  3a3bf6:      	mov	rsi, qword ptr [rax + 0xe8]
  3a3bfd:      	add	edi, dword ptr [r15 + 0x18]
  3a3c01:      	cmp	rsi, rdi
  3a3c04:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3c0a:      	mov	rax, qword ptr [rax + 0xe0]
  3a3c11:      	lea	rcx, [rdi + 2*rdi]
  3a3c15:      	mov	r13d, dword ptr [rax + 4*rcx]
  3a3c19:      	mov	r15d, dword ptr [rax + 4*rcx + 0x4]
  3a3c1e:      	mov	esi, dword ptr [rax + 4*rcx + 0x8]
  3a3c22:      	cmp	r13d, r14d
  3a3c25:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a3c2a:      	ja	0x3a3cd5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c5>
  3a3c30:      	mov	rax, qword ptr [r11 + 0x160]
  3a3c37:      	test	rax, rax
  3a3c3a:      	je	0x3a3c4b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x143b>
  3a3c3c:      	test	r11, r11
  3a3c3f:      	je	0x3a3cae <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x149e>
  3a3c41:      	mov	eax, dword ptr [rax]
  3a3c43:      	test	eax, eax
  3a3c45:      	je	0x3a3fb0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17a0>
  3a3c4b:      	mov	rax, qword ptr [r11 + 0x8]
  3a3c4f:      	test	rax, rax
  3a3c52:      	je	0x3a444e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3e>
  3a3c58:      	mov	rdi, qword ptr [r11 + 0x10]
  3a3c5c:      	mov	rcx, qword ptr [rax + 0x8]
  3a3c60:      	test	byte ptr [rcx + 0x17], 0x20
  3a3c64:      	jne	0x3a3c6b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x145b>
  3a3c66:      	mov	r14, rsi
  3a3c69:      	jmp	0x3a3c85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1475>
  3a3c6b:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a3c6f:      	and	cl, 0x1f
  3a3c72:      	mov	r14, rsi
  3a3c75:      	cmp	cl, 0x16
  3a3c78:      	je	0x3a3c85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1475>
  3a3c7a:      	mov	qword ptr [rbp - 0x90], rdi
  3a3c81:      	mov	rdi, qword ptr [rbp - 0x68]
  3a3c85:      	call	qword ptr [rax + 0x28]
  3a3c88:      	mov	r8d, 0xffffffff
  3a3c8e:      	mov	r9, qword ptr [rbp - 0x38]
  3a3c92:      	mov	r10, qword ptr [rbp - 0x50]
  3a3c96:      	mov	rsi, r14
  3a3c99:      	test	rax, rax
  3a3c9c:      	je	0x3a3cd1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c1>
  3a3c9e:      	cmp	qword ptr [rax + 0x8], 0x0
  3a3ca3:      	mov	r11, qword ptr [rbp - 0x40]
  3a3ca7:      	je	0x3a3cd5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c5>
  3a3ca9:      	jmp	0x3a4228 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a18>
  3a3cae:      	mov	edi, 0x1
  3a3cb3:      	mov	r14, rsi
  3a3cb6:      	call	0x5e7240 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a3cbb:      	mov	rsi, r14
  3a3cbe:      	mov	r10, qword ptr [rbp - 0x50]
  3a3cc2:      	mov	r9, qword ptr [rbp - 0x38]
  3a3cc6:      	mov	r8d, 0xffffffff
  3a3ccc:      	test	rax, rax
  3a3ccf:      	jne	0x3a3c9e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x148e>
  3a3cd1:      	mov	r11, qword ptr [rbp - 0x40]
  3a3cd5:      	mov	rax, qword ptr [rbp - 0x58]
  3a3cd9:      	mov	rax, qword ptr [rax + 0x108]
  3a3ce0:      	lea	edx, [rsi + r15]
  3a3ce4:      	cmp	rax, rdx
  3a3ce7:      	jb	0x3a443c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c2c>
  3a3ced:      	cmp	r15d, edx
  3a3cf0:      	ja	0x3a442d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c1d>
  3a3cf6:      	sub	rax, r15
  3a3cf9:      	lea	rcx, [8*r15]
  3a3d01:      	test	rax, rax
  3a3d04:      	mov	eax, 0x0
  3a3d09:      	cmovle	rcx, rax
  3a3d0d:      	test	esi, esi
  3a3d0f:      	je	0x3a3d63 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1553>
  3a3d11:      	mov	rax, qword ptr [rbp - 0x58]
  3a3d15:      	mov	rax, qword ptr [rax + 0xf8]
  3a3d1c:      	sub	rdx, r15
  3a3d1f:      	add	rcx, rax
  3a3d22:      	add	rcx, 0x4
  3a3d26:      	xor	esi, esi
  3a3d28:      	nop	dword ptr [rax + rax]
  3a3d30:      	mov	edi, dword ptr [rcx + 8*rsi]
  3a3d33:      	cmp	r9, rdi
  3a3d36:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a3d3c:      	mov	eax, dword ptr [rcx + 8*rsi - 0x4]
  3a3d40:      	cmp	r9, rax
  3a3d43:      	jbe	0x3a435b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b4b>
  3a3d49:      	shl	rdi, 0x4
  3a3d4d:      	shl	rax, 0x4
  3a3d51:      	movups	xmm0, xmmword ptr [r10 + rdi]
  3a3d56:      	movups	xmmword ptr [r10 + rax], xmm0
  3a3d5b:      	inc	rsi
  3a3d5e:      	cmp	rdx, rsi
  3a3d61:      	jne	0x3a3d30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1520>
  3a3d63:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3d67:      	jmp	0x3a291b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10b>
  3a3d6c:      	cmp	ecx, 0x3
  3a3d6f:      	je	0x3a415f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x194f>
  3a3d75:      	cmp	ecx, 0x6
  3a3d78:      	je	0x3a4180 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1970>
  3a3d7e:      	lea	ecx, [r9 - 0x12]
  3a3d82:      	cmp	ecx, 0x4
  3a3d85:      	jb	0x3a4135 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1925>
  3a3d8b:      	mov	edi, dword ptr [r15 + 0x14]
  3a3d8f:      	cmp	rdi, r8
  3a3d92:      	je	0x3a3dfe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15ee>
  3a3d94:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3d98:      	jbe	0x3a4366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a3d9e:      	shl	rdi, 0x4
  3a3da2:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3da6:      	movups	xmm0, xmmword ptr [rcx + rdi]
  3a3daa:      	mov	ecx, r9d
  3a3dad:      	and	ecx, 0xffff0000
  3a3db3:      	cmp	ecx, 0xfe0000
  3a3db9:      	jne	0x3a3e12 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1602>
  3a3dbb:      	movzx	r9d, r9w
  3a3dbf:      	cmp	r9d, 0x3
  3a3dc3:      	jne	0x3a4002 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17f2>
  3a3dc9:      	test	r11, r11
  3a3dcc:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3dd2:      	mov	rax, qword ptr [r11]
  3a3dd5:      	test	rax, rax
  3a3dd8:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3dde:      	mov	rdi, qword ptr [rax + 0x8]
  3a3de2:      	call	0x3c4010 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a3de7:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3deb:      	mov	r8d, 0xffffffff
  3a3df1:      	mov	r11, qword ptr [rbp - 0x40]
  3a3df5:      	xor	edx, edx
  3a3df7:      	xor	eax, eax
  3a3df9:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3dfe:      	xorps	xmm0, xmm0
  3a3e01:      	mov	ecx, r9d
  3a3e04:      	and	ecx, 0xffff0000
  3a3e0a:      	cmp	ecx, 0xfe0000
  3a3e10:      	je	0x3a3dbb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15ab>
  3a3e12:      	add	r9d, 0xff03fff8
  3a3e19:      	cmp	r9d, 0xa
  3a3e1d:      	jae	0x3a445d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c4d>
  3a3e23:      	mov	rdx, qword ptr [rbp - 0x58]
  3a3e27:      	mov	rsi, qword ptr [rdx + 0x80]
  3a3e2e:      	mov	edi, dword ptr [r15 + 0x4]
  3a3e32:      	cmp	rsi, rdi
  3a3e35:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3e3b:      	imul	rcx, rdi, 0xe0
  3a3e42:      	add	rcx, qword ptr [rdx + 0x78]
  3a3e46:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3e4c:      	movups	xmm1, xmmword ptr [rcx + 0x98]
  3a3e53:      	movups	xmm2, xmmword ptr [rcx + 0x88]
  3a3e5a:      	movups	xmm3, xmmword ptr [rcx + 0x78]
  3a3e5e:      	movups	xmm4, xmmword ptr [rcx + 0x68]
  3a3e62:      	movups	xmm5, xmmword ptr [rcx + 0x58]
  3a3e66:      	movups	xmm6, xmmword ptr [rcx + 0x48]
  3a3e6a:      	movzx	edx, byte ptr [rcx + 0x40]
  3a3e6e:      	movups	xmm7, xmmword ptr [rcx + 0x30]
  3a3e72:      	movups	xmm8, xmmword ptr [rcx + 0x20]
  3a3e77:      	movups	xmm9, xmmword ptr [rcx + 0x10]
  3a3e7c:      	mov	rsi, qword ptr [rcx + 0x8]
  3a3e80:      	mov	ecx, dword ptr [rcx]
  3a3e82:      	mov	dword ptr [rbp - 0x208], ecx
  3a3e88:      	mov	qword ptr [rbp - 0x200], rsi
  3a3e8f:      	movups	xmmword ptr [rbp - 0x1f8], xmm9
  3a3e97:      	movups	xmmword ptr [rbp - 0x1e8], xmm8
  3a3e9f:      	movups	xmmword ptr [rbp - 0x1d8], xmm7
  3a3ea6:      	mov	byte ptr [rbp - 0x1c8], dl
  3a3eac:      	movups	xmmword ptr [rbp - 0x1c0], xmm6
  3a3eb3:      	movups	xmmword ptr [rbp - 0x1b0], xmm5
  3a3eba:      	movups	xmmword ptr [rbp - 0x1a0], xmm4
  3a3ec1:      	movups	xmmword ptr [rbp - 0x190], xmm3
  3a3ec8:      	movups	xmmword ptr [rbp - 0x180], xmm2
  3a3ecf:      	movups	xmmword ptr [rbp - 0x170], xmm1
  3a3ed6:      	mov	qword ptr [rbp - 0xe0], r10
  3a3edd:      	mov	qword ptr [rbp - 0xd8], rax
  3a3ee4:      	movaps	xmmword ptr [rbp - 0x160], xmm0
  3a3eeb:      	mov	ecx, 0x15
  3a3ef0:      	mov	rdi, rsp
  3a3ef3:      	lea	rsi, [rbp - 0x208]
  3a3efa:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a3efd:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a3f04:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a3f0c:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a3f13:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a3f1b:      	lea	rdi, [rbp - 0x108]
  3a3f22:      	mov	rsi, r11
  3a3f25:      	mov	rcx, r12
  3a3f28:      	mov	rdx, qword ptr [rbp - 0x60]
  3a3f2c:      	call	0x3a9400 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a3f31:      	mov	rax, qword ptr [rbp - 0x108]
  3a3f38:      	mov	rdx, qword ptr [rbp - 0x100]
  3a3f3f:      	movzx	r13d, byte ptr [rbp - 0xf8]
  3a3f47:      	mov	rsi, qword ptr [rbp - 0xf0]
  3a3f4e:      	test	rsi, rsi
  3a3f51:      	je	0x3a40d4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18c4>
  3a3f57:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a3f5c:      	mov	r11, qword ptr [rbp - 0x40]
  3a3f60:      	mov	r8d, 0xffffffff
  3a3f66:      	je	0x3a3bcd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13bd>
  3a3f6c:      	jmp	0x3a4324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b14>
  3a3f71:      	test	rax, rax
  3a3f74:      	je	0x3a40e8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18d8>
  3a3f7a:      	mov	rax, qword ptr [rax]
  3a3f7d:      	test	rdi, rdi
  3a3f80:      	je	0x3a40f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18e5>
  3a3f86:      	mov	rcx, qword ptr [rdi + 0x38]
  3a3f8a:      	test	rcx, rcx
  3a3f8d:      	mov	edx, 0x10000
  3a3f92:      	cmove	rcx, rdx
  3a3f96:      	mov	rdx, rax
  3a3f99:      	or	rdx, rcx
  3a3f9c:      	shr	rdx, 0x20
  3a3fa0:      	je	0x3a40d9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18c9>
  3a3fa6:      	xor	edx, edx
  3a3fa8:      	div	rcx
  3a3fab:      	jmp	0x3a40dd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cd>
  3a3fb0:      	xor	eax, eax
  3a3fb2:      	xor	edx, edx
  3a3fb4:      	test	rax, rax
  3a3fb7:      	jne	0x3a3c9e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x148e>
  3a3fbd:      	jmp	0x3a3cd1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c1>
  3a3fc2:      	mov	eax, 0x20
  3a3fc7:      	nop	word ptr [rax + rax]
  3a3fd0:      	xor	edx, edx
  3a3fd2:      	mov	edi, dword ptr [r15 + 0x8]
  3a3fd6:      	cmp	rdi, r8
  3a3fd9:      	mov	r9, qword ptr [rbp - 0x38]
  3a3fdd:      	mov	r10, qword ptr [rbp - 0x50]
  3a3fe1:      	je	0x3a2910 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a3fe7:      	cmp	r9, rdi
  3a3fea:      	jbe	0x3a433d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a3ff0:      	shl	rdi, 0x4
  3a3ff4:      	mov	qword ptr [r10 + rdi], rax
  3a3ff8:      	mov	qword ptr [r10 + rdi + 0x8], rdx
  3a3ffd:      	jmp	0x3a2910 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a4002:      	mov	rcx, qword ptr [rbp - 0x58]
  3a4006:      	mov	rsi, qword ptr [rcx + 0x80]
  3a400d:      	mov	edi, dword ptr [r15 + 0x4]
  3a4011:      	cmp	rsi, rdi
  3a4014:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a401a:      	imul	rax, rdi, 0xe0
  3a4021:      	add	rax, qword ptr [rcx + 0x78]
  3a4025:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a402b:      	mov	rcx, qword ptr [rbp - 0x60]
  3a402f:      	mov	rsi, qword ptr [rcx + 0x78]
  3a4033:      	mov	edi, dword ptr [rax + 0x18]
  3a4036:      	cmp	rsi, rdi
  3a4039:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a403f:      	shl	rdi, 0x3
  3a4043:      	add	rdi, qword ptr [rcx + 0x70]
  3a4047:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a404d:      	mov	rsi, qword ptr [rdi]
  3a4050:      	test	rsi, rsi
  3a4053:      	je	0x3a41c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a4059:      	mov	rcx, qword ptr [r11]
  3a405c:      	mov	rdx, qword ptr [r11 + 0x8]
  3a4060:      	mov	r8, qword ptr [rcx + 0x8]
  3a4064:      	mov	rcx, qword ptr [r11 + 0x10]
  3a4068:      	mov	rsi, qword ptr [rsi + 0x10]
  3a406c:      	mov	edi, dword ptr [rax + 0x14]
  3a406f:      	mov	eax, dword ptr [rax + 0x1c]
  3a4072:      	shl	rax, 0x20
  3a4076:      	or	rax, rdi
  3a4079:      	movlps	qword ptr [rsp + 0x18], xmm0
  3a407e:      	mov	qword ptr [rsp + 0x10], r10
  3a4083:      	mov	qword ptr [rsp + 0x8], rax
  3a4088:      	mov	qword ptr [rsp], r12
  3a408c:      	lea	rdi, [rbp - 0x130]
  3a4093:      	call	0x3c5320 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a4098:      	mov	rax, qword ptr [rbp - 0x130]
  3a409f:      	mov	rdx, qword ptr [rbp - 0x128]
  3a40a6:      	movzx	r13d, byte ptr [rbp - 0x120]
  3a40ae:      	mov	rsi, qword ptr [rbp - 0x118]
  3a40b5:      	test	rsi, rsi
  3a40b8:      	je	0x3a4114 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1904>
  3a40ba:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a40bf:      	mov	r11, qword ptr [rbp - 0x40]
  3a40c3:      	mov	r8d, 0xffffffff
  3a40c9:      	je	0x3a3bcd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13bd>
  3a40cf:      	jmp	0x3a432d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b1d>
  3a40d4:      	jmp	0x3a3bc3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13b3>
  3a40d9:      	xor	edx, edx
  3a40db:      	div	ecx
  3a40dd:      	xor	edx, edx
  3a40df:      	mov	rsi, qword ptr [rbp - 0x48]
  3a40e3:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a40e8:      	mov	rax, qword ptr [rdi + 0x8]
  3a40ec:      	test	rdi, rdi
  3a40ef:      	jne	0x3a3f86 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1776>
  3a40f5:      	mov	edi, 0x1
  3a40fa:      	call	0x5e7240 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a40ff:      	mov	r8d, 0xffffffff
  3a4105:      	mov	r11, qword ptr [rbp - 0x40]
  3a4109:      	xor	edx, edx
  3a410b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a410f:      	jmp	0x3a3fd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a4114:      	jmp	0x3a3bc3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13b3>
  3a4119:      	mov	rax, qword ptr [rbp - 0x70]
  3a411d:      	mov	dword ptr [rax - 0x38], r14d
  3a4121:      	mov	rdi, r11
  3a4124:      	mov	r14, r11
  3a4127:      	call	0x3a6840 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a412c:      	test	al, al
  3a412e:      	je	0x3a41a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1990>
  3a4130:      	jmp	0x3a430f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a4135:      	mov	rax, qword ptr [rbp - 0x70]
  3a4139:      	mov	dword ptr [rax - 0x38], r14d
  3a413d:      	mov	rdi, r11
  3a4140:      	mov	r14, r11
  3a4143:      	call	0x3a55e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a4148:      	jmp	0x3a4193 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1983>
  3a414a:      	mov	rax, qword ptr [rbp - 0x70]
  3a414e:      	mov	dword ptr [rax - 0x38], r14d
  3a4152:      	mov	rdi, r11
  3a4155:      	mov	r14, r11
  3a4158:      	call	0x3a6340 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a415d:      	jmp	0x3a4193 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1983>
  3a415f:      	mov	rax, qword ptr [rbp - 0x70]
  3a4163:      	mov	dword ptr [rax - 0x38], r14d
  3a4167:      	mov	rdi, r11
  3a416a:      	mov	r14, r11
  3a416d:      	call	0x3a5150 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a4172:      	test	al, al
  3a4174:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a4179:      	je	0x3a41a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1990>
  3a417b:      	jmp	0x3a4339 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b29>
  3a4180:      	mov	rax, qword ptr [rbp - 0x70]
  3a4184:      	mov	dword ptr [rax - 0x38], r14d
  3a4188:      	mov	rdi, r11
  3a418b:      	mov	r14, r11
  3a418e:      	call	0x3a6840 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a4193:      	test	al, al
  3a4195:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a419a:      	jne	0x3a430f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a41a0:      	mov	rcx, qword ptr [r14 + 0x48]
  3a41a4:      	test	rcx, rcx
  3a41a7:      	mov	r11, r14
  3a41aa:      	mov	r8d, 0xffffffff
  3a41b0:      	mov	r14, qword ptr [rbp - 0x68]
  3a41b4:      	jg	0x3a2851 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x41>
  3a41ba:      	mov	edi, 0x333878
  3a41bf:      	mov	esi, 0x299e80
  3a41c4:      	call	0x5e2bf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a41c9:      	mov	edi, 0x1
  3a41ce:      	call	0x5e7240 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a41d3:      	jmp	0x3a430f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a41d8:      	mov	edi, 0x28
  3a41dd:      	call	0x5fd850 <GC_malloc@plt>
  3a41e2:      	test	rax, rax
  3a41e5:      	je	0x3a446c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a41eb:      	mov	r14, rax
  3a41ee:      	add	r13, r12
  3a41f1:      	mov	edi, 0x28
  3a41f6:      	call	0x5e2ed0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a41fb:      	mov	rax, qword ptr [r13 + 0x8]
  3a41ff:      	movaps	xmm0, xmmword ptr [rip + 0x31701a] # 0x6bb220 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a4206:      	movups	xmmword ptr [r14], xmm0
  3a420a:      	mov	qword ptr [r14 + 0x10], rax
  3a420e:      	mov	qword ptr [r14 + 0x18], 0x299e90
  3a4216:      	mov	qword ptr [r14 + 0x20], 0xe
  3a421e:      	mov	esi, 0x2ba890
  3a4223:      	jmp	0x3a4301 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af1>
  3a4228:      	mov	rdi, r11
  3a422b:      	mov	rsi, rax
  3a422e:      	jmp	0x3a4308 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af8>
  3a4233:      	mov	r13d, edx
  3a4236:      	jmp	0x3a429c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8c>
  3a4238:      	mov	r14, qword ptr [rbp - 0x138]
  3a423f:      	mov	edi, 0x8
  3a4244:      	call	0x5fd850 <GC_malloc@plt>
  3a4249:      	test	rax, rax
  3a424c:      	je	0x3a446c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a4252:      	mov	r15, rax
  3a4255:      	mov	edi, 0x8
  3a425a:      	call	0x5e2ed0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a425f:      	mov	edx, 0x299e00
  3a4264:      	mov	rdi, r13
  3a4267:      	mov	rsi, r14
  3a426a:      	mov	rcx, r15
  3a426d:      	call	0x5b9de0 <errors.As>
  3a4272:      	test	al, 0x1
  3a4274:      	je	0x3a4286 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a76>
  3a4276:      	mov	rax, qword ptr [r15]
  3a4279:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a4280:      	mov	ecx, dword ptr [rcx + 0x50]
  3a4283:      	mov	dword ptr [rax + 0x4], ecx
  3a4286:      	mov	rdi, qword ptr [rbp - 0x40]
  3a428a:      	mov	rsi, r13
  3a428d:      	jmp	0x3a4305 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af5>
  3a428f:      	mov	r13b, 0x3
  3a4292:      	jmp	0x3a429c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8c>
  3a4294:      	mov	r13b, 0x1
  3a4297:      	jmp	0x3a429c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8c>
  3a4299:      	mov	r13b, 0x9
  3a429c:      	mov	edi, 0x10
  3a42a1:      	call	0x5fd7e0 <GC_malloc_atomic@plt>
  3a42a6:      	test	rax, rax
  3a42a9:      	je	0x3a446c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a42af:      	mov	r14, rax
  3a42b2:      	mov	edi, 0x10
  3a42b7:      	call	0x5e2ed0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a42bc:      	xorps	xmm0, xmm0
  3a42bf:      	movups	xmmword ptr [r14], xmm0
  3a42c3:      	mov	rax, qword ptr [rbp - 0x58]
  3a42c7:      	mov	rsi, qword ptr [rax + 0x80]
  3a42ce:      	mov	edi, dword ptr [r15 + 0x4]
  3a42d2:      	cmp	rsi, rdi
  3a42d5:      	jbe	0x3a434b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a42d7:      	mov	rax, qword ptr [rax + 0x78]
  3a42db:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a42e2:      	mov	ecx, dword ptr [rcx + 0x50]
  3a42e5:      	imul	rdx, rdi, 0xe0
  3a42ec:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a42f1:      	mov	byte ptr [r14], r13b
  3a42f4:      	mov	dword ptr [r14 + 0x4], ecx
  3a42f8:      	mov	qword ptr [r14 + 0x8], rax
  3a42fc:      	mov	esi, 0x29a218
  3a4301:      	mov	rdi, qword ptr [rbp - 0x40]
  3a4305:      	mov	rdx, r14
  3a4308:      	call	0x3a21e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a430d:      	mov	al, 0x3
  3a430f:      	mov	rdx, rbx
  3a4312:      	add	rsp, 0x2a8
  3a4319:      	pop	rbx
  3a431a:      	pop	r12
  3a431c:      	pop	r13
  3a431e:      	pop	r14
  3a4320:      	pop	r15
  3a4322:      	pop	rbp
  3a4323:      	ret
  3a4324:      	mov	rdx, qword ptr [rbp - 0xe8]
  3a432b:      	jmp	0x3a4334 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b24>
  3a432d:      	mov	rdx, qword ptr [rbp - 0x110]
  3a4334:      	mov	rdi, r11
  3a4337:      	jmp	0x3a4308 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af8>
  3a4339:      	mov	al, 0x1
  3a433b:      	jmp	0x3a430f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a433d:      	mov	rsi, r9
  3a4340:      	call	0x5e2fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4345:      	mov	rdi, r14
  3a4348:      	mov	rsi, r13
  3a434b:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4350:      	mov	rdi, r15
  3a4353:      	mov	rsi, r13
  3a4356:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a435b:      	mov	rdi, rax
  3a435e:      	mov	rsi, r9
  3a4361:      	call	0x5e2fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4366:      	mov	rsi, qword ptr [rbp - 0x38]
  3a436a:      	call	0x5e2fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a436f:      	mov	rdi, r9
  3a4372:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4377:      	mov	rsi, qword ptr [rbp - 0x38]
  3a437b:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4380:      	mov	rdi, rax
  3a4383:      	mov	esi, 0x1
  3a4388:      	mov	ecx, 0x2
  3a438d:      	call	0x5e3cf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4392:      	mov	esi, 0x1
  3a4397:      	mov	rdx, rax
  3a439a:      	mov	ecx, 0x3
  3a439f:      	call	0x5e3cf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a43a4:      	mov	rdi, rax
  3a43a7:      	mov	rsi, r9
  3a43aa:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a43af:      	mov	rsi, rcx
  3a43b2:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a43b7:      	mov	rdi, rsi
  3a43ba:      	mov	rsi, rcx
  3a43bd:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a43c2:      	mov	rdi, rcx
  3a43c5:      	mov	rsi, r9
  3a43c8:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a43cd:      	mov	edi, 0x1
  3a43d2:      	mov	esi, 0x1
  3a43d7:      	call	0x5e2fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a43dc:      	xor	edi, edi
  3a43de:      	xor	esi, esi
  3a43e0:      	call	0x5e2fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a43e5:      	xor	esi, esi
  3a43e7:      	mov	rdx, rcx
  3a43ea:      	mov	ecx, 0x3
  3a43ef:      	call	0x5e3cf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a43f4:      	mov	rdi, rcx
  3a43f7:      	xor	esi, esi
  3a43f9:      	mov	ecx, 0x2
  3a43fe:      	call	0x5e3cf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4403:      	mov	rdi, rax
  3a4406:      	mov	rsi, qword ptr [rbp - 0x38]
  3a440a:      	call	0x5e44e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a440f:      	mov	edi, 0x3
  3a4414:      	mov	esi, 0x3
  3a4419:      	call	0x5e2fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a441e:      	mov	edi, 0x2
  3a4423:      	mov	esi, 0x2
  3a4428:      	call	0x5e2fb0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a442d:      	mov	rdi, r15
  3a4430:      	xor	esi, esi
  3a4432:      	mov	ecx, 0x3
  3a4437:      	call	0x5e3cf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a443c:      	mov	rdi, rdx
  3a443f:      	xor	esi, esi
  3a4441:      	mov	rdx, rax
  3a4444:      	mov	ecx, 0x2
  3a4449:      	call	0x5e3cf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a444e:      	mov	edi, 0x3293c8
  3a4453:      	mov	esi, 0x3296f8
  3a4458:      	call	0x5e2bf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a445d:      	mov	edi, 0x333878
  3a4462:      	mov	esi, 0x299ea0
  3a4467:      	call	0x5e2bf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a446c:      	mov	edi, 0x333878
  3a4471:      	mov	esi, 0x3293a0
  3a4476:      	call	0x5e2bf0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a447b:      	int3
  3a447c:      	int3
  3a447d:      	int3
  3a447e:      	int3
  3a447f:      	int3
