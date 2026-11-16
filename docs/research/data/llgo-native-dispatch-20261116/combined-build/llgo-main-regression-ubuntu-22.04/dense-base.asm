
/home/runner/work/wasmvm/wasmvm/main-regression/dense-base:	file format elf64-x86-64

Disassembly of section .text:

00000000003a2b80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a2b80:      	push	rbp
  3a2b81:      	mov	rbp, rsp
  3a2b84:      	push	r15
  3a2b86:      	push	r14
  3a2b88:      	push	r13
  3a2b8a:      	push	r12
  3a2b8c:      	push	rbx
  3a2b8d:      	sub	rsp, 0x308
  3a2b94:      	mov	rcx, qword ptr [rdi + 0x48]
  3a2b98:      	test	rcx, rcx
  3a2b9b:      	jle	0x3a45f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a75>
  3a2ba1:      	mov	rbx, rsi
  3a2ba4:      	mov	r11, rdi
  3a2ba7:      	movzx	r15d, byte ptr [rdi + 0x20]
  3a2bac:      	mov	r10d, 0xffffffff
  3a2bb2:      	lea	rax, [rbp - 0xb0]
  3a2bb9:      	mov	qword ptr [rbp - 0x80], rax
  3a2bbd:      	mov	qword ptr [rbp - 0x48], rdi
  3a2bc1:      	mov	byte ptr [rbp - 0x29], r15b
  3a2bc5:      	mov	r8, qword ptr [r11 + 0x40]
  3a2bc9:      	shl	rcx, 0x6
  3a2bcd:      	mov	byte ptr [r8 + rcx - 0x30], 0x0
  3a2bd3:      	mov	rsi, qword ptr [r8 + rcx - 0x40]
  3a2bd8:      	mov	rax, qword ptr [rsi + 0x48]
  3a2bdc:      	mov	qword ptr [rbp - 0x60], rax
  3a2be0:      	mov	r9, qword ptr [rax + 0x98]
  3a2be7:      	mov	r12, qword ptr [r8 + rcx - 0x28]
  3a2bec:      	mov	rdx, qword ptr [r11 + 0x88]
  3a2bf3:      	lea	rax, [r9 + r12]
  3a2bf7:      	cmp	rax, rdx
  3a2bfa:      	ja	0x3a4530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b0>
  3a2c00:      	cmp	r12, rax
  3a2c03:      	ja	0x3a4542 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c2>
  3a2c09:      	mov	rdi, qword ptr [rbp - 0x60]
  3a2c0d:      	mov	r14, qword ptr [rdi + 0xb8]
  3a2c14:      	mov	r13d, dword ptr [r8 + rcx - 0x38]
  3a2c19:      	sub	rdx, r12
  3a2c1c:      	shl	r12, 0x4
  3a2c20:      	test	rdx, rdx
  3a2c23:      	mov	eax, 0x0
  3a2c28:      	cmovle	r12, rax
  3a2c2c:      	mov	qword ptr [rbp - 0x68], r14
  3a2c30:      	cmp	r14, r13
  3a2c33:      	jbe	0x3a4557 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19d7>
  3a2c39:      	add	r8, rcx
  3a2c3c:      	mov	qword ptr [rbp - 0x70], r8
  3a2c40:      	mov	qword ptr [rbp - 0x88], rsi
  3a2c47:      	mov	rax, qword ptr [rsi + 0x40]
  3a2c4b:      	mov	qword ptr [rbp - 0x58], rax
  3a2c4f:      	mov	rcx, qword ptr [rdi + 0xb0]
  3a2c56:      	add	r12, qword ptr [r11 + 0x78]
  3a2c5a:      	mov	r14d, r13d
  3a2c5d:      	mov	qword ptr [rbp - 0x38], r9
  3a2c61:      	mov	qword ptr [rbp - 0x50], r12
  3a2c65:      	mov	r8, qword ptr [rbp - 0x68]
  3a2c69:      	mov	qword ptr [rbp - 0x40], rcx
  3a2c6d:      	jmp	0x3a2c9a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11a>
  3a2c6f:      	nop
  3a2c70:      	shl	rdi, 0x4
  3a2c74:      	mov	qword ptr [r12 + rdi], rax
  3a2c78:      	mov	qword ptr [r12 + rdi + 0x8], r8
  3a2c7d:      	inc	r14d
  3a2c80:      	mov	r12d, r14d
  3a2c83:      	mov	r8, qword ptr [rbp - 0x68]
  3a2c87:      	mov	r13d, r12d
  3a2c8a:      	mov	r14d, r12d
  3a2c8d:      	cmp	r8, r13
  3a2c90:      	mov	r12, qword ptr [rbp - 0x50]
  3a2c94:      	jbe	0x3a455b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19db>
  3a2c9a:      	shl	r13, 0x5
  3a2c9e:      	add	r13, rcx
  3a2ca1:      	test	r15b, 0x1
  3a2ca5:      	je	0x3a2d03 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x183>
  3a2ca7:      	mov	rax, qword ptr [rbp - 0x60]
  3a2cab:      	mov	rsi, qword ptr [rax + 0x80]
  3a2cb2:      	mov	edi, dword ptr [r13 + 0x4]
  3a2cb6:      	cmp	rsi, rdi
  3a2cb9:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a2cbf:      	mov	r12, qword ptr [rax + 0x78]
  3a2cc3:      	imul	r15, rdi, 0xe0
  3a2cca:      	mov	eax, dword ptr [r12 + r15]
  3a2cce:      	cmp	eax, 0x5
  3a2cd1:      	setne	cl
  3a2cd4:      	cmp	eax, 0xb
  3a2cd7:      	setne	al
  3a2cda:      	and	al, cl
  3a2cdc:      	movzx	ecx, al
  3a2cdf:      	mov	rax, qword ptr [r11 + 0x150]
  3a2ce6:      	sub	rax, rcx
  3a2ce9:      	jb	0x3a4345 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c5>
  3a2cef:      	mov	qword ptr [r11 + 0x150], rax
  3a2cf6:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a2cfb:      	mov	r12, qword ptr [rbp - 0x50]
  3a2cff:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2d03:      	inc	rbx
  3a2d06:      	test	bl, 0x3f
  3a2d09:      	jne	0x3a2d89 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x209>
  3a2d0b:      	mov	rax, qword ptr [r11 + 0x160]
  3a2d12:      	test	rax, rax
  3a2d15:      	je	0x3a2d2a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aa>
  3a2d17:      	test	r11, r11
  3a2d1a:      	je	0x3a2e28 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2a8>
  3a2d20:      	mov	eax, dword ptr [rax]
  3a2d22:      	test	eax, eax
  3a2d24:      	je	0x3a31c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x647>
  3a2d2a:      	mov	rax, qword ptr [r11 + 0x8]
  3a2d2e:      	test	rax, rax
  3a2d31:      	je	0x3a45c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a47>
  3a2d37:      	mov	rdi, qword ptr [r11 + 0x10]
  3a2d3b:      	mov	rcx, qword ptr [rax + 0x8]
  3a2d3f:      	test	byte ptr [rcx + 0x17], 0x20
  3a2d43:      	je	0x3a2d5c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc>
  3a2d45:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a2d49:      	and	cl, 0x1f
  3a2d4c:      	cmp	cl, 0x16
  3a2d4f:      	je	0x3a2d5c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc>
  3a2d51:      	mov	qword ptr [rbp - 0xb0], rdi
  3a2d58:      	mov	rdi, qword ptr [rbp - 0x80]
  3a2d5c:      	call	qword ptr [rax + 0x28]
  3a2d5f:      	mov	r10d, 0xffffffff
  3a2d65:      	mov	r9, qword ptr [rbp - 0x38]
  3a2d69:      	mov	r8, qword ptr [rbp - 0x68]
  3a2d6d:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2d71:      	test	rax, rax
  3a2d74:      	je	0x3a2e4d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2cd>
  3a2d7a:      	cmp	qword ptr [rax + 0x8], 0x0
  3a2d7f:      	mov	r11, qword ptr [rbp - 0x48]
  3a2d83:      	jne	0x3a43a5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1825>
  3a2d89:      	cmp	dword ptr [r13 + 0x1c], 0x0
  3a2d8e:      	je	0x3a2e60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2d94:      	test	r15b, 0x1
  3a2d98:      	je	0x3a2dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2d9a:      	mov	edx, dword ptr [r13]
  3a2d9e:      	cmp	edx, 0x20
  3a2da1:      	je	0x3a2e60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2da7:      	cmp	edx, 0x41
  3a2daa:      	je	0x3a2e60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2db0:      	jmp	0x3a2e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e4>
  3a2db5:      	nop	word ptr cs:[rax + rax]
  3a2dc0:      	mov	edx, dword ptr [r13]
  3a2dc4:      	cmp	edx, 0x20
  3a2dc7:      	je	0x3a2dd2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x252>
  3a2dc9:      	cmp	edx, 0x41
  3a2dcc:      	jne	0x3a2e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e4>
  3a2dd2:      	mov	eax, ebx
  3a2dd4:      	not	eax
  3a2dd6:      	test	al, 0x3f
  3a2dd8:      	je	0x3a2e60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2dde:      	cmp	edx, 0x41
  3a2de1:      	jne	0x3a2e05 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x285>
  3a2de3:      	mov	edi, dword ptr [r13 + 0x8]
  3a2de7:      	cmp	r9, rdi
  3a2dea:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a2df0:      	mov	eax, dword ptr [r13 + 0x18]
  3a2df4:      	shl	rdi, 0x4
  3a2df8:      	mov	qword ptr [r12 + rdi], rax
  3a2dfc:      	mov	qword ptr [r12 + rdi + 0x8], 0x0
  3a2e05:      	inc	r14d
  3a2e08:      	cmp	r8, r14
  3a2e0b:      	jbe	0x3a440e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x188e>
  3a2e11:      	mov	rax, r14
  3a2e14:      	shl	rax, 0x5
  3a2e18:      	lea	r13, [rcx + rax]
  3a2e1c:      	inc	rbx
  3a2e1f:      	cmp	dword ptr [rcx + rax + 0x1c], 0x0
  3a2e24:      	jne	0x3a2dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2e26:      	jmp	0x3a2e60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2e28:      	mov	edi, 0x1
  3a2e2d:      	call	0x5e61e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a2e32:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2e36:      	mov	r8, qword ptr [rbp - 0x68]
  3a2e3a:      	mov	r9, qword ptr [rbp - 0x38]
  3a2e3e:      	mov	r10d, 0xffffffff
  3a2e44:      	test	rax, rax
  3a2e47:      	jne	0x3a2d7a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa>
  3a2e4d:      	mov	r11, qword ptr [rbp - 0x48]
  3a2e51:      	cmp	dword ptr [r13 + 0x1c], 0x0
  3a2e56:      	jne	0x3a2d94 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x214>
  3a2e5c:      	nop	dword ptr [rax]
  3a2e60:      	mov	edx, dword ptr [r13]
  3a2e64:      	lea	eax, [rdx - 0xe0]
  3a2e6a:      	cmp	eax, 0xffff20
  3a2e6f:      	jae	0x3a3140 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5c0>
  3a2e75:      	mov	eax, edx
  3a2e77:      	shr	eax, 0x10
  3a2e7a:      	cmp	eax, 0xfb
  3a2e7f:      	je	0x3a3045 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4c5>
  3a2e85:      	cmp	eax, 0xfc
  3a2e8a:      	je	0x3a2f26 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3a6>
  3a2e90:      	cmp	eax, 0xfd
  3a2e95:      	jne	0x3a30ba <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x53a>
  3a2e9b:      	mov	edi, dword ptr [r13 + 0xc]
  3a2e9f:      	mov	esi, 0xffffffff
  3a2ea4:      	cmp	rdi, r10
  3a2ea7:      	je	0x3a31d9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x659>
  3a2ead:      	cmp	r9, rdi
  3a2eb0:      	mov	r15, qword ptr [rbp - 0x60]
  3a2eb4:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a2eba:      	shl	rdi, 0x4
  3a2ebe:      	mov	rcx, qword ptr [r12 + rdi]
  3a2ec2:      	mov	r10, qword ptr [r12 + rdi + 0x8]
  3a2ec7:      	mov	edi, dword ptr [r13 + 0x10]
  3a2ecb:      	cmp	rdi, rsi
  3a2ece:      	je	0x3a31ef <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x66f>
  3a2ed4:      	cmp	r9, rdi
  3a2ed7:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a2edd:      	shl	rdi, 0x4
  3a2ee1:      	mov	rax, qword ptr [r12 + rdi]
  3a2ee5:      	mov	r8, qword ptr [r12 + rdi + 0x8]
  3a2eea:      	mov	edi, dword ptr [r13 + 0x14]
  3a2eee:      	cmp	rdi, rsi
  3a2ef1:      	je	0x3a3201 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x681>
  3a2ef7:      	cmp	r9, rdi
  3a2efa:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18a1>
  3a2f00:      	shl	rdi, 0x4
  3a2f04:      	mov	r9, qword ptr [r12 + rdi]
  3a2f08:      	mov	r11, qword ptr [r12 + rdi + 0x8]
  3a2f0d:      	mov	rsi, qword ptr [r15 + 0x80]
  3a2f14:      	mov	edi, dword ptr [r13 + 0x4]
  3a2f18:      	cmp	rsi, rdi
  3a2f1b:      	ja	0x3a321b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x69b>
  3a2f21:      	jmp	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a2f26:      	add	edx, 0xff03ffed
  3a2f2c:      	cmp	edx, 0x4
  3a2f2f:      	jae	0x3a310a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x58a>
  3a2f35:      	mov	rax, qword ptr [rbp - 0x60]
  3a2f39:      	mov	rdx, qword ptr [rax + 0xd8]
  3a2f40:      	mov	edi, dword ptr [r13 + 0x18]
  3a2f44:      	mov	eax, dword ptr [r13 + 0x1c]
  3a2f48:      	lea	ecx, [rdi + rax]
  3a2f4b:      	add	ecx, 0x2
  3a2f4e:      	cmp	rdx, rcx
  3a2f51:      	jb	0x3a44d1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1951>
  3a2f57:      	cmp	edi, ecx
  3a2f59:      	ja	0x3a4521 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19a1>
  3a2f5f:      	sub	rdx, rdi
  3a2f62:      	lea	r8, [4*rdi]
  3a2f6a:      	test	rdx, rdx
  3a2f6d:      	mov	eax, 0x0
  3a2f72:      	cmovle	r8, rax
  3a2f76:      	cmp	ecx, edi
  3a2f78:      	je	0x3a4518 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1998>
  3a2f7e:      	mov	rax, qword ptr [rbp - 0x60]
  3a2f82:      	add	r8, qword ptr [rax + 0xc8]
  3a2f89:      	mov	eax, dword ptr [r8]
  3a2f8c:      	cmp	r9, rax
  3a2f8f:      	jbe	0x3a44e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1960>
  3a2f95:      	sub	rcx, rdi
  3a2f98:      	cmp	ecx, 0x1
  3a2f9b:      	je	0x3a4509 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1989>
  3a2fa1:      	mov	edi, dword ptr [r8 + 0x4]
  3a2fa5:      	cmp	r9, rdi
  3a2fa8:      	jbe	0x3a44c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1948>
  3a2fae:      	shl	rax, 0x4
  3a2fb2:      	mov	r9, qword ptr [r12 + rax]
  3a2fb6:      	shl	rdi, 0x4
  3a2fba:      	mov	rsi, qword ptr [r12 + rdi]
  3a2fbe:      	mov	r10d, dword ptr [r13]
  3a2fc2:      	lea	eax, [r10 - 0xfc0013]
  3a2fc9:      	cmp	eax, 0x2
  3a2fcc:      	jae	0x3a32ed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x76d>
  3a2fd2:      	cmp	ecx, 0x3
  3a2fd5:      	jb	0x3a4581 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a01>
  3a2fdb:      	mov	edi, dword ptr [r8 + 0x8]
  3a2fdf:      	cmp	qword ptr [rbp - 0x38], rdi
  3a2fe3:      	jbe	0x3a44c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1948>
  3a2fe9:      	cmp	ecx, 0x3
  3a2fec:      	je	0x3a4572 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19f2>
  3a2ff2:      	mov	eax, dword ptr [r8 + 0xc]
  3a2ff6:      	cmp	qword ptr [rbp - 0x38], rax
  3a2ffa:      	jbe	0x3a4566 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19e6>
  3a3000:      	shl	rdi, 0x4
  3a3004:      	mov	rdi, qword ptr [r12 + rdi]
  3a3008:      	shl	rax, 0x4
  3a300c:      	mov	rdx, qword ptr [r12 + rax]
  3a3010:      	cmp	r10d, 0xfc0013
  3a3017:      	jne	0x3a3414 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x894>
  3a301d:      	lea	rax, [rdi + r9]
  3a3021:      	mov	r10, rdi
  3a3024:      	and	r10, r9
  3a3027:      	or	rdi, r9
  3a302a:      	mov	r9, rax
  3a302d:      	not	r9
  3a3030:      	and	r9, rdi
  3a3033:      	or	r9, r10
  3a3036:      	shr	r9, 0x3f
  3a303a:      	add	rdx, rsi
  3a303d:      	add	rdx, r9
  3a3040:      	jmp	0x3a3439 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8b9>
  3a3045:      	add	edx, -0x2
  3a3048:      	cmp	dx, 0x21
  3a304c:      	jae	0x3a30cb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x54b>
  3a304e:      	lea	rax, [r10 + 0x1c003e10]
  3a3055:      	movzx	ecx, dx
  3a3058:      	bt	rax, rcx
  3a305c:      	jae	0x3a30cb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x54b>
  3a305e:      	mov	edi, dword ptr [r13 + 0xc]
  3a3062:      	cmp	rdi, r10
  3a3065:      	je	0x3a379e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc1e>
  3a306b:      	cmp	r9, rdi
  3a306e:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a3074:      	shl	rdi, 0x4
  3a3078:      	mov	r8, qword ptr [r12 + rdi]
  3a307c:      	mov	edi, dword ptr [r13 + 0x10]
  3a3080:      	cmp	rdi, r10
  3a3083:      	je	0x3a37ae <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc2e>
  3a3089:      	cmp	r9, rdi
  3a308c:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18a1>
  3a3092:      	shl	rdi, 0x4
  3a3096:      	mov	rax, qword ptr [r12 + rdi]
  3a309a:      	mov	r9, qword ptr [r12 + rdi + 0x8]
  3a309f:      	mov	r10d, 0x0
  3a30a5:      	mov	r11d, 0x0
  3a30ab:      	cmp	dx, 0xc
  3a30af:      	je	0x3a37c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc45>
  3a30b5:      	jmp	0x3a3801 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc81>
  3a30ba:      	cmp	edx, 0xe7
  3a30c0:      	jae	0x3a3140 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5c0>
  3a30c6:      	jmp	0x3a42c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1749>
  3a30cb:      	mov	r12, qword ptr [rbp - 0x70]
  3a30cf:      	mov	dword ptr [r12 - 0x38], r14d
  3a30d4:      	mov	rdi, r11
  3a30d7:      	call	0x3a4860 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a30dc:      	mov	r8, qword ptr [rbp - 0x68]
  3a30e0:      	mov	r9, qword ptr [rbp - 0x38]
  3a30e4:      	mov	r10d, 0xffffffff
  3a30ea:      	mov	r11, qword ptr [rbp - 0x48]
  3a30ee:      	test	al, al
  3a30f0:      	mov	rcx, qword ptr [rbp - 0x40]
  3a30f4:      	jne	0x3a44a6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1926>
  3a30fa:      	inc	r14d
  3a30fd:      	mov	dword ptr [r12 - 0x38], r14d
  3a3102:      	mov	r12d, r14d
  3a3105:      	jmp	0x3a2c87 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x107>
  3a310a:      	mov	edx, dword ptr [r13]
  3a310e:      	cmp	edx, 0xfc0008
  3a3114:      	jae	0x3a3140 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5c0>
  3a3116:      	mov	edi, dword ptr [r13 + 0xc]
  3a311a:      	cmp	r9, rdi
  3a311d:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18a1>
  3a3123:      	shl	rdi, 0x4
  3a3127:      	mov	rsi, qword ptr [r12 + rdi]
  3a312b:      	mov	edi, edx
  3a312d:      	xor	edx, edx
  3a312f:      	call	0x3d0a10 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a3134:      	xor	r8d, r8d
  3a3137:      	jmp	0x3a398a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe0a>
  3a313c:      	nop	dword ptr [rax]
  3a3140:      	mov	edi, dword ptr [r13 + 0xc]
  3a3144:      	cmp	rdi, r10
  3a3147:      	je	0x3a3180 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x600>
  3a3149:      	cmp	r9, rdi
  3a314c:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a3152:      	shl	rdi, 0x4
  3a3156:      	mov	rax, qword ptr [r12 + rdi]
  3a315a:      	mov	r8, qword ptr [r12 + rdi + 0x8]
  3a315f:      	mov	edi, dword ptr [r13 + 0x10]
  3a3163:      	cmp	rdi, r10
  3a3166:      	je	0x3a318e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x60e>
  3a3168:      	cmp	r9, rdi
  3a316b:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18a1>
  3a3171:      	shl	rdi, 0x4
  3a3175:      	mov	rcx, qword ptr [r12 + rdi]
  3a3179:      	mov	r9, qword ptr [r12 + rdi + 0x8]
  3a317e:      	jmp	0x3a3193 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x613>
  3a3180:      	xor	eax, eax
  3a3182:      	xor	r8d, r8d
  3a3185:      	mov	edi, dword ptr [r13 + 0x10]
  3a3189:      	cmp	rdi, r10
  3a318c:      	jne	0x3a3168 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5e8>
  3a318e:      	xor	ecx, ecx
  3a3190:      	xor	r9d, r9d
  3a3193:      	mov	esi, edx
  3a3195:      	and	esi, 0xffffff
  3a319b:      	mov	edi, edx
  3a319d:      	shr	edi, 0x10
  3a31a0:      	and	edi, 0xff00
  3a31a6:      	or	edi, esi
  3a31a8:      	lea	esi, [rdi - 0x45]
  3a31ab:      	cmp	esi, 0x39
  3a31ae:      	ja	0x3a34b6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x936>
  3a31b4:      	jmp	qword ptr [8*rsi + 0x2033c8]
  3a31bb:      	xor	ecx, ecx
  3a31bd:      	test	eax, eax
  3a31bf:      	sete	cl
  3a31c2:      	jmp	0x3a377e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbfe>
  3a31c7:      	xor	eax, eax
  3a31c9:      	xor	edx, edx
  3a31cb:      	test	rax, rax
  3a31ce:      	jne	0x3a2d7a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa>
  3a31d4:      	jmp	0x3a2e4d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2cd>
  3a31d9:      	xor	r10d, r10d
  3a31dc:      	xor	ecx, ecx
  3a31de:      	mov	r15, qword ptr [rbp - 0x60]
  3a31e2:      	mov	edi, dword ptr [r13 + 0x10]
  3a31e6:      	cmp	rdi, rsi
  3a31e9:      	jne	0x3a2ed4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x354>
  3a31ef:      	xor	eax, eax
  3a31f1:      	xor	r8d, r8d
  3a31f4:      	mov	edi, dword ptr [r13 + 0x14]
  3a31f8:      	cmp	rdi, rsi
  3a31fb:      	jne	0x3a2ef7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x377>
  3a3201:      	xor	r9d, r9d
  3a3204:      	xor	r11d, r11d
  3a3207:      	mov	rsi, qword ptr [r15 + 0x80]
  3a320e:      	mov	edi, dword ptr [r13 + 0x4]
  3a3212:      	cmp	rsi, rdi
  3a3215:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a321b:      	imul	rdi, rdi, 0xe0
  3a3222:      	add	rdi, qword ptr [r15 + 0x78]
  3a3226:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a322c:      	movzx	esi, dx
  3a322f:      	cmp	esi, 0xb
  3a3232:      	setb	r15b
  3a3236:      	mov	esi, edx
  3a3238:      	and	esi, 0xfffc
  3a323e:      	cmp	esi, 0x54
  3a3241:      	sete	r12b
  3a3245:      	or	r12b, r15b
  3a3248:      	jne	0x3a325c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6dc>
  3a324a:      	movzx	r15d, dx
  3a324e:      	lea	r12d, [r15 - 0x5c]
  3a3252:      	cmp	r12d, 0x2
  3a3256:      	jae	0x3a331d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x79d>
  3a325c:      	mov	r10, qword ptr [rbp - 0x58]
  3a3260:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3264:      	mov	r9d, dword ptr [rdi + 0x18]
  3a3268:      	cmp	rsi, r9
  3a326b:      	jbe	0x3a44c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1940>
  3a3271:      	shl	r9, 0x3
  3a3275:      	add	r9, qword ptr [r10 + 0x70]
  3a3279:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a327e:      	mov	r11, qword ptr [rbp - 0x80]
  3a3282:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3288:      	mov	rsi, qword ptr [r9]
  3a328b:      	test	rsi, rsi
  3a328e:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3294:      	mov	rsi, qword ptr [rsi + 0x10]
  3a3298:      	mov	r9d, dword ptr [rdi + 0x14]
  3a329c:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a32a0:      	shl	r10, 0x20
  3a32a4:      	or	r10, r9
  3a32a7:      	movzx	edi, byte ptr [rdi + 0x40]
  3a32ab:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a32b0:      	je	0x3a32bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x73d>
  3a32b2:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a32b7:      	je	0x3a33bf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x83f>
  3a32bd:      	mov	qword ptr [rbp - 0x98], rax
  3a32c4:      	mov	qword ptr [rbp - 0x90], r8
  3a32cb:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a32d2:      	movups	xmmword ptr [rsp], xmm0
  3a32d6:      	movzx	r9d, dil
  3a32da:      	mov	rdi, r11
  3a32dd:      	mov	r8, r10
  3a32e0:      	mov	r12, r11
  3a32e3:      	call	0x3c6b30 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a32e8:      	jmp	0x3a33ea <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x86a>
  3a32ed:      	mov	rax, rsi
  3a32f0:      	mul	r9
  3a32f3:      	cmp	dword ptr [r13], 0xfc0015
  3a32fb:      	jne	0x3a3439 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8b9>
  3a3301:      	mov	rdi, r9
  3a3304:      	sar	rdi, 0x3f
  3a3308:      	and	rdi, rsi
  3a330b:      	sar	rsi, 0x3f
  3a330f:      	and	rsi, r9
  3a3312:      	add	rsi, rdi
  3a3315:      	sub	rdx, rsi
  3a3318:      	jmp	0x3a3439 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8b9>
  3a331d:      	cmp	r15d, 0xb
  3a3321:      	je	0x3a332c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7ac>
  3a3323:      	cmp	esi, 0x58
  3a3326:      	jne	0x3a38f1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd71>
  3a332c:      	mov	r10, qword ptr [rbp - 0x58]
  3a3330:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3334:      	mov	r9d, dword ptr [rdi + 0x18]
  3a3338:      	cmp	rsi, r9
  3a333b:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a3340:      	mov	r12, qword ptr [rbp - 0x50]
  3a3344:      	jbe	0x3a44c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1940>
  3a334a:      	shl	r9, 0x3
  3a334e:      	add	r9, qword ptr [r10 + 0x70]
  3a3352:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3358:      	mov	rsi, qword ptr [r9]
  3a335b:      	test	rsi, rsi
  3a335e:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3364:      	mov	rsi, qword ptr [rsi + 0x10]
  3a3368:      	mov	r10d, dword ptr [rdi + 0x14]
  3a336c:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a3370:      	shl	r9, 0x20
  3a3374:      	or	r9, r10
  3a3377:      	movzx	edi, byte ptr [rdi + 0x40]
  3a337b:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a3380:      	je	0x3a338d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x80d>
  3a3382:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a3387:      	je	0x3a3956 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd6>
  3a338d:      	mov	qword ptr [rbp - 0xb0], rax
  3a3394:      	mov	qword ptr [rbp - 0xa8], r8
  3a339b:      	movups	xmm0, xmmword ptr [rbp - 0xb0]
  3a33a2:      	movups	xmmword ptr [rsp], xmm0
  3a33a6:      	movzx	r8d, dil
  3a33aa:      	mov	rdi, rsi
  3a33ad:      	mov	esi, edx
  3a33af:      	mov	rdx, rcx
  3a33b2:      	mov	rcx, r9
  3a33b5:      	call	0x3c8020 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a33ba:      	jmp	0x3a3983 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe03>
  3a33bf:      	mov	qword ptr [rbp - 0x98], rax
  3a33c6:      	mov	qword ptr [rbp - 0x90], r8
  3a33cd:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a33d4:      	movups	xmmword ptr [rsp], xmm0
  3a33d8:      	movzx	r9d, dil
  3a33dc:      	mov	rdi, r11
  3a33df:      	mov	r8, r10
  3a33e2:      	mov	r12, r11
  3a33e5:      	call	0x3c6c70 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a33ea:      	mov	rax, qword ptr [r12]
  3a33ee:      	mov	r8, qword ptr [r12 + 0x8]
  3a33f3:      	mov	qword ptr [rbp - 0x80], r12
  3a33f7:      	movzx	edx, byte ptr [r12 + 0x10]
  3a33fd:      	mov	r11, qword ptr [rbp - 0x48]
  3a3401:      	mov	r10d, 0xffffffff
  3a3407:      	mov	r9, qword ptr [rbp - 0x38]
  3a340b:      	mov	r12, qword ptr [rbp - 0x50]
  3a340f:      	jmp	0x3a3998 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe18>
  3a3414:      	mov	rax, r9
  3a3417:      	sub	rax, rdi
  3a341a:      	not	r9
  3a341d:      	mov	r10, rdi
  3a3420:      	and	r10, r9
  3a3423:      	xor	r9, rdi
  3a3426:      	and	r9, rax
  3a3429:      	or	r9, r10
  3a342c:      	shr	r9, 0x3f
  3a3430:      	add	rdx, r9
  3a3433:      	sub	rsi, rdx
  3a3436:      	mov	rdx, rsi
  3a3439:      	mov	r10d, 0xffffffff
  3a343f:      	mov	edi, dword ptr [r13 + 0x1c]
  3a3443:      	cmp	ecx, edi
  3a3445:      	mov	r9, qword ptr [rbp - 0x38]
  3a3449:      	jbe	0x3a44eb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196b>
  3a344f:      	lea	esi, [rdi + 0x1]
  3a3452:      	cmp	ecx, esi
  3a3454:      	mov	r12, qword ptr [rbp - 0x70]
  3a3458:      	jbe	0x3a44f3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1973>
  3a345e:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a3462:      	cmp	r9, rcx
  3a3465:      	jbe	0x3a44fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x197e>
  3a346b:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a346f:      	shl	rcx, 0x4
  3a3473:      	mov	rsi, qword ptr [rbp - 0x50]
  3a3477:      	mov	qword ptr [rsi + rcx], rax
  3a347b:      	mov	qword ptr [rsi + rcx + 0x8], 0x0
  3a3484:      	cmp	r9, rdi
  3a3487:      	jbe	0x3a4501 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1981>
  3a348d:      	shl	rdi, 0x4
  3a3491:      	mov	qword ptr [rsi + rdi], rdx
  3a3495:      	mov	qword ptr [rsi + rdi + 0x8], 0x0
  3a349e:      	inc	r14d
  3a34a1:      	mov	dword ptr [r12 - 0x38], r14d
  3a34a6:      	mov	r12d, r14d
  3a34a9:      	mov	r8, qword ptr [rbp - 0x68]
  3a34ad:      	mov	rcx, qword ptr [rbp - 0x40]
  3a34b1:      	jmp	0x3a2c87 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x107>
  3a34b6:      	lea	esi, [rdi - 0x5b]
  3a34b9:      	cmp	esi, 0x69
  3a34bc:      	ja	0x3a3a23 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xea3>
  3a34c2:      	jmp	qword ptr [8*rsi + 0x203598]
  3a34c9:      	mov	rsi, rax
  3a34cc:      	mov	rdx, rcx
  3a34cf:      	call	0x3d0a10 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a34d4:      	mov	r10d, 0xffffffff
  3a34da:      	mov	r11, qword ptr [rbp - 0x48]
  3a34de:      	xor	r8d, r8d
  3a34e1:      	jmp	0x3a3994 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe14>
  3a34e6:      	add	rax, rcx
  3a34e9:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a34ee:      	mov	rcx, rax
  3a34f1:      	shr	rcx
  3a34f4:      	movabs	rdx, 0x5555555555555555
  3a34fe:      	and	rcx, rdx
  3a3501:      	and	rax, rdx
  3a3504:      	add	rax, rcx
  3a3507:      	mov	rcx, rax
  3a350a:      	shr	rcx, 0x2
  3a350e:      	movabs	rdx, 0x3333333333333333
  3a3518:      	and	rcx, rdx
  3a351b:      	and	rax, rdx
  3a351e:      	add	rax, rcx
  3a3521:      	mov	rcx, rax
  3a3524:      	shr	rcx, 0x4
  3a3528:      	add	rcx, rax
  3a352b:      	movabs	rax, 0xf0f0f0f0f0f0f0f
  3a3535:      	and	rax, rcx
  3a3538:      	mov	rcx, rax
  3a353b:      	shr	rcx, 0x8
  3a353f:      	add	rcx, rax
  3a3542:      	mov	rdx, rcx
  3a3545:      	shr	rdx, 0x10
  3a3549:      	add	rdx, rcx
  3a354c:      	mov	rax, rdx
  3a354f:      	shr	rax, 0x20
  3a3553:      	add	eax, edx
  3a3555:      	and	eax, 0x7f
  3a3558:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a355d:      	xor	ecx, ecx
  3a355f:      	mov	rdx, rax
  3a3562:      	shr	rdx, 0x20
  3a3566:      	cmove	rdx, rax
  3a356a:      	setne	cl
  3a356d:      	shl	ecx, 0x5
  3a3570:      	mov	rax, rdx
  3a3573:      	shr	rax, 0x10
  3a3577:      	lea	rsi, [rcx + 0x10]
  3a357b:      	cmp	rdx, 0x10000
  3a3582:      	cmovb	rax, rdx
  3a3586:      	cmovb	rsi, rcx
  3a358a:      	mov	ecx, eax
  3a358c:      	shr	ecx, 0x8
  3a358f:      	lea	rdx, [rsi + 0x8]
  3a3593:      	cmp	rax, 0x100
  3a3599:      	cmovb	rcx, rax
  3a359d:      	cmovb	rdx, rsi
  3a35a1:      	movzx	eax, cl
  3a35a4:      	movzx	ecx, byte ptr [rax + 0x328641]
  3a35ab:      	add	rcx, rdx
  3a35ae:      	mov	eax, 0x40
  3a35b3:      	sub	rax, rcx
  3a35b6:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a35bb:      	xor	edx, edx
  3a35bd:      	cmp	rax, rcx
  3a35c0:      	setle	dl
  3a35c3:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a35c8:      	xor	edx, edx
  3a35ca:      	cmp	rax, rcx
  3a35cd:      	setae	dl
  3a35d0:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a35d5:      	add	ecx, eax
  3a35d7:      	jmp	0x3a377e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbfe>
  3a35dc:      	test	eax, eax
  3a35de:      	je	0x3a391f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd9f>
  3a35e4:      	tzcnt	eax, eax
  3a35e8:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a35ed:      	shr	eax, cl
  3a35ef:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a35f4:      	and	eax, ecx
  3a35f6:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a35fb:      	xor	edx, edx
  3a35fd:      	cmp	rax, rcx
  3a3600:      	sete	dl
  3a3603:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3608:      	sub	rax, rcx
  3a360b:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3610:      	xor	ecx, ecx
  3a3612:      	test	rax, rax
  3a3615:      	sete	cl
  3a3618:      	jmp	0x3a377e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbfe>
  3a361d:      	mov	ecx, eax
  3a361f:      	shr	ecx, 0x10
  3a3622:      	xor	edx, edx
  3a3624:      	cmp	eax, 0x10000
  3a3629:      	setb	dl
  3a362c:      	cmovb	ecx, eax
  3a362f:      	shl	edx, 0x4
  3a3632:      	mov	esi, ecx
  3a3634:      	shr	esi, 0x8
  3a3637:      	xor	eax, eax
  3a3639:      	cmp	ecx, 0x100
  3a363f:      	setb	al
  3a3642:      	lea	eax, [rdx + 8*rax]
  3a3645:      	cmovb	esi, ecx
  3a3648:      	movzx	ecx, sil
  3a364c:      	movzx	ecx, byte ptr [rcx + 0x328641]
  3a3653:      	sub	rax, rcx
  3a3656:      	add	rax, 0x8
  3a365a:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a365f:      	xor	edx, edx
  3a3661:      	cmp	rax, rcx
  3a3664:      	setbe	dl
  3a3667:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a366c:      	or	ecx, eax
  3a366e:      	jmp	0x3a377e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbfe>
  3a3673:      	shl	eax, cl
  3a3675:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a367a:      	xor	edx, edx
  3a367c:      	cmp	eax, ecx
  3a367e:      	seta	dl
  3a3681:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3686:      	xor	edx, edx
  3a3688:      	cmp	eax, ecx
  3a368a:      	setle	dl
  3a368d:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3692:      	xor	edx, edx
  3a3694:      	cmp	eax, ecx
  3a3696:      	setb	dl
  3a3699:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a369e:      	mov	ecx, eax
  3a36a0:      	shr	ecx, 0x10
  3a36a3:      	movzx	ecx, cl
  3a36a6:      	movzx	ecx, byte ptr [rcx + 0x328741]
  3a36ad:      	movzx	edx, ah
  3a36b0:      	movzx	esi, al
  3a36b3:      	shr	eax, 0x18
  3a36b6:      	add	cl, byte ptr [rax + 0x328741]
  3a36bc:      	add	cl, byte ptr [rdx + 0x328741]
  3a36c2:      	add	cl, byte ptr [rsi + 0x328741]
  3a36c8:      	movzx	eax, cl
  3a36cb:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a36d0:      	imul	rax, rcx
  3a36d4:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a36d9:      	xor	edx, edx
  3a36db:      	cmp	eax, ecx
  3a36dd:      	setne	dl
  3a36e0:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a36e5:      	xor	edx, edx
  3a36e7:      	cmp	eax, ecx
  3a36e9:      	sete	dl
  3a36ec:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a36f1:      	xor	edx, edx
  3a36f3:      	cmp	eax, ecx
  3a36f5:      	setge	dl
  3a36f8:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a36fa:      	xor	edx, edx
  3a36fc:      	cmp	eax, ecx
  3a36fe:      	setbe	dl
  3a3701:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3703:      	xor	edx, edx
  3a3705:      	cmp	rax, rcx
  3a3708:      	setl	dl
  3a370b:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a370d:      	xor	edx, edx
  3a370f:      	cmp	rax, rcx
  3a3712:      	setb	dl
  3a3715:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3717:      	xor	ecx, eax
  3a3719:      	jmp	0x3a377e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbfe>
  3a371b:      	xor	edx, edx
  3a371d:      	cmp	eax, ecx
  3a371f:      	setg	dl
  3a3722:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3724:      	mov	ecx, 0x40
  3a3729:      	tzcnt	rcx, rax
  3a372e:      	jmp	0x3a377e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbfe>
  3a3730:      	xor	edx, edx
  3a3732:      	cmp	eax, ecx
  3a3734:      	setl	dl
  3a3737:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3739:      	xor	edx, edx
  3a373b:      	cmp	rax, rcx
  3a373e:      	seta	dl
  3a3741:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3743:      	rol	eax, cl
  3a3745:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a374a:      	xor	edx, edx
  3a374c:      	cmp	eax, ecx
  3a374e:      	setae	dl
  3a3751:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3753:      	xor	edx, edx
  3a3755:      	cmp	rax, rcx
  3a3758:      	setg	dl
  3a375b:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a375d:      	xor	edx, edx
  3a375f:      	cmp	rax, rcx
  3a3762:      	setne	dl
  3a3765:      	jmp	0x3a3770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbf0>
  3a3767:      	xor	edx, edx
  3a3769:      	cmp	rax, rcx
  3a376c:      	setge	dl
  3a376f:      	nop
  3a3770:      	xor	r8d, r8d
  3a3773:      	mov	rax, rdx
  3a3776:      	jmp	0x3a3933 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb3>
  3a377b:      	imul	ecx, eax
  3a377e:      	xor	r8d, r8d
  3a3781:      	mov	rax, rcx
  3a3784:      	jmp	0x3a3933 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb3>
  3a3789:      	sub	eax, ecx
  3a378b:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3790:      	ror	eax, cl
  3a3792:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3797:      	sar	eax, cl
  3a3799:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a379e:      	xor	r8d, r8d
  3a37a1:      	mov	edi, dword ptr [r13 + 0x10]
  3a37a5:      	cmp	rdi, r10
  3a37a8:      	jne	0x3a3089 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x509>
  3a37ae:      	xor	eax, eax
  3a37b0:      	xor	r9d, r9d
  3a37b3:      	mov	r10d, 0x0
  3a37b9:      	mov	r11d, 0x0
  3a37bf:      	cmp	dx, 0xc
  3a37c3:      	jne	0x3a3801 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc81>
  3a37c5:      	mov	rcx, qword ptr [rbp - 0x60]
  3a37c9:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a37d0:      	mov	edi, dword ptr [r13 + 0x18]
  3a37d4:      	add	edi, 0x2
  3a37d7:      	cmp	rsi, rdi
  3a37da:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a37e0:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a37e7:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a37ea:      	cmp	qword ptr [rbp - 0x38], rdi
  3a37ee:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18a1>
  3a37f4:      	shl	rdi, 0x4
  3a37f8:      	mov	r10, qword ptr [r12 + rdi]
  3a37fc:      	mov	r11, qword ptr [r12 + rdi + 0x8]
  3a3801:      	cmp	qword ptr [rbp - 0x48], 0x0
  3a3806:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a380c:      	mov	rdx, qword ptr [rbp - 0x60]
  3a3810:      	mov	rsi, qword ptr [rdx + 0x80]
  3a3817:      	mov	edi, dword ptr [r13 + 0x4]
  3a381b:      	cmp	rsi, rdi
  3a381e:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3824:      	imul	rcx, rdi, 0xe0
  3a382b:      	add	rcx, qword ptr [rdx + 0x78]
  3a382f:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3835:      	mov	rdx, qword ptr [rbp - 0x48]
  3a3839:      	mov	rsi, qword ptr [rdx]
  3a383c:      	mov	qword ptr [rbp - 0xd0], rax
  3a3843:      	mov	qword ptr [rbp - 0xc8], r9
  3a384a:      	mov	qword ptr [rbp - 0xc0], r10
  3a3851:      	mov	qword ptr [rbp - 0xb8], r11
  3a3858:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a385f:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a3864:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a386b:      	movups	xmmword ptr [rsp], xmm0
  3a386f:      	lea	rdi, [rbp - 0x180]
  3a3876:      	mov	rdx, qword ptr [rbp - 0x58]
  3a387a:      	call	0x3b6930 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a387f:      	movaps	xmm0, xmmword ptr [rbp - 0x180]
  3a3886:      	mov	rax, qword ptr [rbp - 0x170]
  3a388d:      	test	rax, rax
  3a3890:      	je	0x3a38b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd30>
  3a3892:      	cmp	qword ptr [rax + 0x8], 0x0
  3a3897:      	mov	r10d, 0xffffffff
  3a389d:      	mov	r9, qword ptr [rbp - 0x38]
  3a38a1:      	mov	r8, qword ptr [rbp - 0x68]
  3a38a5:      	mov	rcx, qword ptr [rbp - 0x40]
  3a38a9:      	je	0x3a38c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd42>
  3a38ab:      	jmp	0x3a43b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1830>
  3a38b0:      	mov	r10d, 0xffffffff
  3a38b6:      	mov	r9, qword ptr [rbp - 0x38]
  3a38ba:      	mov	r8, qword ptr [rbp - 0x68]
  3a38be:      	mov	rcx, qword ptr [rbp - 0x40]
  3a38c2:      	mov	edi, dword ptr [r13 + 0x8]
  3a38c6:      	cmp	rdi, r10
  3a38c9:      	mov	r11, qword ptr [rbp - 0x48]
  3a38cd:      	mov	r12, qword ptr [rbp - 0x70]
  3a38d1:      	je	0x3a30fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x57a>
  3a38d7:      	cmp	r9, rdi
  3a38da:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a38e0:      	shl	rdi, 0x4
  3a38e4:      	mov	rax, qword ptr [rbp - 0x50]
  3a38e8:      	movups	xmmword ptr [rax + rdi], xmm0
  3a38ec:      	jmp	0x3a30fa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x57a>
  3a38f1:      	mov	qword ptr [rsp], r11
  3a38f5:      	mov	rsi, rcx
  3a38f8:      	mov	rdx, r10
  3a38fb:      	mov	rcx, rax
  3a38fe:      	call	0x3cebd0 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a3903:      	mov	r8, rdx
  3a3906:      	mov	r11, qword ptr [rbp - 0x48]
  3a390a:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a390f:      	mov	r10d, 0xffffffff
  3a3915:      	mov	r9, qword ptr [rbp - 0x38]
  3a3919:      	mov	r12, qword ptr [rbp - 0x50]
  3a391d:      	jmp	0x3a3937 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb7>
  3a391f:      	mov	eax, 0x20
  3a3924:      	nop	word ptr cs:[rax + rax]
  3a3930:      	xor	r8d, r8d
  3a3933:      	mov	r9, qword ptr [rbp - 0x38]
  3a3937:      	mov	rcx, qword ptr [rbp - 0x40]
  3a393b:      	mov	edi, dword ptr [r13 + 0x8]
  3a393f:      	cmp	rdi, r10
  3a3942:      	je	0x3a2c7d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a3948:      	cmp	r9, rdi
  3a394b:      	ja	0x3a2c70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0>
  3a3951:      	jmp	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a3956:      	mov	qword ptr [rbp - 0xb0], rax
  3a395d:      	mov	qword ptr [rbp - 0xa8], r8
  3a3964:      	movups	xmm0, xmmword ptr [rbp - 0xb0]
  3a396b:      	movups	xmmword ptr [rsp], xmm0
  3a396f:      	movzx	r8d, dil
  3a3973:      	mov	rdi, rsi
  3a3976:      	mov	esi, edx
  3a3978:      	mov	rdx, rcx
  3a397b:      	mov	rcx, r9
  3a397e:      	call	0x3c8300 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a3983:      	mov	edx, eax
  3a3985:      	xor	r8d, r8d
  3a3988:      	xor	eax, eax
  3a398a:      	mov	r11, qword ptr [rbp - 0x48]
  3a398e:      	mov	r10d, 0xffffffff
  3a3994:      	mov	r9, qword ptr [rbp - 0x38]
  3a3998:      	mov	rcx, qword ptr [rbp - 0x40]
  3a399c:      	test	dl, dl
  3a399e:      	je	0x3a393b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdbb>
  3a39a0:      	jmp	0x3a4430 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18b0>
  3a39a5:      	or	rax, rcx
  3a39a8:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39aa:      	mov	eax, eax
  3a39ac:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39ae:      	shr	rax, cl
  3a39b1:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39b6:      	and	rax, rcx
  3a39b9:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39be:      	rol	rax, cl
  3a39c1:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39c6:      	ror	rax, cl
  3a39c9:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39ce:      	mov	eax, eax
  3a39d0:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39d5:      	movsx	rax, ax
  3a39d9:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39de:      	cdqe
  3a39e0:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39e5:      	cdqe
  3a39e7:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39ec:      	sar	rax, cl
  3a39ef:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39f4:      	movsx	rax, ax
  3a39f8:      	mov	eax, eax
  3a39fa:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a39ff:      	shl	rax, cl
  3a3a02:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3a07:      	movsx	rax, al
  3a3a0b:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3a10:      	movsx	rax, al
  3a3a14:      	mov	eax, eax
  3a3a16:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3a1b:      	xor	rax, rcx
  3a3a1e:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3a23:      	lea	esi, [rdi - 0x1]
  3a3a26:      	cmp	esi, 0x108
  3a3a2c:      	ja	0x3a3bcb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x104b>
  3a3a32:      	jmp	qword ptr [8*rsi + 0x2038e8]
  3a3a39:      	mov	rdx, qword ptr [rbp - 0x60]
  3a3a3d:      	mov	rsi, qword ptr [rdx + 0x80]
  3a3a44:      	mov	ecx, dword ptr [r13 + 0x4]
  3a3a48:      	cmp	rsi, rcx
  3a3a4b:      	jbe	0x3a45d6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a56>
  3a3a51:      	imul	r8, rcx, 0xe0
  3a3a58:      	add	r8, qword ptr [rdx + 0x78]
  3a3a5c:      	mov	rdx, qword ptr [rbp - 0x58]
  3a3a60:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3a66:      	mov	rsi, qword ptr [rdx + 0x78]
  3a3a6a:      	mov	ecx, dword ptr [r8 + 0x18]
  3a3a6e:      	cmp	rsi, rcx
  3a3a71:      	jbe	0x3a45d6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a56>
  3a3a77:      	shl	rcx, 0x3
  3a3a7b:      	add	rcx, qword ptr [rdx + 0x70]
  3a3a7f:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3a85:      	mov	rcx, qword ptr [rcx]
  3a3a88:      	test	rcx, rcx
  3a3a8b:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3a91:      	mov	rdx, rdi
  3a3a94:      	mov	rsi, qword ptr [rcx + 0x10]
  3a3a98:      	mov	ecx, dword ptr [r8 + 0x14]
  3a3a9c:      	mov	r8d, dword ptr [r8 + 0x1c]
  3a3aa0:      	shl	r8, 0x20
  3a3aa4:      	or	r8, rcx
  3a3aa7:      	xorps	xmm0, xmm0
  3a3aaa:      	movaps	xmmword ptr [rbp - 0x260], xmm0
  3a3ab1:      	movups	xmmword ptr [rsp], xmm0
  3a3ab5:      	lea	rdi, [rbp - 0x108]
  3a3abc:      	mov	rcx, rax
  3a3abf:      	xor	r9d, r9d
  3a3ac2:      	call	0x3c6aa0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Load>
  3a3ac7:      	mov	rax, qword ptr [rbp - 0x108]
  3a3ace:      	mov	r8, qword ptr [rbp - 0x100]
  3a3ad5:      	movzx	edx, byte ptr [rbp - 0xf8]
  3a3adc:      	jmp	0x3a398a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe0a>
  3a3ae1:      	movzx	ecx, dl
  3a3ae4:      	xor	edi, edi
  3a3ae6:      	add	ecx, -0x2
  3a3ae9:      	cmp	ecx, 0x7
  3a3aec:      	ja	0x3a3b66 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe6>
  3a3aee:      	jmp	qword ptr [8*rcx + 0x204130]
  3a3af5:      	xor	edi, edi
  3a3af7:      	test	eax, eax
  3a3af9:      	sete	dil
  3a3afd:      	jmp	0x3a3b66 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe6>
  3a3aff:      	mov	edi, dword ptr [r13 + 0x18]
  3a3b03:      	mov	r9, qword ptr [rbp - 0x38]
  3a3b07:      	cmp	r9, rdi
  3a3b0a:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a3b10:      	shl	rdi, 0x4
  3a3b14:      	mov	qword ptr [r12 + rdi], rax
  3a3b18:      	mov	qword ptr [r12 + rdi + 0x8], r8
  3a3b1d:      	jmp	0x3a3937 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb7>
  3a3b22:      	mov	edi, dword ptr [r13 + 0x14]
  3a3b26:      	cmp	rdi, r10
  3a3b29:      	je	0x3a3e6d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12ed>
  3a3b2f:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3b33:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18a1>
  3a3b39:      	shl	rdi, 0x4
  3a3b3d:      	mov	rdx, qword ptr [r12 + rdi]
  3a3b41:      	test	edx, edx
  3a3b43:      	cmove	rax, rcx
  3a3b47:      	cmove	r8, r9
  3a3b4b:      	jmp	0x3a3933 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb3>
  3a3b50:      	mov	edi, dword ptr [r13 + 0x1c]
  3a3b54:      	dec	edi
  3a3b56:      	cmp	edi, eax
  3a3b58:      	cmovae	edi, eax
  3a3b5b:      	jmp	0x3a3b66 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe6>
  3a3b5d:      	xor	edi, edi
  3a3b5f:      	test	rax, rax
  3a3b62:      	setne	dil
  3a3b66:      	mov	rax, qword ptr [rbp - 0x60]
  3a3b6a:      	mov	rsi, qword ptr [rax + 0xe8]
  3a3b71:      	add	edi, dword ptr [r13 + 0x18]
  3a3b75:      	cmp	rsi, rdi
  3a3b78:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3b7e:      	mov	rax, qword ptr [rax + 0xe0]
  3a3b85:      	lea	rcx, [rdi + 2*rdi]
  3a3b89:      	mov	r12d, dword ptr [rax + 4*rcx]
  3a3b8d:      	mov	r8d, dword ptr [rax + 4*rcx + 0x4]
  3a3b92:      	mov	r13d, dword ptr [rax + 4*rcx + 0x8]
  3a3b97:      	cmp	r12d, r14d
  3a3b9a:      	ja	0x3a3db3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1233>
  3a3ba0:      	mov	r14, r8
  3a3ba3:      	mov	rdi, r11
  3a3ba6:      	call	0x3a7840 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).poll>
  3a3bab:      	test	rax, rax
  3a3bae:      	je	0x3a3db0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1230>
  3a3bb4:      	cmp	qword ptr [rax + 0x8], 0x0
  3a3bb9:      	mov	rdi, qword ptr [rbp - 0x48]
  3a3bbd:      	mov	r8, r14
  3a3bc0:      	je	0x3a3db3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1233>
  3a3bc6:      	jmp	0x3a43a8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1828>
  3a3bcb:      	lea	esi, [rdi - 0x10]
  3a3bce:      	cmp	esi, 0x34
  3a3bd1:      	ja	0x3a3e78 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12f8>
  3a3bd7:      	jmp	qword ptr [8*rsi + 0x204170]
  3a3bde:      	mov	rax, qword ptr [r13 + 0x18]
  3a3be2:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3be7:      	mov	r10, qword ptr [rbp - 0x60]
  3a3beb:      	mov	rsi, qword ptr [r10 + 0x80]
  3a3bf2:      	mov	edx, dword ptr [r13 + 0x4]
  3a3bf6:      	cmp	rsi, rdx
  3a3bf9:      	jbe	0x3a45de <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a5e>
  3a3bff:      	imul	r8, rdx, 0xe0
  3a3c06:      	add	r8, qword ptr [r10 + 0x78]
  3a3c0a:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3c10:      	mov	r10, qword ptr [rbp - 0x58]
  3a3c14:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3c18:      	mov	edx, dword ptr [r8 + 0x18]
  3a3c1c:      	cmp	rsi, rdx
  3a3c1f:      	jbe	0x3a45de <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a5e>
  3a3c25:      	shl	rdx, 0x3
  3a3c29:      	add	rdx, qword ptr [r10 + 0x70]
  3a3c2d:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3c33:      	mov	rdx, qword ptr [rdx]
  3a3c36:      	test	rdx, rdx
  3a3c39:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3c3f:      	mov	rsi, rdi
  3a3c42:      	mov	rdi, qword ptr [rdx + 0x10]
  3a3c46:      	mov	edx, dword ptr [r8 + 0x14]
  3a3c4a:      	mov	r8d, dword ptr [r8 + 0x1c]
  3a3c4e:      	shl	r8, 0x20
  3a3c52:      	or	r8, rdx
  3a3c55:      	mov	qword ptr [rbp - 0xe0], rcx
  3a3c5c:      	mov	qword ptr [rbp - 0xd8], r9
  3a3c63:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a3c6a:      	movups	xmmword ptr [rsp], xmm0
  3a3c6e:      	mov	rdx, rax
  3a3c71:      	mov	rcx, r8
  3a3c74:      	xor	r8d, r8d
  3a3c77:      	call	0x3c7fc0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Store>
  3a3c7c:      	jmp	0x3a3983 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe03>
  3a3c81:      	mov	rax, qword ptr [rbp - 0x58]
  3a3c85:      	mov	rsi, qword ptr [rax + 0x90]
  3a3c8c:      	mov	edi, dword ptr [r13 + 0x18]
  3a3c90:      	cmp	rsi, rdi
  3a3c93:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3c99:      	mov	rax, qword ptr [rax + 0x88]
  3a3ca0:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a3ca4:      	mov	rax, qword ptr [rcx + 0x20]
  3a3ca8:      	mov	r8, qword ptr [rcx + 0x28]
  3a3cac:      	jmp	0x3a3933 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb3>
  3a3cb1:      	mov	rcx, qword ptr [rbp - 0x58]
  3a3cb5:      	mov	rsi, qword ptr [rcx + 0x90]
  3a3cbc:      	mov	edi, dword ptr [r13 + 0x18]
  3a3cc0:      	cmp	rsi, rdi
  3a3cc3:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3cc9:      	mov	rcx, qword ptr [rcx + 0x88]
  3a3cd0:      	mov	rcx, qword ptr [rcx + 8*rdi]
  3a3cd4:      	mov	qword ptr [rcx + 0x20], rax
  3a3cd8:      	mov	qword ptr [rcx + 0x28], r8
  3a3cdc:      	xor	eax, eax
  3a3cde:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3ce3:      	mov	edi, dword ptr [r13 + 0x18]
  3a3ce7:      	mov	r9, qword ptr [rbp - 0x38]
  3a3ceb:      	cmp	r9, rdi
  3a3cee:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a3cf4:      	shl	rdi, 0x4
  3a3cf8:      	mov	rax, qword ptr [r12 + rdi]
  3a3cfc:      	mov	r8, qword ptr [r12 + rdi + 0x8]
  3a3d01:      	jmp	0x3a3937 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb7>
  3a3d06:      	mov	rcx, qword ptr [rbp - 0x58]
  3a3d0a:      	mov	rsi, qword ptr [rcx + 0x60]
  3a3d0e:      	mov	edi, dword ptr [r13 + 0x18]
  3a3d12:      	cmp	rsi, rdi
  3a3d15:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3d1b:      	mov	rcx, qword ptr [rcx + 0x58]
  3a3d1f:      	lea	rdx, [rcx + 8*rdi]
  3a3d23:      	test	rdx, rdx
  3a3d26:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3d2c:      	mov	rcx, qword ptr [rcx + 8*rdi]
  3a3d30:      	test	rcx, rcx
  3a3d33:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3d39:      	cmp	byte ptr [rcx + 0x29], 0x0
  3a3d3d:      	mov	edx, eax
  3a3d3f:      	cmovne	rdx, rax
  3a3d43:      	cmp	rdx, qword ptr [rcx + 0x40]
  3a3d47:      	jae	0x3a442a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18aa>
  3a3d4d:      	mov	rax, qword ptr [rcx + 0x38]
  3a3d51:      	mov	rax, qword ptr [rax + 8*rdx]
  3a3d55:      	jmp	0x3a3da8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1228>
  3a3d57:      	mov	rdx, qword ptr [rbp - 0x58]
  3a3d5b:      	mov	rsi, qword ptr [rdx + 0x60]
  3a3d5f:      	mov	edi, dword ptr [r13 + 0x18]
  3a3d63:      	cmp	rsi, rdi
  3a3d66:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3d6c:      	mov	rdx, qword ptr [rdx + 0x58]
  3a3d70:      	lea	rsi, [rdx + 8*rdi]
  3a3d74:      	test	rsi, rsi
  3a3d77:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3d7d:      	mov	rdx, qword ptr [rdx + 8*rdi]
  3a3d81:      	test	rdx, rdx
  3a3d84:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3d8a:      	cmp	byte ptr [rdx + 0x29], 0x0
  3a3d8e:      	mov	esi, eax
  3a3d90:      	cmovne	rsi, rax
  3a3d94:      	cmp	rsi, qword ptr [rdx + 0x40]
  3a3d98:      	jae	0x3a442a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18aa>
  3a3d9e:      	mov	rax, qword ptr [rdx + 0x38]
  3a3da2:      	mov	qword ptr [rax + 8*rsi], rcx
  3a3da6:      	xor	eax, eax
  3a3da8:      	xor	r8d, r8d
  3a3dab:      	jmp	0x3a3906 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd86>
  3a3db0:      	mov	r8, r14
  3a3db3:      	mov	rax, qword ptr [rbp - 0x60]
  3a3db7:      	mov	rsi, qword ptr [rax + 0xf8]
  3a3dbe:      	mov	rcx, qword ptr [rax + 0x108]
  3a3dc5:      	add	r13d, r8d
  3a3dc8:      	mov	dword ptr [rsp + 0x10], 0x0
  3a3dd0:      	mov	dword ptr [rsp + 0x8], 0x0
  3a3dd8:      	mov	dword ptr [rsp], 0x0
  3a3ddf:      	mov	edx, 0x8
  3a3de4:      	lea	rdi, [rbp - 0x1a8]
  3a3deb:      	mov	r9, r13
  3a3dee:      	call	0x5f4870 <github.com/xgo-dev/llgo/runtime/internal/runtime.NewSlice2>
  3a3df3:      	mov	rcx, qword ptr [rbp - 0x1a0]
  3a3dfa:      	test	rcx, rcx
  3a3dfd:      	jle	0x3a3e5a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12da>
  3a3dff:      	mov	rdx, qword ptr [rbp - 0x1a8]
  3a3e06:      	xor	esi, esi
  3a3e08:      	mov	r11, qword ptr [rbp - 0x48]
  3a3e0c:      	mov	r10d, 0xffffffff
  3a3e12:      	mov	r9, qword ptr [rbp - 0x38]
  3a3e16:      	mov	edi, dword ptr [rdx + 8*rsi + 0x4]
  3a3e1a:      	cmp	r9, rdi
  3a3e1d:      	jbe	0x3a4419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a3e23:      	mov	eax, dword ptr [rdx + 8*rsi]
  3a3e26:      	cmp	r9, rax
  3a3e29:      	jbe	0x3a45bc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a3c>
  3a3e2f:      	shl	rdi, 0x4
  3a3e33:      	shl	rax, 0x4
  3a3e37:      	mov	r8, qword ptr [rbp - 0x50]
  3a3e3b:      	movups	xmm0, xmmword ptr [r8 + rdi]
  3a3e40:      	movups	xmmword ptr [r8 + rax], xmm0
  3a3e45:      	inc	rsi
  3a3e48:      	cmp	rcx, rsi
  3a3e4b:      	mov	r8, qword ptr [rbp - 0x68]
  3a3e4f:      	jne	0x3a3e16 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1296>
  3a3e51:      	mov	rcx, qword ptr [rbp - 0x40]
  3a3e55:      	jmp	0x3a2c87 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x107>
  3a3e5a:      	mov	r11, qword ptr [rbp - 0x48]
  3a3e5e:      	mov	r10d, 0xffffffff
  3a3e64:      	mov	r9, qword ptr [rbp - 0x38]
  3a3e68:      	jmp	0x3a34a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x929>
  3a3e6d:      	mov	rax, rcx
  3a3e70:      	mov	r8, r9
  3a3e73:      	jmp	0x3a3933 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb3>
  3a3e78:      	lea	esi, [rdi - 0xd0]
  3a3e7e:      	cmp	esi, 0x4
  3a3e81:      	ja	0x3a3e9a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131a>
  3a3e83:      	jmp	qword ptr [8*rsi + 0x204318]
  3a3e8a:      	xor	ecx, ecx
  3a3e8c:      	test	rax, rax
  3a3e8f:      	sete	cl
  3a3e92:      	mov	rax, rcx
  3a3e95:      	jmp	0x3a3da8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1228>
  3a3e9a:      	mov	rsi, rdi
  3a3e9d:      	mov	edi, dword ptr [r13 + 0x14]
  3a3ea1:      	cmp	rdi, r10
  3a3ea4:      	je	0x3a400a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x148a>
  3a3eaa:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3eae:      	jbe	0x3a4421 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18a1>
  3a3eb4:      	shl	rdi, 0x4
  3a3eb8:      	mov	r8, qword ptr [rbp - 0x50]
  3a3ebc:      	movups	xmm0, xmmword ptr [r8 + rdi]
  3a3ec1:      	jmp	0x3a400d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x148d>
  3a3ec6:      	mov	eax, dword ptr [r13 + 0x18]
  3a3eca:      	jmp	0x3a3930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdb0>
  3a3ecf:      	mov	rax, qword ptr [rbp - 0x58]
  3a3ed3:      	mov	rsi, qword ptr [rax + 0x78]
  3a3ed7:      	mov	edi, dword ptr [r13 + 0x18]
  3a3edb:      	cmp	rsi, rdi
  3a3ede:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3ee4:      	shl	rdi, 0x3
  3a3ee8:      	mov	rax, qword ptr [rbp - 0x58]
  3a3eec:      	add	rdi, qword ptr [rax + 0x70]
  3a3ef0:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3ef6:      	mov	rax, qword ptr [rdi]
  3a3ef9:      	test	rax, rax
  3a3efc:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3f02:      	mov	rdi, qword ptr [rax + 0x10]
  3a3f06:      	call	0x3c46c0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Pages>
  3a3f0b:      	jmp	0x3a3da8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1228>
  3a3f10:      	mov	rcx, qword ptr [rbp - 0x58]
  3a3f14:      	mov	rsi, qword ptr [rcx + 0x78]
  3a3f18:      	mov	edi, dword ptr [r13 + 0x18]
  3a3f1c:      	cmp	rsi, rdi
  3a3f1f:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3f25:      	shl	rdi, 0x3
  3a3f29:      	mov	rcx, qword ptr [rbp - 0x58]
  3a3f2d:      	add	rdi, qword ptr [rcx + 0x70]
  3a3f31:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3f37:      	cmp	qword ptr [rbp - 0x48], 0x0
  3a3f3c:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3f42:      	mov	r15, qword ptr [rdi]
  3a3f45:      	test	r15, r15
  3a3f48:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a3f4e:      	mov	rcx, qword ptr [rbp - 0x48]
  3a3f52:      	mov	rdi, qword ptr [rcx]
  3a3f55:      	mov	rcx, qword ptr [r15 + 0x10]
  3a3f59:      	movzx	ecx, byte ptr [rcx + 0x31]
  3a3f5d:      	mov	byte ptr [rbp - 0x71], cl
  3a3f60:      	mov	edx, eax
  3a3f62:      	test	cl, 0x1
  3a3f65:      	cmovne	rdx, rax
  3a3f69:      	mov	rsi, r15
  3a3f6c:      	call	0x3c0ca0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a3f71:      	test	dl, 0x1
  3a3f74:      	jne	0x3a3da8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1228>
  3a3f7a:      	mov	rax, qword ptr [r15 + 0x10]
  3a3f7e:      	movzx	eax, byte ptr [rax + 0x31]
  3a3f82:      	mov	byte ptr [rbp - 0x72], al
  3a3f85:      	and	al, 0x1
  3a3f87:      	neg	al
  3a3f89:      	mov	eax, 0x0
  3a3f8e:      	sbb	rax, rax
  3a3f91:      	mov	r10d, 0xffffffff
  3a3f97:      	or	rax, r10
  3a3f9a:      	xor	r8d, r8d
  3a3f9d:      	mov	r11, qword ptr [rbp - 0x48]
  3a3fa1:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a3fa6:      	jmp	0x3a3915 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd95>
  3a3fab:      	test	rax, rax
  3a3fae:      	mov	r11, qword ptr [rbp - 0x48]
  3a3fb2:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a3fb7:      	mov	r10d, 0xffffffff
  3a3fbd:      	mov	r9, qword ptr [rbp - 0x38]
  3a3fc1:      	mov	r12, qword ptr [rbp - 0x50]
  3a3fc5:      	mov	rcx, qword ptr [rbp - 0x40]
  3a3fc9:      	jne	0x3a393b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdbb>
  3a3fcf:      	jmp	0x3a4590 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a10>
  3a3fd4:      	mov	rax, qword ptr [rbp - 0x58]
  3a3fd8:      	mov	rsi, qword ptr [rax + 0x48]
  3a3fdc:      	mov	edi, dword ptr [r13 + 0x18]
  3a3fe0:      	cmp	rsi, rdi
  3a3fe3:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a3fe9:      	mov	rax, qword ptr [rbp - 0x58]
  3a3fed:      	mov	rax, qword ptr [rax + 0x40]
  3a3ff1:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3ff5:      	jmp	0x3a3da8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1228>
  3a3ffa:      	xor	edx, edx
  3a3ffc:      	cmp	rax, rcx
  3a3fff:      	sete	dl
  3a4002:      	mov	rax, rdx
  3a4005:      	jmp	0x3a3da8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1228>
  3a400a:      	xorps	xmm0, xmm0
  3a400d:      	and	edx, 0xff0000
  3a4013:      	cmp	edx, 0xfe0000
  3a4019:      	jne	0x3a4052 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14d2>
  3a401b:      	movzx	r9d, si
  3a401f:      	cmp	r9d, 0x3
  3a4023:      	jne	0x3a41c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1645>
  3a4029:      	cmp	qword ptr [rbp - 0x48], 0x0
  3a402e:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a4034:      	mov	rax, qword ptr [rbp - 0x48]
  3a4038:      	mov	rax, qword ptr [rax]
  3a403b:      	test	rax, rax
  3a403e:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a4044:      	mov	rdi, qword ptr [rax + 0x8]
  3a4048:      	call	0x3c3cd0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a404d:      	jmp	0x3a3da6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1226>
  3a4052:      	add	esi, 0xff03fff8
  3a4058:      	cmp	esi, 0xa
  3a405b:      	jae	0x3a45e6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a66>
  3a4061:      	mov	rdx, qword ptr [rbp - 0x60]
  3a4065:      	mov	rsi, qword ptr [rdx + 0x80]
  3a406c:      	mov	edi, dword ptr [r13 + 0x4]
  3a4070:      	cmp	rsi, rdi
  3a4073:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a4079:      	imul	rdx, rdi, 0xe0
  3a4080:      	mov	rsi, qword ptr [rbp - 0x60]
  3a4084:      	add	rdx, qword ptr [rsi + 0x78]
  3a4088:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a408e:      	movups	xmm1, xmmword ptr [rdx + 0x98]
  3a4095:      	movups	xmm2, xmmword ptr [rdx + 0x88]
  3a409c:      	movups	xmm3, xmmword ptr [rdx + 0x78]
  3a40a0:      	movups	xmm4, xmmword ptr [rdx + 0x68]
  3a40a4:      	movups	xmm5, xmmword ptr [rdx + 0x58]
  3a40a8:      	movups	xmm6, xmmword ptr [rdx + 0x48]
  3a40ac:      	movzx	esi, byte ptr [rdx + 0x40]
  3a40b0:      	movups	xmm7, xmmword ptr [rdx + 0x30]
  3a40b4:      	movups	xmm8, xmmword ptr [rdx + 0x20]
  3a40b9:      	movups	xmm9, xmmword ptr [rdx + 0x10]
  3a40be:      	mov	rdi, qword ptr [rdx + 0x8]
  3a40c2:      	mov	edx, dword ptr [rdx]
  3a40c4:      	mov	dword ptr [rbp - 0x250], edx
  3a40ca:      	mov	qword ptr [rbp - 0x248], rdi
  3a40d1:      	movups	xmmword ptr [rbp - 0x240], xmm9
  3a40d9:      	movups	xmmword ptr [rbp - 0x230], xmm8
  3a40e1:      	movups	xmmword ptr [rbp - 0x220], xmm7
  3a40e8:      	mov	byte ptr [rbp - 0x210], sil
  3a40ef:      	movups	xmmword ptr [rbp - 0x208], xmm6
  3a40f6:      	movups	xmmword ptr [rbp - 0x1f8], xmm5
  3a40fd:      	movups	xmmword ptr [rbp - 0x1e8], xmm4
  3a4104:      	movups	xmmword ptr [rbp - 0x1d8], xmm3
  3a410b:      	movups	xmmword ptr [rbp - 0x1c8], xmm2
  3a4112:      	movups	xmmword ptr [rbp - 0x1b8], xmm1
  3a4119:      	mov	qword ptr [rbp - 0xf0], rcx
  3a4120:      	mov	qword ptr [rbp - 0xe8], r9
  3a4127:      	movaps	xmmword ptr [rbp - 0x190], xmm0
  3a412e:      	mov	ecx, 0x15
  3a4133:      	mov	rdi, rsp
  3a4136:      	lea	rsi, [rbp - 0x250]
  3a413d:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a4140:      	movaps	xmm0, xmmword ptr [rbp - 0x190]
  3a4147:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a414f:      	movups	xmm0, xmmword ptr [rbp - 0xf0]
  3a4156:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a415e:      	lea	rdi, [rbp - 0x130]
  3a4165:      	mov	rsi, qword ptr [rbp - 0x48]
  3a4169:      	mov	rcx, rax
  3a416c:      	mov	rdx, qword ptr [rbp - 0x58]
  3a4170:      	call	0x3a95a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a4175:      	mov	rax, qword ptr [rbp - 0x130]
  3a417c:      	mov	r8, qword ptr [rbp - 0x128]
  3a4183:      	movzx	edx, byte ptr [rbp - 0x120]
  3a418a:      	mov	rsi, qword ptr [rbp - 0x118]
  3a4191:      	test	rsi, rsi
  3a4194:      	je	0x3a42b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1739>
  3a419a:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a419f:      	mov	r11, qword ptr [rbp - 0x48]
  3a41a3:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a41a8:      	mov	r10d, 0xffffffff
  3a41ae:      	mov	r9, qword ptr [rbp - 0x38]
  3a41b2:      	mov	r12, qword ptr [rbp - 0x50]
  3a41b6:      	mov	rcx, qword ptr [rbp - 0x40]
  3a41ba:      	je	0x3a399c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe1c>
  3a41c0:      	jmp	0x3a459e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a1e>
  3a41c5:      	mov	rdx, qword ptr [rbp - 0x60]
  3a41c9:      	mov	rsi, qword ptr [rdx + 0x80]
  3a41d0:      	mov	edi, dword ptr [r13 + 0x4]
  3a41d4:      	cmp	rsi, rdi
  3a41d7:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a41dd:      	imul	r11, rdi, 0xe0
  3a41e4:      	mov	rdx, qword ptr [rbp - 0x60]
  3a41e8:      	add	r11, qword ptr [rdx + 0x78]
  3a41ec:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a41f2:      	mov	rdx, qword ptr [rbp - 0x58]
  3a41f6:      	mov	rsi, qword ptr [rdx + 0x78]
  3a41fa:      	mov	edi, dword ptr [r11 + 0x18]
  3a41fe:      	cmp	rsi, rdi
  3a4201:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a4207:      	shl	rdi, 0x3
  3a420b:      	mov	rdx, qword ptr [rbp - 0x58]
  3a420f:      	add	rdi, qword ptr [rdx + 0x70]
  3a4213:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a4219:      	mov	rsi, qword ptr [rdi]
  3a421c:      	test	rsi, rsi
  3a421f:      	je	0x3a4396 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1816>
  3a4225:      	mov	r10, qword ptr [rbp - 0x48]
  3a4229:      	mov	rdi, qword ptr [r10]
  3a422c:      	mov	rdx, qword ptr [r10 + 0x8]
  3a4230:      	mov	r8, qword ptr [rdi + 0x8]
  3a4234:      	mov	r10, qword ptr [r10 + 0x10]
  3a4238:      	mov	rsi, qword ptr [rsi + 0x10]
  3a423c:      	mov	edi, dword ptr [r11 + 0x14]
  3a4240:      	mov	r11d, dword ptr [r11 + 0x1c]
  3a4244:      	shl	r11, 0x20
  3a4248:      	or	r11, rdi
  3a424b:      	movlps	qword ptr [rsp + 0x18], xmm0
  3a4250:      	mov	qword ptr [rsp + 0x10], rcx
  3a4255:      	mov	qword ptr [rsp + 0x8], r11
  3a425a:      	mov	qword ptr [rsp], rax
  3a425e:      	lea	rdi, [rbp - 0x158]
  3a4265:      	mov	rcx, r10
  3a4268:      	call	0x3c4fe0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a426d:      	mov	rax, qword ptr [rbp - 0x158]
  3a4274:      	mov	r8, qword ptr [rbp - 0x150]
  3a427b:      	movzx	edx, byte ptr [rbp - 0x148]
  3a4282:      	mov	rsi, qword ptr [rbp - 0x140]
  3a4289:      	test	rsi, rsi
  3a428c:      	je	0x3a42bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x173b>
  3a428e:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a4293:      	mov	r11, qword ptr [rbp - 0x48]
  3a4297:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a429c:      	mov	r10d, 0xffffffff
  3a42a2:      	mov	r9, qword ptr [rbp - 0x38]
  3a42a6:      	mov	r12, qword ptr [rbp - 0x50]
  3a42aa:      	mov	rcx, qword ptr [rbp - 0x40]
  3a42ae:      	je	0x3a399c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe1c>
  3a42b4:      	jmp	0x3a45ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a2d>
  3a42b9:      	jmp	0x3a42bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x173b>
  3a42bb:      	mov	r11, qword ptr [rbp - 0x48]
  3a42bf:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a42c4:      	jmp	0x3a3401 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x881>
  3a42c9:      	mov	rax, qword ptr [rbp - 0x70]
  3a42cd:      	mov	dword ptr [rax - 0x38], r14d
  3a42d1:      	mov	rdi, r11
  3a42d4:      	mov	r14, r11
  3a42d7:      	call	0x3a69e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a42dc:      	test	al, al
  3a42de:      	jne	0x3a44a6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1926>
  3a42e4:      	mov	rcx, qword ptr [r14 + 0x48]
  3a42e8:      	test	rcx, rcx
  3a42eb:      	mov	r11, r14
  3a42ee:      	mov	r10d, 0xffffffff
  3a42f4:      	jg	0x3a2bc5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x45>
  3a42fa:      	jmp	0x3a45f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a75>
  3a42ff:      	mov	rax, qword ptr [rbp - 0x70]
  3a4303:      	mov	dword ptr [rax - 0x38], r14d
  3a4307:      	mov	rdi, r11
  3a430a:      	mov	r14, r11
  3a430d:      	call	0x3a64e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a4312:      	jmp	0x3a42dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x175c>
  3a4314:      	mov	rax, qword ptr [rbp - 0x70]
  3a4318:      	mov	dword ptr [rax - 0x38], r14d
  3a431c:      	mov	rdi, r11
  3a431f:      	mov	r14, r11
  3a4322:      	call	0x3a52f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a4327:      	test	al, al
  3a4329:      	je	0x3a42e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1764>
  3a432b:      	jmp	0x3a4597 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a17>
  3a4330:      	mov	rax, qword ptr [rbp - 0x70]
  3a4334:      	mov	dword ptr [rax - 0x38], r14d
  3a4338:      	mov	rdi, r11
  3a433b:      	mov	r14, r11
  3a433e:      	call	0x3a5780 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a4343:      	jmp	0x3a42dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x175c>
  3a4345:      	mov	edi, 0x28
  3a434a:      	call	0x5fc7f0 <GC_malloc@plt>
  3a434f:      	test	rax, rax
  3a4352:      	je	0x3a4604 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a84>
  3a4358:      	mov	r14, rax
  3a435b:      	add	r12, r15
  3a435e:      	mov	edi, 0x28
  3a4363:      	call	0x5e1e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a4368:      	mov	rax, qword ptr [r12 + 0x8]
  3a436d:      	movaps	xmm0, xmmword ptr [rip + 0x3153ac] # 0x6b9720 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a4374:      	movups	xmmword ptr [r14], xmm0
  3a4378:      	mov	qword ptr [r14 + 0x10], rax
  3a437c:      	mov	qword ptr [r14 + 0x18], 0x29a1e8
  3a4384:      	mov	qword ptr [r14 + 0x20], 0xe
  3a438c:      	mov	esi, 0x2babe8
  3a4391:      	jmp	0x3a4498 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1918>
  3a4396:      	mov	edi, 0x1
  3a439b:      	call	0x5e61e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a43a0:      	jmp	0x3a44a6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1926>
  3a43a5:      	mov	rdi, r11
  3a43a8:      	mov	rsi, rax
  3a43ab:      	jmp	0x3a449f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x191f>
  3a43b0:      	mov	r15, rax
  3a43b3:      	mov	r14, qword ptr [rbp - 0x168]
  3a43ba:      	mov	edi, 0x8
  3a43bf:      	call	0x5fc7f0 <GC_malloc@plt>
  3a43c4:      	test	rax, rax
  3a43c7:      	je	0x3a4604 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a84>
  3a43cd:      	mov	r12, rax
  3a43d0:      	mov	edi, 0x8
  3a43d5:      	call	0x5e1e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a43da:      	mov	edx, 0x29a158
  3a43df:      	mov	rdi, r15
  3a43e2:      	mov	rsi, r14
  3a43e5:      	mov	rcx, r12
  3a43e8:      	call	0x5b8aa0 <errors.As>
  3a43ed:      	test	al, 0x1
  3a43ef:      	je	0x3a4402 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1882>
  3a43f1:      	mov	rax, qword ptr [r12]
  3a43f5:      	mov	rcx, qword ptr [rbp - 0x88]
  3a43fc:      	mov	ecx, dword ptr [rcx + 0x50]
  3a43ff:      	mov	dword ptr [rax + 0x4], ecx
  3a4402:      	mov	rdi, qword ptr [rbp - 0x48]
  3a4406:      	mov	rsi, r15
  3a4409:      	jmp	0x3a449c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x191c>
  3a440e:      	mov	rdi, r14
  3a4411:      	mov	rsi, r8
  3a4414:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4419:      	mov	rsi, r9
  3a441c:      	call	0x5e1f50 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4421:      	mov	rsi, qword ptr [rbp - 0x38]
  3a4425:      	call	0x5e1f50 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a442a:      	mov	dl, 0x3
  3a442c:      	jmp	0x3a4430 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18b0>
  3a442e:      	mov	dl, 0x1
  3a4430:      	mov	r15d, edx
  3a4433:      	mov	edi, 0x10
  3a4438:      	call	0x5fc780 <GC_malloc_atomic@plt>
  3a443d:      	test	rax, rax
  3a4440:      	je	0x3a4604 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a84>
  3a4446:      	mov	r14, rax
  3a4449:      	mov	edi, 0x10
  3a444e:      	call	0x5e1e70 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a4453:      	xorps	xmm0, xmm0
  3a4456:      	movups	xmmword ptr [r14], xmm0
  3a445a:      	mov	rax, qword ptr [rbp - 0x60]
  3a445e:      	mov	rsi, qword ptr [rax + 0x80]
  3a4465:      	mov	edi, dword ptr [r13 + 0x4]
  3a4469:      	cmp	rsi, rdi
  3a446c:      	jbe	0x3a44bb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x193b>
  3a446e:      	mov	rax, qword ptr [rax + 0x78]
  3a4472:      	mov	rcx, qword ptr [rbp - 0x88]
  3a4479:      	mov	ecx, dword ptr [rcx + 0x50]
  3a447c:      	imul	rdx, rdi, 0xe0
  3a4483:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a4488:      	mov	byte ptr [r14], r15b
  3a448b:      	mov	dword ptr [r14 + 0x4], ecx
  3a448f:      	mov	qword ptr [r14 + 0x8], rax
  3a4493:      	mov	esi, 0x29a570
  3a4498:      	mov	rdi, qword ptr [rbp - 0x48]
  3a449c:      	mov	rdx, r14
  3a449f:      	call	0x3a2550 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a44a4:      	mov	al, 0x3
  3a44a6:      	mov	rdx, rbx
  3a44a9:      	add	rsp, 0x308
  3a44b0:      	pop	rbx
  3a44b1:      	pop	r12
  3a44b3:      	pop	r13
  3a44b5:      	pop	r14
  3a44b7:      	pop	r15
  3a44b9:      	pop	rbp
  3a44ba:      	ret
  3a44bb:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a44c0:      	mov	rdi, r9
  3a44c3:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a44c8:      	mov	rsi, qword ptr [rbp - 0x38]
  3a44cc:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a44d1:      	mov	rdi, rcx
  3a44d4:      	xor	esi, esi
  3a44d6:      	mov	ecx, 0x2
  3a44db:      	call	0x5e2c90 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a44e0:      	mov	rdi, rax
  3a44e3:      	mov	rsi, r9
  3a44e6:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a44eb:      	mov	rsi, rcx
  3a44ee:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a44f3:      	mov	rdi, rsi
  3a44f6:      	mov	rsi, rcx
  3a44f9:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a44fe:      	mov	rdi, rcx
  3a4501:      	mov	rsi, r9
  3a4504:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4509:      	mov	edi, 0x1
  3a450e:      	mov	esi, 0x1
  3a4513:      	call	0x5e1f50 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4518:      	xor	edi, edi
  3a451a:      	xor	esi, esi
  3a451c:      	call	0x5e1f50 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4521:      	xor	esi, esi
  3a4523:      	mov	rdx, rcx
  3a4526:      	mov	ecx, 0x3
  3a452b:      	call	0x5e2c90 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4530:      	mov	rdi, rax
  3a4533:      	mov	esi, 0x1
  3a4538:      	mov	ecx, 0x2
  3a453d:      	call	0x5e2c90 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4542:      	mov	rdi, r12
  3a4545:      	mov	esi, 0x1
  3a454a:      	mov	rdx, rax
  3a454d:      	mov	ecx, 0x3
  3a4552:      	call	0x5e2c90 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4557:      	mov	r8, qword ptr [rbp - 0x68]
  3a455b:      	mov	rdi, r13
  3a455e:      	mov	rsi, r8
  3a4561:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4566:      	mov	rdi, rax
  3a4569:      	mov	rsi, qword ptr [rbp - 0x38]
  3a456d:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4572:      	mov	edi, 0x3
  3a4577:      	mov	esi, 0x3
  3a457c:      	call	0x5e1f50 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4581:      	mov	edi, 0x2
  3a4586:      	mov	esi, 0x2
  3a458b:      	call	0x5e1f50 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4590:      	mov	dl, 0x9
  3a4592:      	jmp	0x3a4430 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18b0>
  3a4597:      	mov	al, 0x1
  3a4599:      	jmp	0x3a44a6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1926>
  3a459e:      	mov	rdx, qword ptr [rbp - 0x110]
  3a45a5:      	mov	rdi, r11
  3a45a8:      	jmp	0x3a449f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x191f>
  3a45ad:      	mov	rdx, qword ptr [rbp - 0x138]
  3a45b4:      	mov	rdi, r11
  3a45b7:      	jmp	0x3a449f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x191f>
  3a45bc:      	mov	rdi, rax
  3a45bf:      	mov	rsi, r9
  3a45c2:      	call	0x5e1f50 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a45c7:      	mov	edi, 0x329720
  3a45cc:      	mov	esi, 0x329a50
  3a45d1:      	call	0x5e1b90 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a45d6:      	mov	rdi, rcx
  3a45d9:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45de:      	mov	rdi, rdx
  3a45e1:      	call	0x5e3480 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45e6:      	mov	edi, 0x333bd8
  3a45eb:      	mov	esi, 0x29a1f8
  3a45f0:      	call	0x5e1b90 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a45f5:      	mov	edi, 0x333bd8
  3a45fa:      	mov	esi, 0x29a1d8
  3a45ff:      	call	0x5e1b90 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a4604:      	mov	edi, 0x333bd8
  3a4609:      	mov	esi, 0x3296f8
  3a460e:      	call	0x5e1b90 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a4613:      	int3
  3a4614:      	int3
  3a4615:      	int3
  3a4616:      	int3
  3a4617:      	int3
  3a4618:      	int3
  3a4619:      	int3
  3a461a:      	int3
  3a461b:      	int3
  3a461c:      	int3
  3a461d:      	int3
  3a461e:      	int3
  3a461f:      	int3
