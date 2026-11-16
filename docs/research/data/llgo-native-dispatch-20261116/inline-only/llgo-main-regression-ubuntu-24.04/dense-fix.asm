
/home/runner/work/wasmvm/wasmvm/main-regression/dense-fix:	file format elf64-x86-64

Disassembly of section .text:

00000000003a2990 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a2990:      	push	rbp
  3a2991:      	mov	rbp, rsp
  3a2994:      	push	r15
  3a2996:      	push	r14
  3a2998:      	push	r13
  3a299a:      	push	r12
  3a299c:      	push	rbx
  3a299d:      	sub	rsp, 0x2b8
  3a29a4:      	mov	rcx, qword ptr [rdi + 0x48]
  3a29a8:      	test	rcx, rcx
  3a29ab:      	jle	0x3a456f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bdf>
  3a29b1:      	mov	rbx, rsi
  3a29b4:      	mov	r11, rdi
  3a29b7:      	movzx	r15d, byte ptr [rdi + 0x20]
  3a29bc:      	mov	r8d, 0xffffffff
  3a29c2:      	lea	rax, [rbp - 0x98]
  3a29c9:      	mov	qword ptr [rbp - 0x78], rax
  3a29cd:      	mov	qword ptr [rbp - 0x50], rdi
  3a29d1:      	mov	byte ptr [rbp - 0x29], r15b
  3a29d5:      	mov	r10, qword ptr [r11 + 0x40]
  3a29d9:      	shl	rcx, 0x6
  3a29dd:      	mov	byte ptr [r10 + rcx - 0x30], 0x0
  3a29e3:      	mov	rsi, qword ptr [r10 + rcx - 0x40]
  3a29e8:      	mov	rax, qword ptr [rsi + 0x48]
  3a29ec:      	mov	qword ptr [rbp - 0x68], rax
  3a29f0:      	mov	r9, qword ptr [rax + 0x98]
  3a29f7:      	mov	r12, qword ptr [r10 + rcx - 0x28]
  3a29fc:      	mov	rdx, qword ptr [r11 + 0x88]
  3a2a03:      	lea	rax, [r9 + r12]
  3a2a07:      	cmp	rax, rdx
  3a2a0a:      	ja	0x3a4496 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b06>
  3a2a10:      	cmp	r12, rax
  3a2a13:      	ja	0x3a44a8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b18>
  3a2a19:      	mov	rdi, qword ptr [rbp - 0x68]
  3a2a1d:      	mov	r14, qword ptr [rdi + 0xb8]
  3a2a24:      	mov	r13d, dword ptr [r10 + rcx - 0x38]
  3a2a29:      	sub	rdx, r12
  3a2a2c:      	shl	r12, 0x4
  3a2a30:      	test	rdx, rdx
  3a2a33:      	mov	eax, 0x0
  3a2a38:      	cmovle	r12, rax
  3a2a3c:      	mov	qword ptr [rbp - 0x40], r14
  3a2a40:      	cmp	r14, r13
  3a2a43:      	jbe	0x3a44bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2a49:      	add	r10, rcx
  3a2a4c:      	mov	qword ptr [rbp - 0x70], r10
  3a2a50:      	mov	qword ptr [rbp - 0xb8], rsi
  3a2a57:      	mov	rax, qword ptr [rsi + 0x40]
  3a2a5b:      	mov	qword ptr [rbp - 0x60], rax
  3a2a5f:      	mov	rsi, qword ptr [rdi + 0xb0]
  3a2a66:      	add	r12, qword ptr [r11 + 0x78]
  3a2a6a:      	mov	r14d, r13d
  3a2a6d:      	mov	qword ptr [rbp - 0x38], r9
  3a2a71:      	mov	qword ptr [rbp - 0x58], r12
  3a2a75:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2a79:      	mov	qword ptr [rbp - 0x48], rsi
  3a2a7d:      	jmp	0x3a2aa6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x116>
  3a2a7f:      	nop
  3a2a80:      	shl	rdi, 0x4
  3a2a84:      	mov	qword ptr [r12 + rdi], rax
  3a2a88:      	mov	qword ptr [r12 + rdi + 0x8], rdx
  3a2a8d:      	inc	r14d
  3a2a90:      	mov	r12d, r14d
  3a2a93:      	mov	r13d, r12d
  3a2a96:      	mov	r14d, r12d
  3a2a99:      	cmp	rcx, r13
  3a2a9c:      	mov	r12, qword ptr [rbp - 0x58]
  3a2aa0:      	jbe	0x3a44c1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b31>
  3a2aa6:      	shl	r13, 0x5
  3a2aaa:      	add	r13, rsi
  3a2aad:      	test	r15b, 0x1
  3a2ab1:      	je	0x3a2b13 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x183>
  3a2ab3:      	mov	rax, qword ptr [rbp - 0x68]
  3a2ab7:      	mov	rsi, qword ptr [rax + 0x80]
  3a2abe:      	mov	edi, dword ptr [r13 + 0x4]
  3a2ac2:      	cmp	rsi, rdi
  3a2ac5:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a2acb:      	mov	r12, qword ptr [rax + 0x78]
  3a2acf:      	imul	r15, rdi, 0xe0
  3a2ad6:      	mov	eax, dword ptr [r12 + r15]
  3a2ada:      	cmp	eax, 0x5
  3a2add:      	setne	cl
  3a2ae0:      	cmp	eax, 0xb
  3a2ae3:      	setne	al
  3a2ae6:      	and	al, cl
  3a2ae8:      	movzx	ecx, al
  3a2aeb:      	mov	rax, qword ptr [r11 + 0x150]
  3a2af2:      	sub	rax, rcx
  3a2af5:      	jb	0x3a42a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1919>
  3a2afb:      	mov	qword ptr [r11 + 0x150], rax
  3a2b02:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a2b07:      	mov	r12, qword ptr [rbp - 0x58]
  3a2b0b:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2b0f:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2b13:      	inc	rbx
  3a2b16:      	test	bl, 0x3f
  3a2b19:      	jne	0x3a2b99 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x209>
  3a2b1b:      	mov	rax, qword ptr [r11 + 0x160]
  3a2b22:      	test	rax, rax
  3a2b25:      	je	0x3a2b3a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aa>
  3a2b27:      	test	r11, r11
  3a2b2a:      	je	0x3a2c3b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2ab>
  3a2b30:      	mov	eax, dword ptr [rax]
  3a2b32:      	test	eax, eax
  3a2b34:      	je	0x3a2fe8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x658>
  3a2b3a:      	mov	rax, qword ptr [r11 + 0x8]
  3a2b3e:      	test	rax, rax
  3a2b41:      	je	0x3a4528 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b98>
  3a2b47:      	mov	rdi, qword ptr [r11 + 0x10]
  3a2b4b:      	mov	rcx, qword ptr [rax + 0x8]
  3a2b4f:      	test	byte ptr [rcx + 0x17], 0x20
  3a2b53:      	je	0x3a2b6c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc>
  3a2b55:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a2b59:      	and	cl, 0x1f
  3a2b5c:      	cmp	cl, 0x16
  3a2b5f:      	je	0x3a2b6c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc>
  3a2b61:      	mov	qword ptr [rbp - 0x98], rdi
  3a2b68:      	mov	rdi, qword ptr [rbp - 0x78]
  3a2b6c:      	call	qword ptr [rax + 0x28]
  3a2b6f:      	mov	r8d, 0xffffffff
  3a2b75:      	mov	r9, qword ptr [rbp - 0x38]
  3a2b79:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2b7d:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2b81:      	test	rax, rax
  3a2b84:      	je	0x3a2c60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d0>
  3a2b8a:      	cmp	qword ptr [rax + 0x8], 0x0
  3a2b8f:      	mov	r11, qword ptr [rbp - 0x50]
  3a2b93:      	jne	0x3a4309 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1979>
  3a2b99:      	cmp	dword ptr [r13 + 0x1c], 0x0
  3a2b9e:      	je	0x3a2c70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2ba4:      	test	r15b, 0x1
  3a2ba8:      	je	0x3a2bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2baa:      	mov	r10d, dword ptr [r13]
  3a2bae:      	cmp	r10d, 0x20
  3a2bb2:      	je	0x3a2c70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2bb8:      	cmp	r10d, 0x41
  3a2bbc:      	je	0x3a2c70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2bc2:      	jmp	0x3a2c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e4>
  3a2bc7:      	nop	word ptr [rax + rax]
  3a2bd0:      	mov	r10d, dword ptr [r13]
  3a2bd4:      	cmp	r10d, 0x20
  3a2bd8:      	je	0x3a2be4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x254>
  3a2bda:      	cmp	r10d, 0x41
  3a2bde:      	jne	0x3a2c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e4>
  3a2be4:      	mov	eax, ebx
  3a2be6:      	not	eax
  3a2be8:      	test	al, 0x3f
  3a2bea:      	je	0x3a2c70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2bf0:      	cmp	r10d, 0x41
  3a2bf4:      	jne	0x3a2c18 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x288>
  3a2bf6:      	mov	edi, dword ptr [r13 + 0x8]
  3a2bfa:      	cmp	r9, rdi
  3a2bfd:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a2c03:      	mov	eax, dword ptr [r13 + 0x18]
  3a2c07:      	shl	rdi, 0x4
  3a2c0b:      	mov	qword ptr [r12 + rdi], rax
  3a2c0f:      	mov	qword ptr [r12 + rdi + 0x8], 0x0
  3a2c18:      	inc	r14d
  3a2c1b:      	cmp	rcx, r14
  3a2c1e:      	jbe	0x3a4372 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19e2>
  3a2c24:      	mov	rax, r14
  3a2c27:      	shl	rax, 0x5
  3a2c2b:      	lea	r13, [rsi + rax]
  3a2c2f:      	inc	rbx
  3a2c32:      	cmp	dword ptr [rsi + rax + 0x1c], 0x0
  3a2c37:      	jne	0x3a2bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2c39:      	jmp	0x3a2c70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c3b:      	mov	edi, 0x1
  3a2c40:      	call	0x5e7350 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a2c45:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2c49:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2c4d:      	mov	r9, qword ptr [rbp - 0x38]
  3a2c51:      	mov	r8d, 0xffffffff
  3a2c57:      	test	rax, rax
  3a2c5a:      	jne	0x3a2b8a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa>
  3a2c60:      	mov	r11, qword ptr [rbp - 0x50]
  3a2c64:      	cmp	dword ptr [r13 + 0x1c], 0x0
  3a2c69:      	jne	0x3a2ba4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x214>
  3a2c6f:      	nop
  3a2c70:      	mov	r10d, dword ptr [r13]
  3a2c74:      	lea	eax, [r10 - 0xe0]
  3a2c7b:      	cmp	eax, 0xffff20
  3a2c80:      	jae	0x3a2f60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d0>
  3a2c86:      	mov	eax, r10d
  3a2c89:      	shr	eax, 0x10
  3a2c8c:      	cmp	eax, 0xfb
  3a2c91:      	je	0x3a2e59 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4c9>
  3a2c97:      	cmp	eax, 0xfc
  3a2c9c:      	je	0x3a2d38 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3a8>
  3a2ca2:      	cmp	eax, 0xfd
  3a2ca7:      	jne	0x3a2edb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x54b>
  3a2cad:      	mov	edi, dword ptr [r13 + 0xc]
  3a2cb1:      	cmp	rdi, r8
  3a2cb4:      	je	0x3a2ffa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x66a>
  3a2cba:      	cmp	r9, rdi
  3a2cbd:      	mov	r15, qword ptr [rbp - 0x68]
  3a2cc1:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a2cc7:      	shl	rdi, 0x4
  3a2ccb:      	mov	rcx, qword ptr [r12 + rdi]
  3a2ccf:      	mov	rdx, qword ptr [r12 + rdi + 0x8]
  3a2cd4:      	mov	edi, dword ptr [r13 + 0x10]
  3a2cd8:      	mov	esi, 0xffffffff
  3a2cdd:      	cmp	rdi, r8
  3a2ce0:      	je	0x3a3014 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x684>
  3a2ce6:      	cmp	r9, rdi
  3a2ce9:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a2cef:      	shl	rdi, 0x4
  3a2cf3:      	mov	rax, qword ptr [r12 + rdi]
  3a2cf7:      	mov	r8, qword ptr [r12 + rdi + 0x8]
  3a2cfc:      	mov	edi, dword ptr [r13 + 0x14]
  3a2d00:      	cmp	rdi, rsi
  3a2d03:      	je	0x3a3026 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x696>
  3a2d09:      	cmp	r9, rdi
  3a2d0c:      	jbe	0x3a4385 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19f5>
  3a2d12:      	shl	rdi, 0x4
  3a2d16:      	mov	r9, qword ptr [r12 + rdi]
  3a2d1a:      	mov	r11, qword ptr [r12 + rdi + 0x8]
  3a2d1f:      	mov	rsi, qword ptr [r15 + 0x80]
  3a2d26:      	mov	edi, dword ptr [r13 + 0x4]
  3a2d2a:      	cmp	rsi, rdi
  3a2d2d:      	ja	0x3a3040 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6b0>
  3a2d33:      	jmp	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a2d38:      	add	r10d, 0xff03ffed
  3a2d3f:      	cmp	r10d, 0x4
  3a2d43:      	jae	0x3a2f29 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x599>
  3a2d49:      	mov	rax, qword ptr [rbp - 0x68]
  3a2d4d:      	mov	rdx, qword ptr [rax + 0xd8]
  3a2d54:      	mov	edi, dword ptr [r13 + 0x18]
  3a2d58:      	mov	eax, dword ptr [r13 + 0x1c]
  3a2d5c:      	lea	ecx, [rdi + rax]
  3a2d5f:      	add	ecx, 0x2
  3a2d62:      	cmp	rdx, rcx
  3a2d65:      	jb	0x3a4437 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aa7>
  3a2d6b:      	cmp	edi, ecx
  3a2d6d:      	ja	0x3a4487 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af7>
  3a2d73:      	sub	rdx, rdi
  3a2d76:      	lea	r8, [4*rdi]
  3a2d7e:      	test	rdx, rdx
  3a2d81:      	mov	eax, 0x0
  3a2d86:      	cmovle	r8, rax
  3a2d8a:      	cmp	ecx, edi
  3a2d8c:      	je	0x3a447e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aee>
  3a2d92:      	mov	rax, qword ptr [rbp - 0x68]
  3a2d96:      	add	r8, qword ptr [rax + 0xc8]
  3a2d9d:      	mov	eax, dword ptr [r8]
  3a2da0:      	cmp	r9, rax
  3a2da3:      	jbe	0x3a4446 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ab6>
  3a2da9:      	sub	rcx, rdi
  3a2dac:      	cmp	ecx, 0x1
  3a2daf:      	je	0x3a446f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1adf>
  3a2db5:      	mov	edi, dword ptr [r8 + 0x4]
  3a2db9:      	cmp	r9, rdi
  3a2dbc:      	jbe	0x3a442e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a9e>
  3a2dc2:      	shl	rax, 0x4
  3a2dc6:      	mov	r9, qword ptr [r12 + rax]
  3a2dca:      	shl	rdi, 0x4
  3a2dce:      	mov	rsi, qword ptr [r12 + rdi]
  3a2dd2:      	mov	r10d, dword ptr [r13]
  3a2dd6:      	lea	eax, [r10 - 0xfc0013]
  3a2ddd:      	cmp	eax, 0x2
  3a2de0:      	jae	0x3a3112 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x782>
  3a2de6:      	cmp	ecx, 0x3
  3a2de9:      	jb	0x3a44e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b57>
  3a2def:      	mov	edi, dword ptr [r8 + 0x8]
  3a2df3:      	cmp	qword ptr [rbp - 0x38], rdi
  3a2df7:      	jbe	0x3a442e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a9e>
  3a2dfd:      	cmp	ecx, 0x3
  3a2e00:      	je	0x3a44d8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b48>
  3a2e06:      	mov	eax, dword ptr [r8 + 0xc]
  3a2e0a:      	cmp	qword ptr [rbp - 0x38], rax
  3a2e0e:      	jbe	0x3a44cc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3c>
  3a2e14:      	shl	rdi, 0x4
  3a2e18:      	mov	rdi, qword ptr [r12 + rdi]
  3a2e1c:      	shl	rax, 0x4
  3a2e20:      	mov	rdx, qword ptr [r12 + rax]
  3a2e24:      	cmp	r10d, 0xfc0013
  3a2e2b:      	jne	0x3a323c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8ac>
  3a2e31:      	lea	rax, [rdi + r9]
  3a2e35:      	mov	r10, rdi
  3a2e38:      	and	r10, r9
  3a2e3b:      	or	rdi, r9
  3a2e3e:      	mov	r9, rax
  3a2e41:      	not	r9
  3a2e44:      	and	r9, rdi
  3a2e47:      	or	r9, r10
  3a2e4a:      	shr	r9, 0x3f
  3a2e4e:      	add	rdx, rsi
  3a2e51:      	add	rdx, r9
  3a2e54:      	jmp	0x3a3261 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8d1>
  3a2e59:      	add	r10d, -0x2
  3a2e5d:      	cmp	r10w, 0x21
  3a2e62:      	jae	0x3a2eed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x55d>
  3a2e68:      	lea	rax, [r8 + 0x1c003e10]
  3a2e6f:      	movzx	ecx, r10w
  3a2e73:      	bt	rax, rcx
  3a2e77:      	jae	0x3a2eed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x55d>
  3a2e79:      	mov	edi, dword ptr [r13 + 0xc]
  3a2e7d:      	mov	eax, 0xffffffff
  3a2e82:      	cmp	rdi, r8
  3a2e85:      	je	0x3a35dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc4c>
  3a2e8b:      	cmp	r9, rdi
  3a2e8e:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a2e94:      	shl	rdi, 0x4
  3a2e98:      	mov	r8, qword ptr [r12 + rdi]
  3a2e9c:      	mov	edi, dword ptr [r13 + 0x10]
  3a2ea0:      	cmp	rdi, rax
  3a2ea3:      	je	0x3a35ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc5c>
  3a2ea9:      	cmp	r9, rdi
  3a2eac:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a2eb2:      	shl	rdi, 0x4
  3a2eb6:      	mov	rax, qword ptr [r12 + rdi]
  3a2eba:      	mov	rdx, qword ptr [r12 + rdi + 0x8]
  3a2ebf:      	mov	r9d, 0x0
  3a2ec5:      	mov	r11d, 0x0
  3a2ecb:      	cmp	r10w, 0xc
  3a2ed0:      	je	0x3a3603 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc73>
  3a2ed6:      	jmp	0x3a363f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcaf>
  3a2edb:      	cmp	r10d, 0xe7
  3a2ee2:      	jae	0x3a2f60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d0>
  3a2ee8:      	jmp	0x3a422d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x189d>
  3a2eed:      	mov	r12, qword ptr [rbp - 0x70]
  3a2ef1:      	mov	dword ptr [r12 - 0x38], r14d
  3a2ef6:      	mov	rdi, r11
  3a2ef9:      	call	0x3a47d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a2efe:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2f02:      	mov	r9, qword ptr [rbp - 0x38]
  3a2f06:      	mov	r8d, 0xffffffff
  3a2f0c:      	mov	r11, qword ptr [rbp - 0x50]
  3a2f10:      	test	al, al
  3a2f12:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2f16:      	jne	0x3a440c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7c>
  3a2f1c:      	inc	r14d
  3a2f1f:      	mov	dword ptr [r12 - 0x38], r14d
  3a2f24:      	jmp	0x3a2a90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a2f29:      	mov	r10d, dword ptr [r13]
  3a2f2d:      	cmp	r10d, 0xfc0008
  3a2f34:      	jae	0x3a2f60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d0>
  3a2f36:      	mov	edi, dword ptr [r13 + 0xc]
  3a2f3a:      	cmp	r9, rdi
  3a2f3d:      	jbe	0x3a4385 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19f5>
  3a2f43:      	shl	rdi, 0x4
  3a2f47:      	mov	rsi, qword ptr [r12 + rdi]
  3a2f4b:      	mov	edi, r10d
  3a2f4e:      	xor	edx, edx
  3a2f50:      	call	0x3d0e20 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a2f55:      	mov	edi, edx
  3a2f57:      	xor	edx, edx
  3a2f59:      	jmp	0x3a37bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe2d>
  3a2f5e:      	nop
  3a2f60:      	mov	edi, dword ptr [r13 + 0xc]
  3a2f64:      	cmp	rdi, r8
  3a2f67:      	je	0x3a2fa0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x610>
  3a2f69:      	cmp	r9, rdi
  3a2f6c:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a2f72:      	shl	rdi, 0x4
  3a2f76:      	mov	rax, qword ptr [r12 + rdi]
  3a2f7a:      	mov	rdx, qword ptr [r12 + rdi + 0x8]
  3a2f7f:      	mov	edi, dword ptr [r13 + 0x10]
  3a2f83:      	cmp	rdi, r8
  3a2f86:      	je	0x3a2fad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x61d>
  3a2f88:      	cmp	r9, rdi
  3a2f8b:      	jbe	0x3a4385 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19f5>
  3a2f91:      	shl	rdi, 0x4
  3a2f95:      	mov	rcx, qword ptr [r12 + rdi]
  3a2f99:      	mov	r9, qword ptr [r12 + rdi + 0x8]
  3a2f9e:      	jmp	0x3a2fb2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x622>
  3a2fa0:      	xor	eax, eax
  3a2fa2:      	xor	edx, edx
  3a2fa4:      	mov	edi, dword ptr [r13 + 0x10]
  3a2fa8:      	cmp	rdi, r8
  3a2fab:      	jne	0x3a2f88 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5f8>
  3a2fad:      	xor	ecx, ecx
  3a2faf:      	xor	r9d, r9d
  3a2fb2:      	mov	esi, r10d
  3a2fb5:      	and	esi, 0xffffff
  3a2fbb:      	mov	edi, r10d
  3a2fbe:      	shr	edi, 0x10
  3a2fc1:      	and	edi, 0xff00
  3a2fc7:      	or	edi, esi
  3a2fc9:      	lea	esi, [rdi - 0x45]
  3a2fcc:      	cmp	esi, 0x39
  3a2fcf:      	ja	0x3a32d9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x949>
  3a2fd5:      	jmp	qword ptr [8*rsi + 0x2033d8]
  3a2fdc:      	xor	ecx, ecx
  3a2fde:      	test	eax, eax
  3a2fe0:      	sete	cl
  3a2fe3:      	jmp	0x3a35bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc2d>
  3a2fe8:      	xor	eax, eax
  3a2fea:      	xor	edx, edx
  3a2fec:      	test	rax, rax
  3a2fef:      	jne	0x3a2b8a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa>
  3a2ff5:      	jmp	0x3a2c60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d0>
  3a2ffa:      	xor	edx, edx
  3a2ffc:      	xor	ecx, ecx
  3a2ffe:      	mov	r15, qword ptr [rbp - 0x68]
  3a3002:      	mov	edi, dword ptr [r13 + 0x10]
  3a3006:      	mov	esi, 0xffffffff
  3a300b:      	cmp	rdi, r8
  3a300e:      	jne	0x3a2ce6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x356>
  3a3014:      	xor	eax, eax
  3a3016:      	xor	r8d, r8d
  3a3019:      	mov	edi, dword ptr [r13 + 0x14]
  3a301d:      	cmp	rdi, rsi
  3a3020:      	jne	0x3a2d09 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x379>
  3a3026:      	xor	r9d, r9d
  3a3029:      	xor	r11d, r11d
  3a302c:      	mov	rsi, qword ptr [r15 + 0x80]
  3a3033:      	mov	edi, dword ptr [r13 + 0x4]
  3a3037:      	cmp	rsi, rdi
  3a303a:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3040:      	imul	rdi, rdi, 0xe0
  3a3047:      	add	rdi, qword ptr [r15 + 0x78]
  3a304b:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3051:      	movzx	esi, r10w
  3a3055:      	cmp	esi, 0xb
  3a3058:      	setb	r15b
  3a305c:      	mov	esi, r10d
  3a305f:      	and	esi, 0xfffc
  3a3065:      	cmp	esi, 0x54
  3a3068:      	sete	r12b
  3a306c:      	or	r12b, r15b
  3a306f:      	jne	0x3a3083 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6f3>
  3a3071:      	movzx	r15d, r10w
  3a3075:      	lea	r12d, [r15 - 0x5c]
  3a3079:      	cmp	r12d, 0x2
  3a307d:      	jae	0x3a3142 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7b2>
  3a3083:      	mov	r9, qword ptr [rbp - 0x60]
  3a3087:      	mov	rsi, qword ptr [r9 + 0x78]
  3a308b:      	mov	edx, dword ptr [rdi + 0x18]
  3a308e:      	cmp	rsi, rdx
  3a3091:      	jbe	0x3a4426 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a96>
  3a3097:      	shl	rdx, 0x3
  3a309b:      	add	rdx, qword ptr [r9 + 0x70]
  3a309f:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a30a4:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a30aa:      	mov	rdx, qword ptr [rdx]
  3a30ad:      	test	rdx, rdx
  3a30b0:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a30b6:      	mov	rsi, qword ptr [rdx + 0x10]
  3a30ba:      	mov	edx, dword ptr [rdi + 0x14]
  3a30bd:      	mov	r11d, dword ptr [rdi + 0x1c]
  3a30c1:      	shl	r11, 0x20
  3a30c5:      	or	r11, rdx
  3a30c8:      	movzx	edx, byte ptr [rdi + 0x40]
  3a30cc:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a30d1:      	je	0x3a30de <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x74e>
  3a30d3:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a30d8:      	je	0x3a31e3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x853>
  3a30de:      	mov	qword ptr [rbp - 0xb0], rax
  3a30e5:      	mov	qword ptr [rbp - 0xa8], r8
  3a30ec:      	movups	xmm0, xmmword ptr [rbp - 0xb0]
  3a30f3:      	movups	xmmword ptr [rsp], xmm0
  3a30f7:      	movzx	r9d, dl
  3a30fb:      	mov	r12, qword ptr [rbp - 0x78]
  3a30ff:      	mov	rdi, r12
  3a3102:      	mov	edx, r10d
  3a3105:      	mov	r8, r11
  3a3108:      	call	0x3c6f80 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a310d:      	jmp	0x3a3212 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x882>
  3a3112:      	mov	rax, rsi
  3a3115:      	mul	r9
  3a3118:      	cmp	dword ptr [r13], 0xfc0015
  3a3120:      	jne	0x3a3261 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8d1>
  3a3126:      	mov	rdi, r9
  3a3129:      	sar	rdi, 0x3f
  3a312d:      	and	rdi, rsi
  3a3130:      	sar	rsi, 0x3f
  3a3134:      	and	rsi, r9
  3a3137:      	add	rsi, rdi
  3a313a:      	sub	rdx, rsi
  3a313d:      	jmp	0x3a3261 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8d1>
  3a3142:      	cmp	r15d, 0xb
  3a3146:      	je	0x3a3151 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7c1>
  3a3148:      	cmp	esi, 0x58
  3a314b:      	jne	0x3a372f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd9f>
  3a3151:      	mov	r9, qword ptr [rbp - 0x60]
  3a3155:      	mov	rsi, qword ptr [r9 + 0x78]
  3a3159:      	mov	edx, dword ptr [rdi + 0x18]
  3a315c:      	cmp	rsi, rdx
  3a315f:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a3164:      	mov	r12, qword ptr [rbp - 0x58]
  3a3168:      	jbe	0x3a4426 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a96>
  3a316e:      	shl	rdx, 0x3
  3a3172:      	add	rdx, qword ptr [r9 + 0x70]
  3a3176:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a317c:      	mov	rdx, qword ptr [rdx]
  3a317f:      	test	rdx, rdx
  3a3182:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3188:      	mov	rdx, qword ptr [rdx + 0x10]
  3a318c:      	mov	esi, dword ptr [rdi + 0x14]
  3a318f:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a3193:      	shl	r9, 0x20
  3a3197:      	or	r9, rsi
  3a319a:      	movzx	esi, byte ptr [rdi + 0x40]
  3a319e:      	cmp	qword ptr [rdx + 0x40], 0x0
  3a31a3:      	je	0x3a31b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x820>
  3a31a5:      	cmp	qword ptr [rdx + 0x60], 0x0
  3a31aa:      	je	0x3a3789 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdf9>
  3a31b0:      	mov	qword ptr [rbp - 0x98], rax
  3a31b7:      	mov	qword ptr [rbp - 0x90], r8
  3a31be:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a31c5:      	movups	xmmword ptr [rsp], xmm0
  3a31c9:      	movzx	r8d, sil
  3a31cd:      	mov	rdi, rdx
  3a31d0:      	mov	esi, r10d
  3a31d3:      	mov	rdx, rcx
  3a31d6:      	mov	rcx, r9
  3a31d9:      	call	0x3c8470 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a31de:      	jmp	0x3a37b7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe27>
  3a31e3:      	mov	qword ptr [rbp - 0xb0], rax
  3a31ea:      	mov	qword ptr [rbp - 0xa8], r8
  3a31f1:      	movups	xmm0, xmmword ptr [rbp - 0xb0]
  3a31f8:      	movups	xmmword ptr [rsp], xmm0
  3a31fc:      	movzx	r9d, dl
  3a3200:      	mov	r12, qword ptr [rbp - 0x78]
  3a3204:      	mov	rdi, r12
  3a3207:      	mov	edx, r10d
  3a320a:      	mov	r8, r11
  3a320d:      	call	0x3c70c0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a3212:      	mov	rax, qword ptr [r12]
  3a3216:      	mov	rdx, qword ptr [r12 + 0x8]
  3a321b:      	mov	qword ptr [rbp - 0x78], r12
  3a321f:      	movzx	edi, byte ptr [r12 + 0x10]
  3a3225:      	mov	r11, qword ptr [rbp - 0x50]
  3a3229:      	mov	r8d, 0xffffffff
  3a322f:      	mov	r9, qword ptr [rbp - 0x38]
  3a3233:      	mov	r12, qword ptr [rbp - 0x58]
  3a3237:      	jmp	0x3a37cb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe3b>
  3a323c:      	mov	rax, r9
  3a323f:      	sub	rax, rdi
  3a3242:      	not	r9
  3a3245:      	mov	r10, rdi
  3a3248:      	and	r10, r9
  3a324b:      	xor	r9, rdi
  3a324e:      	and	r9, rax
  3a3251:      	or	r9, r10
  3a3254:      	shr	r9, 0x3f
  3a3258:      	add	rdx, r9
  3a325b:      	sub	rsi, rdx
  3a325e:      	mov	rdx, rsi
  3a3261:      	mov	r9, qword ptr [rbp - 0x38]
  3a3265:      	mov	edi, dword ptr [r13 + 0x1c]
  3a3269:      	cmp	ecx, edi
  3a326b:      	jbe	0x3a4451 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ac1>
  3a3271:      	lea	esi, [rdi + 0x1]
  3a3274:      	cmp	ecx, esi
  3a3276:      	mov	r10, qword ptr [rbp - 0x70]
  3a327a:      	jbe	0x3a4459 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ac9>
  3a3280:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a3284:      	cmp	r9, rcx
  3a3287:      	jbe	0x3a4464 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ad4>
  3a328d:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a3291:      	shl	rcx, 0x4
  3a3295:      	mov	qword ptr [r12 + rcx], rax
  3a3299:      	mov	qword ptr [r12 + rcx + 0x8], 0x0
  3a32a2:      	cmp	r9, rdi
  3a32a5:      	jbe	0x3a4467 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ad7>
  3a32ab:      	shl	rdi, 0x4
  3a32af:      	mov	qword ptr [r12 + rdi], rdx
  3a32b3:      	mov	qword ptr [r12 + rdi + 0x8], 0x0
  3a32bc:      	inc	r14d
  3a32bf:      	mov	dword ptr [r10 - 0x38], r14d
  3a32c3:      	mov	r12d, r14d
  3a32c6:      	mov	r8d, 0xffffffff
  3a32cc:      	mov	rcx, qword ptr [rbp - 0x40]
  3a32d0:      	mov	rsi, qword ptr [rbp - 0x48]
  3a32d4:      	jmp	0x3a2a93 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x103>
  3a32d9:      	lea	esi, [rdi - 0x5b]
  3a32dc:      	cmp	esi, 0x69
  3a32df:      	ja	0x3a3861 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xed1>
  3a32e5:      	jmp	qword ptr [8*rsi + 0x2035a8]
  3a32ec:      	mov	rsi, rax
  3a32ef:      	mov	rdx, rcx
  3a32f2:      	call	0x3d0e20 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a32f7:      	mov	r8d, 0xffffffff
  3a32fd:      	mov	r11, qword ptr [rbp - 0x50]
  3a3301:      	mov	edi, edx
  3a3303:      	xor	edx, edx
  3a3305:      	jmp	0x3a37c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe37>
  3a330a:      	add	rax, rcx
  3a330d:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3312:      	mov	rcx, rax
  3a3315:      	shr	rcx
  3a3318:      	movabs	rdx, 0x5555555555555555
  3a3322:      	and	rcx, rdx
  3a3325:      	and	rax, rdx
  3a3328:      	add	rax, rcx
  3a332b:      	mov	rcx, rax
  3a332e:      	shr	rcx, 0x2
  3a3332:      	movabs	rdx, 0x3333333333333333
  3a333c:      	and	rcx, rdx
  3a333f:      	and	rax, rdx
  3a3342:      	add	rax, rcx
  3a3345:      	mov	rcx, rax
  3a3348:      	shr	rcx, 0x4
  3a334c:      	add	rcx, rax
  3a334f:      	movabs	rax, 0xf0f0f0f0f0f0f0f
  3a3359:      	and	rax, rcx
  3a335c:      	mov	rcx, rax
  3a335f:      	shr	rcx, 0x8
  3a3363:      	add	rcx, rax
  3a3366:      	mov	rdx, rcx
  3a3369:      	shr	rdx, 0x10
  3a336d:      	add	rdx, rcx
  3a3370:      	mov	rax, rdx
  3a3373:      	shr	rax, 0x20
  3a3377:      	add	eax, edx
  3a3379:      	and	eax, 0x7f
  3a337c:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3381:      	xor	ecx, ecx
  3a3383:      	mov	rdx, rax
  3a3386:      	shr	rdx, 0x20
  3a338a:      	cmove	rdx, rax
  3a338e:      	setne	cl
  3a3391:      	shl	ecx, 0x5
  3a3394:      	mov	rax, rdx
  3a3397:      	shr	rax, 0x10
  3a339b:      	lea	rsi, [rcx + 0x10]
  3a339f:      	cmp	rdx, 0x10000
  3a33a6:      	cmovb	rax, rdx
  3a33aa:      	cmovb	rsi, rcx
  3a33ae:      	mov	ecx, eax
  3a33b0:      	shr	ecx, 0x8
  3a33b3:      	lea	rdx, [rsi + 0x8]
  3a33b7:      	cmp	rax, 0x100
  3a33bd:      	cmovb	rcx, rax
  3a33c1:      	cmovb	rdx, rsi
  3a33c5:      	movzx	eax, cl
  3a33c8:      	movzx	ecx, byte ptr [rax + 0x328461]
  3a33cf:      	add	rcx, rdx
  3a33d2:      	mov	eax, 0x40
  3a33d7:      	sub	rax, rcx
  3a33da:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a33df:      	xor	esi, esi
  3a33e1:      	cmp	rax, rcx
  3a33e4:      	setle	sil
  3a33e8:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a33ed:      	xor	esi, esi
  3a33ef:      	cmp	rax, rcx
  3a33f2:      	setae	sil
  3a33f6:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a33fb:      	add	ecx, eax
  3a33fd:      	jmp	0x3a35bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc2d>
  3a3402:      	test	eax, eax
  3a3404:      	je	0x3a3757 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdc7>
  3a340a:      	tzcnt	eax, eax
  3a340e:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3413:      	shr	eax, cl
  3a3415:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a341a:      	and	eax, ecx
  3a341c:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3421:      	xor	esi, esi
  3a3423:      	cmp	rax, rcx
  3a3426:      	sete	sil
  3a342a:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a342f:      	sub	rax, rcx
  3a3432:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3437:      	xor	ecx, ecx
  3a3439:      	test	rax, rax
  3a343c:      	sete	cl
  3a343f:      	jmp	0x3a35bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc2d>
  3a3444:      	mov	ecx, eax
  3a3446:      	shr	ecx, 0x10
  3a3449:      	xor	edx, edx
  3a344b:      	cmp	eax, 0x10000
  3a3450:      	setb	dl
  3a3453:      	cmovb	ecx, eax
  3a3456:      	shl	edx, 0x4
  3a3459:      	mov	esi, ecx
  3a345b:      	shr	esi, 0x8
  3a345e:      	xor	eax, eax
  3a3460:      	cmp	ecx, 0x100
  3a3466:      	setb	al
  3a3469:      	lea	eax, [rdx + 8*rax]
  3a346c:      	cmovb	esi, ecx
  3a346f:      	movzx	ecx, sil
  3a3473:      	movzx	ecx, byte ptr [rcx + 0x328461]
  3a347a:      	sub	rax, rcx
  3a347d:      	add	rax, 0x8
  3a3481:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3486:      	xor	esi, esi
  3a3488:      	cmp	rax, rcx
  3a348b:      	setbe	sil
  3a348f:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a3494:      	or	ecx, eax
  3a3496:      	jmp	0x3a35bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc2d>
  3a349b:      	shl	eax, cl
  3a349d:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a34a2:      	xor	esi, esi
  3a34a4:      	cmp	eax, ecx
  3a34a6:      	seta	sil
  3a34aa:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a34af:      	xor	esi, esi
  3a34b1:      	cmp	eax, ecx
  3a34b3:      	setle	sil
  3a34b7:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a34bc:      	xor	esi, esi
  3a34be:      	cmp	eax, ecx
  3a34c0:      	setb	sil
  3a34c4:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a34c9:      	mov	ecx, eax
  3a34cb:      	shr	ecx, 0x10
  3a34ce:      	movzx	ecx, cl
  3a34d1:      	movzx	ecx, byte ptr [rcx + 0x328561]
  3a34d8:      	movzx	edx, ah
  3a34db:      	movzx	esi, al
  3a34de:      	shr	eax, 0x18
  3a34e1:      	add	cl, byte ptr [rax + 0x328561]
  3a34e7:      	add	cl, byte ptr [rdx + 0x328561]
  3a34ed:      	add	cl, byte ptr [rsi + 0x328561]
  3a34f3:      	movzx	eax, cl
  3a34f6:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a34fb:      	imul	rax, rcx
  3a34ff:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3504:      	xor	esi, esi
  3a3506:      	cmp	eax, ecx
  3a3508:      	setne	sil
  3a350c:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a3511:      	xor	esi, esi
  3a3513:      	cmp	eax, ecx
  3a3515:      	sete	sil
  3a3519:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a351e:      	xor	esi, esi
  3a3520:      	cmp	eax, ecx
  3a3522:      	setge	sil
  3a3526:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a352b:      	xor	esi, esi
  3a352d:      	cmp	eax, ecx
  3a352f:      	setbe	sil
  3a3533:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a3535:      	xor	esi, esi
  3a3537:      	cmp	rax, rcx
  3a353a:      	setl	sil
  3a353e:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a3540:      	xor	esi, esi
  3a3542:      	cmp	rax, rcx
  3a3545:      	setb	sil
  3a3549:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a354b:      	xor	ecx, eax
  3a354d:      	jmp	0x3a35bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc2d>
  3a354f:      	xor	esi, esi
  3a3551:      	cmp	eax, ecx
  3a3553:      	setg	sil
  3a3557:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a3559:      	mov	ecx, 0x40
  3a355e:      	tzcnt	rcx, rax
  3a3563:      	jmp	0x3a35bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc2d>
  3a3565:      	xor	esi, esi
  3a3567:      	cmp	eax, ecx
  3a3569:      	setl	sil
  3a356d:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a356f:      	xor	esi, esi
  3a3571:      	cmp	rax, rcx
  3a3574:      	seta	sil
  3a3578:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a357a:      	rol	eax, cl
  3a357c:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3581:      	xor	esi, esi
  3a3583:      	cmp	eax, ecx
  3a3585:      	setae	sil
  3a3589:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a358b:      	xor	esi, esi
  3a358d:      	cmp	rax, rcx
  3a3590:      	setg	sil
  3a3594:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a3596:      	xor	esi, esi
  3a3598:      	cmp	rax, rcx
  3a359b:      	setne	sil
  3a359f:      	jmp	0x3a35b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc20>
  3a35a1:      	xor	esi, esi
  3a35a3:      	cmp	rax, rcx
  3a35a6:      	setge	sil
  3a35aa:      	nop	word ptr [rax + rax]
  3a35b0:      	xor	edx, edx
  3a35b2:      	mov	rax, rsi
  3a35b5:      	jmp	0x3a3762 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd2>
  3a35ba:      	imul	ecx, eax
  3a35bd:      	xor	edx, edx
  3a35bf:      	mov	rax, rcx
  3a35c2:      	jmp	0x3a3762 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd2>
  3a35c7:      	sub	eax, ecx
  3a35c9:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a35ce:      	ror	eax, cl
  3a35d0:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a35d5:      	sar	eax, cl
  3a35d7:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a35dc:      	xor	r8d, r8d
  3a35df:      	mov	edi, dword ptr [r13 + 0x10]
  3a35e3:      	cmp	rdi, rax
  3a35e6:      	jne	0x3a2ea9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x519>
  3a35ec:      	xor	eax, eax
  3a35ee:      	xor	edx, edx
  3a35f0:      	mov	r9d, 0x0
  3a35f6:      	mov	r11d, 0x0
  3a35fc:      	cmp	r10w, 0xc
  3a3601:      	jne	0x3a363f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcaf>
  3a3603:      	mov	rcx, qword ptr [rbp - 0x68]
  3a3607:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a360e:      	mov	edi, dword ptr [r13 + 0x18]
  3a3612:      	add	edi, 0x2
  3a3615:      	cmp	rsi, rdi
  3a3618:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a361e:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a3625:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a3628:      	cmp	qword ptr [rbp - 0x38], rdi
  3a362c:      	jbe	0x3a4385 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19f5>
  3a3632:      	shl	rdi, 0x4
  3a3636:      	mov	r9, qword ptr [r12 + rdi]
  3a363a:      	mov	r11, qword ptr [r12 + rdi + 0x8]
  3a363f:      	cmp	qword ptr [rbp - 0x50], 0x0
  3a3644:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a364a:      	mov	r10, qword ptr [rbp - 0x68]
  3a364e:      	mov	rsi, qword ptr [r10 + 0x80]
  3a3655:      	mov	edi, dword ptr [r13 + 0x4]
  3a3659:      	cmp	rsi, rdi
  3a365c:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3662:      	imul	rcx, rdi, 0xe0
  3a3669:      	add	rcx, qword ptr [r10 + 0x78]
  3a366d:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3673:      	mov	rsi, qword ptr [rbp - 0x50]
  3a3677:      	mov	rsi, qword ptr [rsi]
  3a367a:      	mov	qword ptr [rbp - 0xd8], rax
  3a3681:      	mov	qword ptr [rbp - 0xd0], rdx
  3a3688:      	mov	qword ptr [rbp - 0xc8], r9
  3a368f:      	mov	qword ptr [rbp - 0xc0], r11
  3a3696:      	movups	xmm0, xmmword ptr [rbp - 0xc8]
  3a369d:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a36a2:      	movups	xmm0, xmmword ptr [rbp - 0xd8]
  3a36a9:      	movups	xmmword ptr [rsp], xmm0
  3a36ad:      	lea	rdi, [rbp - 0x160]
  3a36b4:      	mov	rdx, qword ptr [rbp - 0x60]
  3a36b8:      	call	0x3b6990 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a36bd:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a36c4:      	mov	rax, qword ptr [rbp - 0x150]
  3a36cb:      	test	rax, rax
  3a36ce:      	je	0x3a36ee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd5e>
  3a36d0:      	cmp	qword ptr [rax + 0x8], 0x0
  3a36d5:      	mov	r8d, 0xffffffff
  3a36db:      	mov	r9, qword ptr [rbp - 0x38]
  3a36df:      	mov	rcx, qword ptr [rbp - 0x40]
  3a36e3:      	mov	rsi, qword ptr [rbp - 0x48]
  3a36e7:      	je	0x3a3700 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd70>
  3a36e9:      	jmp	0x3a4314 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1984>
  3a36ee:      	mov	r8d, 0xffffffff
  3a36f4:      	mov	r9, qword ptr [rbp - 0x38]
  3a36f8:      	mov	rcx, qword ptr [rbp - 0x40]
  3a36fc:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3700:      	mov	edi, dword ptr [r13 + 0x8]
  3a3704:      	cmp	rdi, r8
  3a3707:      	mov	r11, qword ptr [rbp - 0x50]
  3a370b:      	mov	r12, qword ptr [rbp - 0x70]
  3a370f:      	je	0x3a2f1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x58c>
  3a3715:      	cmp	r9, rdi
  3a3718:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a371e:      	shl	rdi, 0x4
  3a3722:      	mov	rax, qword ptr [rbp - 0x58]
  3a3726:      	movups	xmmword ptr [rax + rdi], xmm0
  3a372a:      	jmp	0x3a2f1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x58c>
  3a372f:      	mov	qword ptr [rsp], r11
  3a3733:      	mov	rsi, rcx
  3a3736:      	mov	rcx, rax
  3a3739:      	call	0x3cf020 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a373e:      	mov	r11, qword ptr [rbp - 0x50]
  3a3742:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a3747:      	mov	r8d, 0xffffffff
  3a374d:      	mov	r9, qword ptr [rbp - 0x38]
  3a3751:      	mov	r12, qword ptr [rbp - 0x58]
  3a3755:      	jmp	0x3a3766 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd6>
  3a3757:      	mov	eax, 0x20
  3a375c:      	nop	dword ptr [rax]
  3a3760:      	xor	edx, edx
  3a3762:      	mov	r9, qword ptr [rbp - 0x38]
  3a3766:      	mov	rcx, qword ptr [rbp - 0x40]
  3a376a:      	mov	rsi, qword ptr [rbp - 0x48]
  3a376e:      	mov	edi, dword ptr [r13 + 0x8]
  3a3772:      	cmp	rdi, r8
  3a3775:      	je	0x3a2a8d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a377b:      	cmp	r9, rdi
  3a377e:      	ja	0x3a2a80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0>
  3a3784:      	jmp	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a3789:      	mov	qword ptr [rbp - 0x98], rax
  3a3790:      	mov	qword ptr [rbp - 0x90], r8
  3a3797:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a379e:      	movups	xmmword ptr [rsp], xmm0
  3a37a2:      	movzx	r8d, sil
  3a37a6:      	mov	rdi, rdx
  3a37a9:      	mov	esi, r10d
  3a37ac:      	mov	rdx, rcx
  3a37af:      	mov	rcx, r9
  3a37b2:      	call	0x3c8750 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a37b7:      	mov	edi, eax
  3a37b9:      	xor	edx, edx
  3a37bb:      	xor	eax, eax
  3a37bd:      	mov	r11, qword ptr [rbp - 0x50]
  3a37c1:      	mov	r8d, 0xffffffff
  3a37c7:      	mov	r9, qword ptr [rbp - 0x38]
  3a37cb:      	mov	rcx, qword ptr [rbp - 0x40]
  3a37cf:      	test	dil, dil
  3a37d2:      	mov	rsi, qword ptr [rbp - 0x48]
  3a37d6:      	je	0x3a376e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdde>
  3a37d8:      	jmp	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a06>
  3a37dd:      	or	rax, rcx
  3a37e0:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a37e5:      	mov	eax, eax
  3a37e7:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a37ec:      	shr	rax, cl
  3a37ef:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a37f4:      	and	rax, rcx
  3a37f7:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a37fc:      	rol	rax, cl
  3a37ff:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3804:      	ror	rax, cl
  3a3807:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a380c:      	mov	eax, eax
  3a380e:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3813:      	movsx	rax, ax
  3a3817:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a381c:      	cdqe
  3a381e:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3823:      	cdqe
  3a3825:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a382a:      	sar	rax, cl
  3a382d:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3832:      	movsx	rax, ax
  3a3836:      	mov	eax, eax
  3a3838:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a383d:      	shl	rax, cl
  3a3840:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3845:      	movsx	rax, al
  3a3849:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a384e:      	movsx	rax, al
  3a3852:      	mov	eax, eax
  3a3854:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3859:      	xor	rax, rcx
  3a385c:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3861:      	lea	esi, [rdi - 0x1]
  3a3864:      	cmp	esi, 0x108
  3a386a:      	ja	0x3a3a4d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10bd>
  3a3870:      	jmp	qword ptr [8*rsi + 0x2038f8]
  3a3877:      	mov	rdx, qword ptr [rbp - 0x68]
  3a387b:      	mov	rsi, qword ptr [rdx + 0x80]
  3a3882:      	mov	ecx, dword ptr [r13 + 0x4]
  3a3886:      	cmp	rsi, rcx
  3a3889:      	jbe	0x3a4537 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ba7>
  3a388f:      	imul	r8, rcx, 0xe0
  3a3896:      	add	r8, qword ptr [rdx + 0x78]
  3a389a:      	mov	r10, qword ptr [rbp - 0x78]
  3a389e:      	mov	rdx, qword ptr [rbp - 0x60]
  3a38a2:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a38a8:      	mov	rsi, qword ptr [rdx + 0x78]
  3a38ac:      	mov	ecx, dword ptr [r8 + 0x18]
  3a38b0:      	cmp	rsi, rcx
  3a38b3:      	jbe	0x3a4537 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ba7>
  3a38b9:      	shl	rcx, 0x3
  3a38bd:      	add	rcx, qword ptr [rdx + 0x70]
  3a38c1:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a38c7:      	mov	rcx, qword ptr [rcx]
  3a38ca:      	test	rcx, rcx
  3a38cd:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a38d3:      	mov	rdx, rdi
  3a38d6:      	mov	rsi, qword ptr [rcx + 0x10]
  3a38da:      	mov	ecx, dword ptr [r8 + 0x14]
  3a38de:      	mov	r8d, dword ptr [r8 + 0x1c]
  3a38e2:      	shl	r8, 0x20
  3a38e6:      	or	r8, rcx
  3a38e9:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a38ee:      	je	0x3a38f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf67>
  3a38f0:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a38f5:      	je	0x3a3941 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfb1>
  3a38f7:      	xorps	xmm0, xmm0
  3a38fa:      	movaps	xmmword ptr [rbp - 0xb0], xmm0
  3a3901:      	movups	xmmword ptr [rsp], xmm0
  3a3905:      	mov	rdi, r10
  3a3908:      	mov	rcx, rax
  3a390b:      	xor	r9d, r9d
  3a390e:      	mov	r12, r10
  3a3911:      	call	0x3c6f80 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a3916:      	jmp	0x3a3212 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x882>
  3a391b:      	movzx	ecx, r10b
  3a391f:      	xor	edi, edi
  3a3921:      	add	ecx, -0x2
  3a3924:      	cmp	ecx, 0x7
  3a3927:      	ja	0x3a39c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1037>
  3a392d:      	jmp	qword ptr [8*rcx + 0x204140]
  3a3934:      	xor	edi, edi
  3a3936:      	test	eax, eax
  3a3938:      	sete	dil
  3a393c:      	jmp	0x3a39c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1037>
  3a3941:      	xorps	xmm0, xmm0
  3a3944:      	movaps	xmmword ptr [rbp - 0xb0], xmm0
  3a394b:      	movups	xmmword ptr [rsp], xmm0
  3a394f:      	mov	rdi, r10
  3a3952:      	mov	rcx, rax
  3a3955:      	xor	r9d, r9d
  3a3958:      	mov	r12, r10
  3a395b:      	jmp	0x3a320d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x87d>
  3a3960:      	mov	edi, dword ptr [r13 + 0x18]
  3a3964:      	mov	r9, qword ptr [rbp - 0x38]
  3a3968:      	cmp	r9, rdi
  3a396b:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a3971:      	shl	rdi, 0x4
  3a3975:      	mov	qword ptr [r12 + rdi], rax
  3a3979:      	mov	qword ptr [r12 + rdi + 0x8], rdx
  3a397e:      	jmp	0x3a3766 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd6>
  3a3983:      	mov	edi, dword ptr [r13 + 0x14]
  3a3987:      	cmp	rdi, r8
  3a398a:      	je	0x3a3c85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12f5>
  3a3990:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3994:      	jbe	0x3a4385 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19f5>
  3a399a:      	shl	rdi, 0x4
  3a399e:      	mov	rsi, qword ptr [r12 + rdi]
  3a39a2:      	test	esi, esi
  3a39a4:      	cmove	rax, rcx
  3a39a8:      	cmove	rdx, r9
  3a39ac:      	jmp	0x3a3762 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd2>
  3a39b1:      	mov	edi, dword ptr [r13 + 0x1c]
  3a39b5:      	dec	edi
  3a39b7:      	cmp	edi, eax
  3a39b9:      	cmovae	edi, eax
  3a39bc:      	jmp	0x3a39c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1037>
  3a39be:      	xor	edi, edi
  3a39c0:      	test	rax, rax
  3a39c3:      	setne	dil
  3a39c7:      	mov	rax, qword ptr [rbp - 0x68]
  3a39cb:      	mov	rsi, qword ptr [rax + 0xe8]
  3a39d2:      	add	edi, dword ptr [r13 + 0x18]
  3a39d6:      	cmp	rsi, rdi
  3a39d9:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a39df:      	mov	rax, qword ptr [rax + 0xe0]
  3a39e6:      	lea	rcx, [rdi + 2*rdi]
  3a39ea:      	mov	r12d, dword ptr [rax + 4*rcx]
  3a39ee:      	mov	r13d, dword ptr [rax + 4*rcx + 0x4]
  3a39f3:      	mov	esi, dword ptr [rax + 4*rcx + 0x8]
  3a39f7:      	cmp	r12d, r14d
  3a39fa:      	mov	r9, qword ptr [rbp - 0x38]
  3a39fe:      	ja	0x3a3ceb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x135b>
  3a3a04:      	mov	rax, qword ptr [r11 + 0x160]
  3a3a0b:      	test	rax, rax
  3a3a0e:      	je	0x3a3a23 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1093>
  3a3a10:      	test	r11, r11
  3a3a13:      	je	0x3a3c90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1300>
  3a3a19:      	mov	eax, dword ptr [rax]
  3a3a1b:      	test	eax, eax
  3a3a1d:      	je	0x3a3cc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1332>
  3a3a23:      	mov	rax, qword ptr [r11 + 0x8]
  3a3a27:      	test	rax, rax
  3a3a2a:      	je	0x3a4528 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b98>
  3a3a30:      	mov	rdi, qword ptr [r11 + 0x10]
  3a3a34:      	mov	rcx, qword ptr [rax + 0x8]
  3a3a38:      	test	byte ptr [rcx + 0x17], 0x20
  3a3a3c:      	jne	0x3a3c66 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12d6>
  3a3a42:      	mov	r14, rsi
  3a3a45:      	call	qword ptr [rax + 0x28]
  3a3a48:      	jmp	0x3a3cc9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1339>
  3a3a4d:      	lea	esi, [rdi - 0x10]
  3a3a50:      	cmp	esi, 0x34
  3a3a53:      	ja	0x3a3c9f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x130f>
  3a3a59:      	jmp	qword ptr [8*rsi + 0x204180]
  3a3a60:      	mov	rax, qword ptr [r13 + 0x18]
  3a3a64:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3a69:      	mov	r10, qword ptr [rbp - 0x68]
  3a3a6d:      	mov	rsi, qword ptr [r10 + 0x80]
  3a3a74:      	mov	edx, dword ptr [r13 + 0x4]
  3a3a78:      	cmp	rsi, rdx
  3a3a7b:      	jbe	0x3a4426 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a96>
  3a3a81:      	imul	r8, rdx, 0xe0
  3a3a88:      	add	r8, qword ptr [r10 + 0x78]
  3a3a8c:      	mov	r10, qword ptr [rbp - 0x60]
  3a3a90:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3a96:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3a9a:      	mov	edx, dword ptr [r8 + 0x18]
  3a3a9e:      	cmp	rsi, rdx
  3a3aa1:      	jbe	0x3a4426 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a96>
  3a3aa7:      	shl	rdx, 0x3
  3a3aab:      	add	rdx, qword ptr [r10 + 0x70]
  3a3aaf:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3ab5:      	mov	rdx, qword ptr [rdx]
  3a3ab8:      	test	rdx, rdx
  3a3abb:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3ac1:      	mov	rsi, rdi
  3a3ac4:      	mov	rdi, qword ptr [rdx + 0x10]
  3a3ac8:      	mov	edx, dword ptr [r8 + 0x14]
  3a3acc:      	mov	r8d, dword ptr [r8 + 0x1c]
  3a3ad0:      	shl	r8, 0x20
  3a3ad4:      	or	r8, rdx
  3a3ad7:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a3adc:      	je	0x3a3ae5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1155>
  3a3ade:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a3ae3:      	je	0x3a3b11 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1181>
  3a3ae5:      	mov	qword ptr [rbp - 0x98], rcx
  3a3aec:      	mov	qword ptr [rbp - 0x90], r9
  3a3af3:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a3afa:      	movups	xmmword ptr [rsp], xmm0
  3a3afe:      	mov	rdx, rax
  3a3b01:      	mov	rcx, r8
  3a3b04:      	xor	r8d, r8d
  3a3b07:      	call	0x3c8470 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a3b0c:      	jmp	0x3a37b7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe27>
  3a3b11:      	mov	qword ptr [rbp - 0x98], rcx
  3a3b18:      	mov	qword ptr [rbp - 0x90], r9
  3a3b1f:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a3b26:      	movups	xmmword ptr [rsp], xmm0
  3a3b2a:      	mov	rdx, rax
  3a3b2d:      	mov	rcx, r8
  3a3b30:      	xor	r8d, r8d
  3a3b33:      	jmp	0x3a37b2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe22>
  3a3b38:      	mov	rax, qword ptr [rbp - 0x60]
  3a3b3c:      	mov	rsi, qword ptr [rax + 0x90]
  3a3b43:      	mov	edi, dword ptr [r13 + 0x18]
  3a3b47:      	cmp	rsi, rdi
  3a3b4a:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3b50:      	mov	rax, qword ptr [rax + 0x88]
  3a3b57:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a3b5b:      	mov	rax, qword ptr [rcx + 0x20]
  3a3b5f:      	mov	rdx, qword ptr [rcx + 0x28]
  3a3b63:      	jmp	0x3a3762 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd2>
  3a3b68:      	mov	rcx, qword ptr [rbp - 0x60]
  3a3b6c:      	mov	rsi, qword ptr [rcx + 0x90]
  3a3b73:      	mov	edi, dword ptr [r13 + 0x18]
  3a3b77:      	cmp	rsi, rdi
  3a3b7a:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3b80:      	mov	rcx, qword ptr [rcx + 0x88]
  3a3b87:      	mov	rcx, qword ptr [rcx + 8*rdi]
  3a3b8b:      	mov	qword ptr [rcx + 0x20], rax
  3a3b8f:      	mov	qword ptr [rcx + 0x28], rdx
  3a3b93:      	xor	eax, eax
  3a3b95:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3b9a:      	mov	edi, dword ptr [r13 + 0x18]
  3a3b9e:      	mov	r9, qword ptr [rbp - 0x38]
  3a3ba2:      	cmp	r9, rdi
  3a3ba5:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a3bab:      	shl	rdi, 0x4
  3a3baf:      	mov	rax, qword ptr [r12 + rdi]
  3a3bb3:      	mov	rdx, qword ptr [r12 + rdi + 0x8]
  3a3bb8:      	jmp	0x3a3766 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd6>
  3a3bbd:      	mov	rcx, qword ptr [rbp - 0x60]
  3a3bc1:      	mov	rsi, qword ptr [rcx + 0x60]
  3a3bc5:      	mov	edi, dword ptr [r13 + 0x18]
  3a3bc9:      	cmp	rsi, rdi
  3a3bcc:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3bd2:      	mov	rcx, qword ptr [rcx + 0x58]
  3a3bd6:      	lea	rdx, [rcx + 8*rdi]
  3a3bda:      	test	rdx, rdx
  3a3bdd:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3be3:      	mov	rcx, qword ptr [rcx + 8*rdi]
  3a3be7:      	test	rcx, rcx
  3a3bea:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3bf0:      	cmp	byte ptr [rcx + 0x29], 0x0
  3a3bf4:      	mov	edx, eax
  3a3bf6:      	cmovne	rdx, rax
  3a3bfa:      	cmp	rdx, qword ptr [rcx + 0x40]
  3a3bfe:      	jae	0x3a438e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19fe>
  3a3c04:      	mov	rax, qword ptr [rcx + 0x38]
  3a3c08:      	mov	rax, qword ptr [rax + 8*rdx]
  3a3c0c:      	jmp	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a3c0e:      	mov	rdx, qword ptr [rbp - 0x60]
  3a3c12:      	mov	rsi, qword ptr [rdx + 0x60]
  3a3c16:      	mov	edi, dword ptr [r13 + 0x18]
  3a3c1a:      	cmp	rsi, rdi
  3a3c1d:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3c23:      	mov	rdx, qword ptr [rdx + 0x58]
  3a3c27:      	lea	rsi, [rdx + 8*rdi]
  3a3c2b:      	test	rsi, rsi
  3a3c2e:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3c34:      	mov	rdx, qword ptr [rdx + 8*rdi]
  3a3c38:      	test	rdx, rdx
  3a3c3b:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3c41:      	cmp	byte ptr [rdx + 0x29], 0x0
  3a3c45:      	mov	esi, eax
  3a3c47:      	cmovne	rsi, rax
  3a3c4b:      	cmp	rsi, qword ptr [rdx + 0x40]
  3a3c4f:      	jae	0x3a438e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19fe>
  3a3c55:      	mov	rax, qword ptr [rdx + 0x38]
  3a3c59:      	mov	qword ptr [rax + 8*rsi], rcx
  3a3c5d:      	xor	eax, eax
  3a3c5f:      	xor	edx, edx
  3a3c61:      	jmp	0x3a373e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdae>
  3a3c66:      	mov	r14, rsi
  3a3c69:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a3c6d:      	and	cl, 0x1f
  3a3c70:      	cmp	cl, 0x16
  3a3c73:      	je	0x3a3c80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12f0>
  3a3c75:      	mov	qword ptr [rbp - 0x98], rdi
  3a3c7c:      	mov	rdi, qword ptr [rbp - 0x78]
  3a3c80:      	call	qword ptr [rax + 0x28]
  3a3c83:      	jmp	0x3a3cc9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1339>
  3a3c85:      	mov	rax, rcx
  3a3c88:      	mov	rdx, r9
  3a3c8b:      	jmp	0x3a3762 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd2>
  3a3c90:      	mov	r14, rsi
  3a3c93:      	mov	edi, 0x1
  3a3c98:      	call	0x5e7350 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a3c9d:      	jmp	0x3a3cc9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1339>
  3a3c9f:      	lea	esi, [rdi - 0xd0]
  3a3ca5:      	cmp	esi, 0x4
  3a3ca8:      	ja	0x3a3d7e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13ee>
  3a3cae:      	jmp	qword ptr [8*rsi + 0x204328]
  3a3cb5:      	xor	ecx, ecx
  3a3cb7:      	test	rax, rax
  3a3cba:      	sete	cl
  3a3cbd:      	mov	rax, rcx
  3a3cc0:      	jmp	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a3cc2:      	mov	r14, rsi
  3a3cc5:      	xor	eax, eax
  3a3cc7:      	xor	edx, edx
  3a3cc9:      	test	rax, rax
  3a3ccc:      	mov	r11, qword ptr [rbp - 0x50]
  3a3cd0:      	mov	r8d, 0xffffffff
  3a3cd6:      	mov	r9, qword ptr [rbp - 0x38]
  3a3cda:      	mov	rsi, r14
  3a3cdd:      	je	0x3a3ceb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x135b>
  3a3cdf:      	cmp	qword ptr [rax + 0x8], 0x0
  3a3ce4:      	je	0x3a3ceb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x135b>
  3a3ce6:      	jmp	0x3a4309 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1979>
  3a3ceb:      	mov	rax, qword ptr [rbp - 0x68]
  3a3cef:      	mov	rdx, qword ptr [rax + 0x108]
  3a3cf6:      	lea	eax, [rsi + r13]
  3a3cfa:      	cmp	rdx, rax
  3a3cfd:      	jb	0x3a453f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1baf>
  3a3d03:      	cmp	r13d, eax
  3a3d06:      	ja	0x3a454e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bbe>
  3a3d0c:      	sub	rdx, r13
  3a3d0f:      	lea	rcx, [8*r13]
  3a3d17:      	test	rdx, rdx
  3a3d1a:      	mov	edx, 0x0
  3a3d1f:      	cmovle	rcx, rdx
  3a3d23:      	test	esi, esi
  3a3d25:      	je	0x3a32cc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x93c>
  3a3d2b:      	mov	rdx, qword ptr [rbp - 0x68]
  3a3d2f:      	mov	rdx, qword ptr [rdx + 0xf8]
  3a3d36:      	sub	rax, r13
  3a3d39:      	add	rdx, rcx
  3a3d3c:      	add	rdx, 0x4
  3a3d40:      	xor	esi, esi
  3a3d42:      	mov	edi, dword ptr [rdx + 8*rsi]
  3a3d45:      	cmp	r9, rdi
  3a3d48:      	jbe	0x3a437d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19ed>
  3a3d4e:      	mov	ecx, dword ptr [rdx + 8*rsi - 0x4]
  3a3d52:      	cmp	r9, rcx
  3a3d55:      	jbe	0x3a451d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b8d>
  3a3d5b:      	shl	rdi, 0x4
  3a3d5f:      	shl	rcx, 0x4
  3a3d63:      	mov	r10, qword ptr [rbp - 0x58]
  3a3d67:      	movups	xmm0, xmmword ptr [r10 + rdi]
  3a3d6c:      	movups	xmmword ptr [r10 + rcx], xmm0
  3a3d71:      	inc	rsi
  3a3d74:      	cmp	rax, rsi
  3a3d77:      	jne	0x3a3d42 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13b2>
  3a3d79:      	jmp	0x3a32cc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x93c>
  3a3d7e:      	mov	rdx, rdi
  3a3d81:      	mov	edi, dword ptr [r13 + 0x14]
  3a3d85:      	cmp	rdi, r8
  3a3d88:      	je	0x3a3f08 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1578>
  3a3d8e:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3d92:      	jbe	0x3a4385 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19f5>
  3a3d98:      	shl	rdi, 0x4
  3a3d9c:      	mov	rsi, qword ptr [rbp - 0x58]
  3a3da0:      	movups	xmm0, xmmword ptr [rsi + rdi]
  3a3da4:      	jmp	0x3a3f0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157b>
  3a3da9:      	mov	eax, dword ptr [r13 + 0x18]
  3a3dad:      	jmp	0x3a3760 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd0>
  3a3db2:      	mov	rax, qword ptr [rbp - 0x60]
  3a3db6:      	mov	rsi, qword ptr [rax + 0x78]
  3a3dba:      	mov	edi, dword ptr [r13 + 0x18]
  3a3dbe:      	cmp	rsi, rdi
  3a3dc1:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3dc7:      	shl	rdi, 0x3
  3a3dcb:      	mov	rax, qword ptr [rbp - 0x60]
  3a3dcf:      	add	rdi, qword ptr [rax + 0x70]
  3a3dd3:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3dd9:      	mov	rax, qword ptr [rdi]
  3a3ddc:      	test	rax, rax
  3a3ddf:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3de5:      	mov	rdi, qword ptr [rax + 0x10]
  3a3de9:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a3dee:      	mov	rax, qword ptr [rdi + 0x60]
  3a3df2:      	je	0x3a40c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1735>
  3a3df8:      	test	rax, rax
  3a3dfb:      	jne	0x3a40ce <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x173e>
  3a3e01:      	call	0x3c4b80 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a3e06:      	jmp	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a3e0b:      	mov	rcx, qword ptr [rbp - 0x60]
  3a3e0f:      	mov	rsi, qword ptr [rcx + 0x78]
  3a3e13:      	mov	edi, dword ptr [r13 + 0x18]
  3a3e17:      	cmp	rsi, rdi
  3a3e1a:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3e20:      	shl	rdi, 0x3
  3a3e24:      	mov	rcx, qword ptr [rbp - 0x60]
  3a3e28:      	add	rdi, qword ptr [rcx + 0x70]
  3a3e2c:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3e32:      	cmp	qword ptr [rbp - 0x50], 0x0
  3a3e37:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3e3d:      	mov	r15, qword ptr [rdi]
  3a3e40:      	test	r15, r15
  3a3e43:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3e49:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3e4d:      	mov	rdi, qword ptr [rcx]
  3a3e50:      	mov	rcx, qword ptr [r15 + 0x10]
  3a3e54:      	movzx	ecx, byte ptr [rcx + 0x31]
  3a3e58:      	mov	byte ptr [rbp - 0x79], cl
  3a3e5b:      	mov	edx, eax
  3a3e5d:      	test	cl, 0x1
  3a3e60:      	cmovne	rdx, rax
  3a3e64:      	mov	rsi, r15
  3a3e67:      	call	0x3c10f0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a3e6c:      	test	dl, 0x1
  3a3e6f:      	jne	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a3e75:      	mov	rax, qword ptr [r15 + 0x10]
  3a3e79:      	movzx	eax, byte ptr [rax + 0x31]
  3a3e7d:      	mov	byte ptr [rbp - 0x7a], al
  3a3e80:      	and	al, 0x1
  3a3e82:      	neg	al
  3a3e84:      	mov	eax, 0x0
  3a3e89:      	sbb	rax, rax
  3a3e8c:      	mov	r8d, 0xffffffff
  3a3e92:      	or	rax, r8
  3a3e95:      	xor	edx, edx
  3a3e97:      	mov	r11, qword ptr [rbp - 0x50]
  3a3e9b:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a3ea0:      	jmp	0x3a374d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdbd>
  3a3ea5:      	test	rax, rax
  3a3ea8:      	mov	r11, qword ptr [rbp - 0x50]
  3a3eac:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a3eb1:      	mov	r8d, 0xffffffff
  3a3eb7:      	mov	r9, qword ptr [rbp - 0x38]
  3a3ebb:      	mov	r12, qword ptr [rbp - 0x58]
  3a3ebf:      	mov	rcx, qword ptr [rbp - 0x40]
  3a3ec3:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3ec7:      	jne	0x3a376e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdde>
  3a3ecd:      	jmp	0x3a44f6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b66>
  3a3ed2:      	mov	rax, qword ptr [rbp - 0x60]
  3a3ed6:      	mov	rsi, qword ptr [rax + 0x48]
  3a3eda:      	mov	edi, dword ptr [r13 + 0x18]
  3a3ede:      	cmp	rsi, rdi
  3a3ee1:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3ee7:      	mov	rax, qword ptr [rbp - 0x60]
  3a3eeb:      	mov	rax, qword ptr [rax + 0x40]
  3a3eef:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3ef3:      	jmp	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a3ef8:      	xor	edx, edx
  3a3efa:      	cmp	rax, rcx
  3a3efd:      	sete	dl
  3a3f00:      	mov	rax, rdx
  3a3f03:      	jmp	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a3f08:      	xorps	xmm0, xmm0
  3a3f0b:      	and	r10d, 0xff0000
  3a3f12:      	cmp	r10d, 0xfe0000
  3a3f19:      	jne	0x3a3f52 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15c2>
  3a3f1b:      	movzx	r9d, dx
  3a3f1f:      	cmp	r9d, 0x3
  3a3f23:      	jne	0x3a4104 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1774>
  3a3f29:      	cmp	qword ptr [rbp - 0x50], 0x0
  3a3f2e:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3f34:      	mov	rax, qword ptr [rbp - 0x50]
  3a3f38:      	mov	rax, qword ptr [rax]
  3a3f3b:      	test	rax, rax
  3a3f3e:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3f44:      	mov	rdi, qword ptr [rax + 0x8]
  3a3f48:      	call	0x3c4120 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a3f4d:      	jmp	0x3a3c5d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cd>
  3a3f52:      	add	edx, 0xff03fff8
  3a3f58:      	cmp	edx, 0xa
  3a3f5b:      	jae	0x3a4560 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd0>
  3a3f61:      	mov	rdx, qword ptr [rbp - 0x68]
  3a3f65:      	mov	rsi, qword ptr [rdx + 0x80]
  3a3f6c:      	mov	edi, dword ptr [r13 + 0x4]
  3a3f70:      	cmp	rsi, rdi
  3a3f73:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3f79:      	imul	rdx, rdi, 0xe0
  3a3f80:      	mov	rsi, qword ptr [rbp - 0x68]
  3a3f84:      	add	rdx, qword ptr [rsi + 0x78]
  3a3f88:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a3f8e:      	movups	xmm1, xmmword ptr [rdx + 0x98]
  3a3f95:      	movups	xmm2, xmmword ptr [rdx + 0x88]
  3a3f9c:      	movups	xmm3, xmmword ptr [rdx + 0x78]
  3a3fa0:      	movups	xmm4, xmmword ptr [rdx + 0x68]
  3a3fa4:      	movups	xmm5, xmmword ptr [rdx + 0x58]
  3a3fa8:      	movups	xmm6, xmmword ptr [rdx + 0x48]
  3a3fac:      	movzx	esi, byte ptr [rdx + 0x40]
  3a3fb0:      	movups	xmm7, xmmword ptr [rdx + 0x30]
  3a3fb4:      	movups	xmm8, xmmword ptr [rdx + 0x20]
  3a3fb9:      	movups	xmm9, xmmword ptr [rdx + 0x10]
  3a3fbe:      	mov	rdi, qword ptr [rdx + 0x8]
  3a3fc2:      	mov	edx, dword ptr [rdx]
  3a3fc4:      	mov	dword ptr [rbp - 0x218], edx
  3a3fca:      	mov	qword ptr [rbp - 0x210], rdi
  3a3fd1:      	movups	xmmword ptr [rbp - 0x208], xmm9
  3a3fd9:      	movups	xmmword ptr [rbp - 0x1f8], xmm8
  3a3fe1:      	movups	xmmword ptr [rbp - 0x1e8], xmm7
  3a3fe8:      	mov	byte ptr [rbp - 0x1d8], sil
  3a3fef:      	movups	xmmword ptr [rbp - 0x1d0], xmm6
  3a3ff6:      	movups	xmmword ptr [rbp - 0x1c0], xmm5
  3a3ffd:      	movups	xmmword ptr [rbp - 0x1b0], xmm4
  3a4004:      	movups	xmmword ptr [rbp - 0x1a0], xmm3
  3a400b:      	movups	xmmword ptr [rbp - 0x190], xmm2
  3a4012:      	movups	xmmword ptr [rbp - 0x180], xmm1
  3a4019:      	mov	qword ptr [rbp - 0xe8], rcx
  3a4020:      	mov	qword ptr [rbp - 0xe0], r9
  3a4027:      	movaps	xmmword ptr [rbp - 0x170], xmm0
  3a402e:      	mov	ecx, 0x15
  3a4033:      	mov	rdi, rsp
  3a4036:      	lea	rsi, [rbp - 0x218]
  3a403d:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a4040:      	movaps	xmm0, xmmword ptr [rbp - 0x170]
  3a4047:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a404f:      	movups	xmm0, xmmword ptr [rbp - 0xe8]
  3a4056:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a405e:      	lea	rdi, [rbp - 0x110]
  3a4065:      	mov	rsi, qword ptr [rbp - 0x50]
  3a4069:      	mov	rcx, rax
  3a406c:      	mov	rdx, qword ptr [rbp - 0x60]
  3a4070:      	call	0x3a9510 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a4075:      	mov	rax, qword ptr [rbp - 0x110]
  3a407c:      	mov	rdx, qword ptr [rbp - 0x108]
  3a4083:      	movzx	edi, byte ptr [rbp - 0x100]
  3a408a:      	mov	rsi, qword ptr [rbp - 0xf8]
  3a4091:      	test	rsi, rsi
  3a4094:      	je	0x3a41f8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1868>
  3a409a:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a409f:      	mov	r11, qword ptr [rbp - 0x50]
  3a40a3:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a40a8:      	mov	r8d, 0xffffffff
  3a40ae:      	mov	r9, qword ptr [rbp - 0x38]
  3a40b2:      	mov	r12, qword ptr [rbp - 0x58]
  3a40b6:      	mov	rcx, qword ptr [rbp - 0x40]
  3a40ba:      	je	0x3a37cf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe3f>
  3a40c0:      	jmp	0x3a4505 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b75>
  3a40c5:      	test	rax, rax
  3a40c8:      	je	0x3a4203 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1873>
  3a40ce:      	mov	rax, qword ptr [rax]
  3a40d1:      	test	rdi, rdi
  3a40d4:      	je	0x3a4210 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1880>
  3a40da:      	mov	rcx, qword ptr [rdi + 0x38]
  3a40de:      	test	rcx, rcx
  3a40e1:      	mov	edx, 0x10000
  3a40e6:      	cmove	rcx, rdx
  3a40ea:      	mov	rdx, rax
  3a40ed:      	or	rdx, rcx
  3a40f0:      	shr	rdx, 0x20
  3a40f4:      	je	0x3a41fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x186a>
  3a40fa:      	xor	edx, edx
  3a40fc:      	div	rcx
  3a40ff:      	jmp	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a4104:      	mov	rdx, qword ptr [rbp - 0x68]
  3a4108:      	mov	rsi, qword ptr [rdx + 0x80]
  3a410f:      	mov	edi, dword ptr [r13 + 0x4]
  3a4113:      	cmp	rsi, rdi
  3a4116:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a411c:      	imul	r11, rdi, 0xe0
  3a4123:      	mov	rdx, qword ptr [rbp - 0x68]
  3a4127:      	add	r11, qword ptr [rdx + 0x78]
  3a412b:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a4131:      	mov	rdx, qword ptr [rbp - 0x60]
  3a4135:      	mov	rsi, qword ptr [rdx + 0x78]
  3a4139:      	mov	edi, dword ptr [r11 + 0x18]
  3a413d:      	cmp	rsi, rdi
  3a4140:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a4146:      	shl	rdi, 0x3
  3a414a:      	mov	rdx, qword ptr [rbp - 0x60]
  3a414e:      	add	rdi, qword ptr [rdx + 0x70]
  3a4152:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a4158:      	mov	rsi, qword ptr [rdi]
  3a415b:      	test	rsi, rsi
  3a415e:      	je	0x3a42fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196a>
  3a4164:      	mov	r10, qword ptr [rbp - 0x50]
  3a4168:      	mov	rdi, qword ptr [r10]
  3a416b:      	mov	rdx, qword ptr [r10 + 0x8]
  3a416f:      	mov	r8, qword ptr [rdi + 0x8]
  3a4173:      	mov	r10, qword ptr [r10 + 0x10]
  3a4177:      	mov	rsi, qword ptr [rsi + 0x10]
  3a417b:      	mov	edi, dword ptr [r11 + 0x14]
  3a417f:      	mov	r11d, dword ptr [r11 + 0x1c]
  3a4183:      	shl	r11, 0x20
  3a4187:      	or	r11, rdi
  3a418a:      	movlps	qword ptr [rsp + 0x18], xmm0
  3a418f:      	mov	qword ptr [rsp + 0x10], rcx
  3a4194:      	mov	qword ptr [rsp + 0x8], r11
  3a4199:      	mov	qword ptr [rsp], rax
  3a419d:      	lea	rdi, [rbp - 0x138]
  3a41a4:      	mov	rcx, r10
  3a41a7:      	call	0x3c5430 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a41ac:      	mov	rax, qword ptr [rbp - 0x138]
  3a41b3:      	mov	rdx, qword ptr [rbp - 0x130]
  3a41ba:      	movzx	edi, byte ptr [rbp - 0x128]
  3a41c1:      	mov	rsi, qword ptr [rbp - 0x120]
  3a41c8:      	test	rsi, rsi
  3a41cb:      	je	0x3a421f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x188f>
  3a41cd:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a41d2:      	mov	r11, qword ptr [rbp - 0x50]
  3a41d6:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a41db:      	mov	r8d, 0xffffffff
  3a41e1:      	mov	r9, qword ptr [rbp - 0x38]
  3a41e5:      	mov	r12, qword ptr [rbp - 0x58]
  3a41e9:      	mov	rcx, qword ptr [rbp - 0x40]
  3a41ed:      	je	0x3a37cf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe3f>
  3a41f3:      	jmp	0x3a450e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b7e>
  3a41f8:      	jmp	0x3a421f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x188f>
  3a41fa:      	xor	edx, edx
  3a41fc:      	div	ecx
  3a41fe:      	jmp	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a4203:      	mov	rax, qword ptr [rdi + 0x8]
  3a4207:      	test	rdi, rdi
  3a420a:      	jne	0x3a40da <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x174a>
  3a4210:      	mov	edi, 0x1
  3a4215:      	call	0x5e7350 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a421a:      	jmp	0x3a3c5f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12cf>
  3a421f:      	mov	r11, qword ptr [rbp - 0x50]
  3a4223:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a4228:      	jmp	0x3a3229 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x899>
  3a422d:      	mov	rax, qword ptr [rbp - 0x70]
  3a4231:      	mov	dword ptr [rax - 0x38], r14d
  3a4235:      	mov	rdi, r11
  3a4238:      	mov	r14, r11
  3a423b:      	call	0x3a6950 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a4240:      	test	al, al
  3a4242:      	jne	0x3a440c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7c>
  3a4248:      	mov	rcx, qword ptr [r14 + 0x48]
  3a424c:      	test	rcx, rcx
  3a424f:      	mov	r11, r14
  3a4252:      	mov	r8d, 0xffffffff
  3a4258:      	jg	0x3a29d5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x45>
  3a425e:      	jmp	0x3a456f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bdf>
  3a4263:      	mov	rax, qword ptr [rbp - 0x70]
  3a4267:      	mov	dword ptr [rax - 0x38], r14d
  3a426b:      	mov	rdi, r11
  3a426e:      	mov	r14, r11
  3a4271:      	call	0x3a6450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a4276:      	jmp	0x3a4240 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18b0>
  3a4278:      	mov	rax, qword ptr [rbp - 0x70]
  3a427c:      	mov	dword ptr [rax - 0x38], r14d
  3a4280:      	mov	rdi, r11
  3a4283:      	mov	r14, r11
  3a4286:      	call	0x3a5260 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a428b:      	test	al, al
  3a428d:      	je	0x3a4248 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18b8>
  3a428f:      	jmp	0x3a44fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b6e>
  3a4294:      	mov	rax, qword ptr [rbp - 0x70]
  3a4298:      	mov	dword ptr [rax - 0x38], r14d
  3a429c:      	mov	rdi, r11
  3a429f:      	mov	r14, r11
  3a42a2:      	call	0x3a56f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a42a7:      	jmp	0x3a4240 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18b0>
  3a42a9:      	mov	edi, 0x28
  3a42ae:      	call	0x5fd920 <GC_malloc@plt>
  3a42b3:      	test	rax, rax
  3a42b6:      	je	0x3a457e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bee>
  3a42bc:      	mov	r14, rax
  3a42bf:      	add	r12, r15
  3a42c2:      	mov	edi, 0x28
  3a42c7:      	call	0x5e2fe0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a42cc:      	mov	rax, qword ptr [r12 + 0x8]
  3a42d1:      	movaps	xmm0, xmmword ptr [rip + 0x317038] # 0x6bb310 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a42d8:      	movups	xmmword ptr [r14], xmm0
  3a42dc:      	mov	qword ptr [r14 + 0x10], rax
  3a42e0:      	mov	qword ptr [r14 + 0x18], 0x29a008
  3a42e8:      	mov	qword ptr [r14 + 0x20], 0xe
  3a42f0:      	mov	esi, 0x2baa08
  3a42f5:      	jmp	0x3a43fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a6e>
  3a42fa:      	mov	edi, 0x1
  3a42ff:      	call	0x5e7350 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a4304:      	jmp	0x3a440c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7c>
  3a4309:      	mov	rdi, r11
  3a430c:      	mov	rsi, rax
  3a430f:      	jmp	0x3a4405 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a75>
  3a4314:      	mov	r15, rax
  3a4317:      	mov	r14, qword ptr [rbp - 0x148]
  3a431e:      	mov	edi, 0x8
  3a4323:      	call	0x5fd920 <GC_malloc@plt>
  3a4328:      	test	rax, rax
  3a432b:      	je	0x3a457e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bee>
  3a4331:      	mov	r12, rax
  3a4334:      	mov	edi, 0x8
  3a4339:      	call	0x5e2fe0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a433e:      	mov	edx, 0x299f78
  3a4343:      	mov	rdi, r15
  3a4346:      	mov	rsi, r14
  3a4349:      	mov	rcx, r12
  3a434c:      	call	0x5b9ef0 <errors.As>
  3a4351:      	test	al, 0x1
  3a4353:      	je	0x3a4366 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19d6>
  3a4355:      	mov	rax, qword ptr [r12]
  3a4359:      	mov	rcx, qword ptr [rbp - 0xb8]
  3a4360:      	mov	ecx, dword ptr [rcx + 0x50]
  3a4363:      	mov	dword ptr [rax + 0x4], ecx
  3a4366:      	mov	rdi, qword ptr [rbp - 0x50]
  3a436a:      	mov	rsi, r15
  3a436d:      	jmp	0x3a4402 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a72>
  3a4372:      	mov	rdi, r14
  3a4375:      	mov	rsi, rcx
  3a4378:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a437d:      	mov	rsi, r9
  3a4380:      	call	0x5e30c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4385:      	mov	rsi, qword ptr [rbp - 0x38]
  3a4389:      	call	0x5e30c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a438e:      	mov	dil, 0x3
  3a4391:      	jmp	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a06>
  3a4393:      	mov	dil, 0x1
  3a4396:      	mov	r15d, edi
  3a4399:      	mov	edi, 0x10
  3a439e:      	call	0x5fd9d0 <GC_malloc_atomic@plt>
  3a43a3:      	test	rax, rax
  3a43a6:      	je	0x3a457e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bee>
  3a43ac:      	mov	r14, rax
  3a43af:      	mov	edi, 0x10
  3a43b4:      	call	0x5e2fe0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a43b9:      	xorps	xmm0, xmm0
  3a43bc:      	movups	xmmword ptr [r14], xmm0
  3a43c0:      	mov	rax, qword ptr [rbp - 0x68]
  3a43c4:      	mov	rsi, qword ptr [rax + 0x80]
  3a43cb:      	mov	edi, dword ptr [r13 + 0x4]
  3a43cf:      	cmp	rsi, rdi
  3a43d2:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a43d4:      	mov	rax, qword ptr [rax + 0x78]
  3a43d8:      	mov	rcx, qword ptr [rbp - 0xb8]
  3a43df:      	mov	ecx, dword ptr [rcx + 0x50]
  3a43e2:      	imul	rdx, rdi, 0xe0
  3a43e9:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a43ee:      	mov	byte ptr [r14], r15b
  3a43f1:      	mov	dword ptr [r14 + 0x4], ecx
  3a43f5:      	mov	qword ptr [r14 + 0x8], rax
  3a43f9:      	mov	esi, 0x29a390
  3a43fe:      	mov	rdi, qword ptr [rbp - 0x50]
  3a4402:      	mov	rdx, r14
  3a4405:      	call	0x3a2360 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a440a:      	mov	al, 0x3
  3a440c:      	mov	rdx, rbx
  3a440f:      	add	rsp, 0x2b8
  3a4416:      	pop	rbx
  3a4417:      	pop	r12
  3a4419:      	pop	r13
  3a441b:      	pop	r14
  3a441d:      	pop	r15
  3a441f:      	pop	rbp
  3a4420:      	ret
  3a4421:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4426:      	mov	rdi, rdx
  3a4429:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a442e:      	mov	rsi, qword ptr [rbp - 0x38]
  3a4432:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4437:      	mov	rdi, rcx
  3a443a:      	xor	esi, esi
  3a443c:      	mov	ecx, 0x2
  3a4441:      	call	0x5e3e00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4446:      	mov	rdi, rax
  3a4449:      	mov	rsi, r9
  3a444c:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4451:      	mov	rsi, rcx
  3a4454:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4459:      	mov	rdi, rsi
  3a445c:      	mov	rsi, rcx
  3a445f:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4464:      	mov	rdi, rcx
  3a4467:      	mov	rsi, r9
  3a446a:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a446f:      	mov	edi, 0x1
  3a4474:      	mov	esi, 0x1
  3a4479:      	call	0x5e30c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a447e:      	xor	edi, edi
  3a4480:      	xor	esi, esi
  3a4482:      	call	0x5e30c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4487:      	xor	esi, esi
  3a4489:      	mov	rdx, rcx
  3a448c:      	mov	ecx, 0x3
  3a4491:      	call	0x5e3e00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4496:      	mov	rdi, rax
  3a4499:      	mov	esi, 0x1
  3a449e:      	mov	ecx, 0x2
  3a44a3:      	call	0x5e3e00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a44a8:      	mov	rdi, r12
  3a44ab:      	mov	esi, 0x1
  3a44b0:      	mov	rdx, rax
  3a44b3:      	mov	ecx, 0x3
  3a44b8:      	call	0x5e3e00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a44bd:      	mov	rcx, qword ptr [rbp - 0x40]
  3a44c1:      	mov	rdi, r13
  3a44c4:      	mov	rsi, rcx
  3a44c7:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a44cc:      	mov	rdi, rax
  3a44cf:      	mov	rsi, qword ptr [rbp - 0x38]
  3a44d3:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a44d8:      	mov	edi, 0x3
  3a44dd:      	mov	esi, 0x3
  3a44e2:      	call	0x5e30c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a44e7:      	mov	edi, 0x2
  3a44ec:      	mov	esi, 0x2
  3a44f1:      	call	0x5e30c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a44f6:      	mov	dil, 0x9
  3a44f9:      	jmp	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a06>
  3a44fe:      	mov	al, 0x1
  3a4500:      	jmp	0x3a440c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7c>
  3a4505:      	mov	rdx, qword ptr [rbp - 0xf0]
  3a450c:      	jmp	0x3a4515 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b85>
  3a450e:      	mov	rdx, qword ptr [rbp - 0x118]
  3a4515:      	mov	rdi, r11
  3a4518:      	jmp	0x3a4405 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a75>
  3a451d:      	mov	rdi, rcx
  3a4520:      	mov	rsi, r9
  3a4523:      	call	0x5e30c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4528:      	mov	edi, 0x329540
  3a452d:      	mov	esi, 0x329870
  3a4532:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a4537:      	mov	rdi, rcx
  3a453a:      	call	0x5e45f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a453f:      	mov	rdi, rax
  3a4542:      	xor	esi, esi
  3a4544:      	mov	ecx, 0x2
  3a4549:      	call	0x5e3e00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a454e:      	mov	rdi, r13
  3a4551:      	xor	esi, esi
  3a4553:      	mov	rdx, rax
  3a4556:      	mov	ecx, 0x3
  3a455b:      	call	0x5e3e00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4560:      	mov	edi, 0x3339f8
  3a4565:      	mov	esi, 0x29a018
  3a456a:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a456f:      	mov	edi, 0x3339f8
  3a4574:      	mov	esi, 0x299ff8
  3a4579:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a457e:      	mov	edi, 0x3339f8
  3a4583:      	mov	esi, 0x329518
  3a4588:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a458d:      	int3
  3a458e:      	int3
  3a458f:      	int3
