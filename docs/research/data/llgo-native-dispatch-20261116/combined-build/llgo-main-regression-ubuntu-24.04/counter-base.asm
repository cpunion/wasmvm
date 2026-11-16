
/home/runner/work/wasmvm/wasmvm/main-regression/counter-base:	file format elf64-x86-64

Disassembly of section .text:

00000000003a2a00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a2a00:      	push	rbp
  3a2a01:      	mov	rbp, rsp
  3a2a04:      	push	r15
  3a2a06:      	push	r14
  3a2a08:      	push	r13
  3a2a0a:      	push	r12
  3a2a0c:      	push	rbx
  3a2a0d:      	sub	rsp, 0x2a8
  3a2a14:      	mov	rcx, qword ptr [rdi + 0x48]
  3a2a18:      	test	rcx, rcx
  3a2a1b:      	jle	0x3a43aa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19aa>
  3a2a21:      	mov	rbx, rsi
  3a2a24:      	mov	r11, rdi
  3a2a27:      	movzx	r12d, byte ptr [rdi + 0x20]
  3a2a2c:      	mov	r8d, 0xffffffff
  3a2a32:      	lea	r14, [rbp - 0x90]
  3a2a39:      	mov	qword ptr [rbp - 0x40], rdi
  3a2a3d:      	mov	byte ptr [rbp - 0x29], r12b
  3a2a41:      	mov	r10, qword ptr [r11 + 0x40]
  3a2a45:      	shl	rcx, 0x6
  3a2a49:      	mov	byte ptr [r10 + rcx - 0x30], 0x0
  3a2a4f:      	mov	rsi, qword ptr [r10 + rcx - 0x40]
  3a2a54:      	mov	rax, qword ptr [rsi + 0x48]
  3a2a58:      	mov	qword ptr [rbp - 0x58], rax
  3a2a5c:      	mov	r9, qword ptr [rax + 0x98]
  3a2a63:      	mov	rdi, qword ptr [r10 + rcx - 0x28]
  3a2a68:      	mov	rdx, qword ptr [r11 + 0x88]
  3a2a6f:      	lea	rax, [r9 + rdi]
  3a2a73:      	cmp	rax, rdx
  3a2a76:      	ja	0x3a4570 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b70>
  3a2a7c:      	mov	qword ptr [rbp - 0x68], r14
  3a2a80:      	cmp	rdi, rax
  3a2a83:      	ja	0x3a4582 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b82>
  3a2a89:      	mov	r14, qword ptr [rbp - 0x58]
  3a2a8d:      	mov	r13, qword ptr [r14 + 0xb8]
  3a2a94:      	mov	r15d, dword ptr [r10 + rcx - 0x38]
  3a2a99:      	sub	rdx, rdi
  3a2a9c:      	shl	rdi, 0x4
  3a2aa0:      	test	rdx, rdx
  3a2aa3:      	mov	eax, 0x0
  3a2aa8:      	cmovle	rdi, rax
  3a2aac:      	cmp	r13, r15
  3a2aaf:      	jbe	0x3a4540 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b40>
  3a2ab5:      	add	r10, rcx
  3a2ab8:      	mov	qword ptr [rbp - 0x70], r10
  3a2abc:      	mov	qword ptr [rbp - 0xa8], rsi
  3a2ac3:      	mov	rax, qword ptr [rsi + 0x40]
  3a2ac7:      	mov	qword ptr [rbp - 0x60], rax
  3a2acb:      	mov	rsi, qword ptr [r14 + 0xb0]
  3a2ad2:      	add	rdi, qword ptr [r11 + 0x78]
  3a2ad6:      	mov	r14d, r15d
  3a2ad9:      	mov	qword ptr [rbp - 0x38], r9
  3a2add:      	mov	r10, rdi
  3a2ae0:      	mov	qword ptr [rbp - 0x50], rdi
  3a2ae4:      	mov	qword ptr [rbp - 0xb0], r13
  3a2aeb:      	mov	qword ptr [rbp - 0x48], rsi
  3a2aef:      	jmp	0x3a2b21 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x121>
  3a2af1:      	nop	word ptr cs:[rax + rax]
  3a2b00:      	inc	r14d
  3a2b03:      	mov	r13d, r14d
  3a2b06:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a2b0b:      	mov	r15d, r13d
  3a2b0e:      	mov	r14d, r13d
  3a2b11:      	mov	r13, qword ptr [rbp - 0xb0]
  3a2b18:      	cmp	r13, r15
  3a2b1b:      	jbe	0x3a4540 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b40>
  3a2b21:      	shl	r15, 0x5
  3a2b25:      	add	r15, rsi
  3a2b28:      	test	r12b, 0x1
  3a2b2c:      	je	0x3a2b8e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18e>
  3a2b2e:      	mov	rax, qword ptr [rbp - 0x58]
  3a2b32:      	mov	rsi, qword ptr [rax + 0x80]
  3a2b39:      	mov	edi, dword ptr [r15 + 0x4]
  3a2b3d:      	cmp	rsi, rdi
  3a2b40:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a2b46:      	mov	r13, qword ptr [rax + 0x78]
  3a2b4a:      	imul	r12, rdi, 0xe0
  3a2b51:      	mov	eax, dword ptr [r13 + r12]
  3a2b56:      	cmp	eax, 0x5
  3a2b59:      	setne	cl
  3a2b5c:      	cmp	eax, 0xb
  3a2b5f:      	setne	al
  3a2b62:      	and	al, cl
  3a2b64:      	movzx	ecx, al
  3a2b67:      	mov	rax, qword ptr [r11 + 0x150]
  3a2b6e:      	sub	rax, rcx
  3a2b71:      	jb	0x3a43c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19c8>
  3a2b77:      	mov	qword ptr [r11 + 0x150], rax
  3a2b7e:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a2b83:      	mov	r13, qword ptr [rbp - 0xb0]
  3a2b8a:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2b8e:      	inc	rbx
  3a2b91:      	test	bl, 0x3f
  3a2b94:      	jne	0x3a2c14 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x214>
  3a2b96:      	mov	rax, qword ptr [r11 + 0x160]
  3a2b9d:      	test	rax, rax
  3a2ba0:      	je	0x3a2bb5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5>
  3a2ba2:      	test	r11, r11
  3a2ba5:      	je	0x3a2ca7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2a7>
  3a2bab:      	mov	eax, dword ptr [rax]
  3a2bad:      	test	eax, eax
  3a2baf:      	je	0x3a3138 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x738>
  3a2bb5:      	mov	rax, qword ptr [r11 + 0x8]
  3a2bb9:      	test	rax, rax
  3a2bbc:      	je	0x3a463e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3e>
  3a2bc2:      	mov	rdi, qword ptr [r11 + 0x10]
  3a2bc6:      	mov	rcx, qword ptr [rax + 0x8]
  3a2bca:      	test	byte ptr [rcx + 0x17], 0x20
  3a2bce:      	je	0x3a2be7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e7>
  3a2bd0:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a2bd4:      	and	cl, 0x1f
  3a2bd7:      	cmp	cl, 0x16
  3a2bda:      	je	0x3a2be7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e7>
  3a2bdc:      	mov	qword ptr [rbp - 0x90], rdi
  3a2be3:      	mov	rdi, qword ptr [rbp - 0x68]
  3a2be7:      	call	qword ptr [rax + 0x28]
  3a2bea:      	mov	r8d, 0xffffffff
  3a2bf0:      	mov	r9, qword ptr [rbp - 0x38]
  3a2bf4:      	mov	r10, qword ptr [rbp - 0x50]
  3a2bf8:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2bfc:      	test	rax, rax
  3a2bff:      	je	0x3a2ccc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2cc>
  3a2c05:      	cmp	qword ptr [rax + 0x8], 0x0
  3a2c0a:      	mov	r11, qword ptr [rbp - 0x40]
  3a2c0e:      	jne	0x3a4418 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a18>
  3a2c14:      	cmp	dword ptr [r15 + 0x1c], 0x0
  3a2c19:      	je	0x3a2ce0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c1f:      	test	r12b, 0x1
  3a2c23:      	je	0x3a2c40 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2c25:      	mov	edx, dword ptr [r15]
  3a2c28:      	cmp	edx, 0x20
  3a2c2b:      	je	0x3a2ce0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c31:      	cmp	edx, 0x41
  3a2c34:      	je	0x3a2ce0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c3a:      	jmp	0x3a2ce3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e3>
  3a2c3f:      	nop
  3a2c40:      	mov	edx, dword ptr [r15]
  3a2c43:      	cmp	edx, 0x20
  3a2c46:      	je	0x3a2c51 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x251>
  3a2c48:      	cmp	edx, 0x41
  3a2c4b:      	jne	0x3a2ce3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e3>
  3a2c51:      	mov	eax, ebx
  3a2c53:      	not	eax
  3a2c55:      	test	al, 0x3f
  3a2c57:      	je	0x3a2ce0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c5d:      	cmp	edx, 0x41
  3a2c60:      	jne	0x3a2c84 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x284>
  3a2c62:      	mov	edi, dword ptr [r15 + 0x8]
  3a2c66:      	cmp	r9, rdi
  3a2c69:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2c6f:      	mov	eax, dword ptr [r15 + 0x18]
  3a2c73:      	shl	rdi, 0x4
  3a2c77:      	mov	qword ptr [r10 + rdi], rax
  3a2c7b:      	mov	qword ptr [r10 + rdi + 0x8], 0x0
  3a2c84:      	inc	r14d
  3a2c87:      	cmp	r13, r14
  3a2c8a:      	jbe	0x3a4535 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b35>
  3a2c90:      	mov	rax, r14
  3a2c93:      	shl	rax, 0x5
  3a2c97:      	lea	r15, [rsi + rax]
  3a2c9b:      	inc	rbx
  3a2c9e:      	cmp	dword ptr [rsi + rax + 0x1c], 0x0
  3a2ca3:      	jne	0x3a2c40 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2ca5:      	jmp	0x3a2ce0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2ca7:      	mov	edi, 0x1
  3a2cac:      	call	0x5e6230 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a2cb1:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2cb5:      	mov	r10, qword ptr [rbp - 0x50]
  3a2cb9:      	mov	r9, qword ptr [rbp - 0x38]
  3a2cbd:      	mov	r8d, 0xffffffff
  3a2cc3:      	test	rax, rax
  3a2cc6:      	jne	0x3a2c05 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x205>
  3a2ccc:      	mov	r11, qword ptr [rbp - 0x40]
  3a2cd0:      	cmp	dword ptr [r15 + 0x1c], 0x0
  3a2cd5:      	jne	0x3a2c1f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21f>
  3a2cdb:      	nop	dword ptr [rax + rax]
  3a2ce0:      	mov	edx, dword ptr [r15]
  3a2ce3:      	lea	eax, [rdx - 0xe0]
  3a2ce9:      	cmp	eax, 0xffff20
  3a2cee:      	jae	0x3a2f60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x560>
  3a2cf4:      	mov	eax, edx
  3a2cf6:      	shr	eax, 0x10
  3a2cf9:      	cmp	eax, 0xfb
  3a2cfe:      	je	0x3a2ed0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4d0>
  3a2d04:      	cmp	eax, 0xfc
  3a2d09:      	je	0x3a2dae <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3ae>
  3a2d0f:      	cmp	eax, 0xfd
  3a2d14:      	jne	0x3a2f52 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x552>
  3a2d1a:      	mov	edi, dword ptr [r15 + 0xc]
  3a2d1e:      	cmp	rdi, r8
  3a2d21:      	je	0x3a31f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7f0>
  3a2d27:      	cmp	r9, rdi
  3a2d2a:      	mov	r12, qword ptr [rbp - 0x58]
  3a2d2e:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2d34:      	shl	rdi, 0x4
  3a2d38:      	mov	rcx, qword ptr [r10 + rdi]
  3a2d3c:      	mov	r10, qword ptr [r10 + rdi + 0x8]
  3a2d41:      	mov	edi, dword ptr [r15 + 0x10]
  3a2d45:      	mov	esi, 0xffffffff
  3a2d4a:      	cmp	rdi, r8
  3a2d4d:      	je	0x3a320b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x80b>
  3a2d53:      	cmp	r9, rdi
  3a2d56:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2d5c:      	shl	rdi, 0x4
  3a2d60:      	mov	r8, qword ptr [rbp - 0x50]
  3a2d64:      	mov	rax, qword ptr [r8 + rdi]
  3a2d68:      	mov	r8, qword ptr [r8 + rdi + 0x8]
  3a2d6d:      	mov	edi, dword ptr [r15 + 0x14]
  3a2d71:      	cmp	rdi, rsi
  3a2d74:      	je	0x3a321d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x81d>
  3a2d7a:      	cmp	r9, rdi
  3a2d7d:      	jbe	0x3a4556 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a2d83:      	shl	rdi, 0x4
  3a2d87:      	mov	rsi, qword ptr [rbp - 0x50]
  3a2d8b:      	mov	r9, qword ptr [rsi + rdi]
  3a2d8f:      	mov	r11, qword ptr [rsi + rdi + 0x8]
  3a2d94:      	mov	rsi, qword ptr [r12 + 0x80]
  3a2d9c:      	mov	edi, dword ptr [r15 + 0x4]
  3a2da0:      	cmp	rsi, rdi
  3a2da3:      	ja	0x3a3238 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x838>
  3a2da9:      	jmp	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a2dae:      	add	edx, 0xff03ffed
  3a2db4:      	cmp	edx, 0x4
  3a2db7:      	jae	0x3a2f60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x560>
  3a2dbd:      	mov	rax, qword ptr [rbp - 0x58]
  3a2dc1:      	mov	rdx, qword ptr [rax + 0xd8]
  3a2dc8:      	mov	edi, dword ptr [r15 + 0x18]
  3a2dcc:      	mov	eax, dword ptr [r15 + 0x1c]
  3a2dd0:      	lea	ecx, [rdi + rax]
  3a2dd3:      	add	ecx, 0x2
  3a2dd6:      	cmp	rdx, rcx
  3a2dd9:      	jb	0x3a45e4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1be4>
  3a2ddf:      	cmp	edi, ecx
  3a2de1:      	ja	0x3a45d5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd5>
  3a2de7:      	sub	rdx, rdi
  3a2dea:      	lea	r8, [4*rdi]
  3a2df2:      	test	rdx, rdx
  3a2df5:      	mov	eax, 0x0
  3a2dfa:      	cmovle	r8, rax
  3a2dfe:      	cmp	ecx, edi
  3a2e00:      	je	0x3a45cc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bcc>
  3a2e06:      	mov	rax, qword ptr [rbp - 0x58]
  3a2e0a:      	add	r8, qword ptr [rax + 0xc8]
  3a2e11:      	mov	eax, dword ptr [r8]
  3a2e14:      	cmp	r9, rax
  3a2e17:      	jbe	0x3a4594 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b94>
  3a2e1d:      	sub	rcx, rdi
  3a2e20:      	cmp	ecx, 0x1
  3a2e23:      	je	0x3a45bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bbd>
  3a2e29:      	mov	edi, dword ptr [r8 + 0x4]
  3a2e2d:      	cmp	r9, rdi
  3a2e30:      	jbe	0x3a4567 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b67>
  3a2e36:      	shl	rax, 0x4
  3a2e3a:      	mov	r9, qword ptr [r10 + rax]
  3a2e3e:      	shl	rdi, 0x4
  3a2e42:      	mov	rsi, qword ptr [r10 + rdi]
  3a2e46:      	mov	r10d, dword ptr [r15]
  3a2e49:      	lea	eax, [r10 - 0xfc0013]
  3a2e50:      	cmp	eax, 0x2
  3a2e53:      	jae	0x3a3351 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x951>
  3a2e59:      	cmp	ecx, 0x3
  3a2e5c:      	jb	0x3a460e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c0e>
  3a2e62:      	mov	edi, dword ptr [r8 + 0x8]
  3a2e66:      	cmp	qword ptr [rbp - 0x38], rdi
  3a2e6a:      	jbe	0x3a4567 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b67>
  3a2e70:      	cmp	ecx, 0x3
  3a2e73:      	je	0x3a45ff <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bff>
  3a2e79:      	mov	eax, dword ptr [r8 + 0xc]
  3a2e7d:      	cmp	qword ptr [rbp - 0x38], rax
  3a2e81:      	jbe	0x3a45f3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bf3>
  3a2e87:      	shl	rdi, 0x4
  3a2e8b:      	mov	rdx, qword ptr [rbp - 0x50]
  3a2e8f:      	mov	rdi, qword ptr [rdx + rdi]
  3a2e93:      	shl	rax, 0x4
  3a2e97:      	mov	rdx, qword ptr [rdx + rax]
  3a2e9b:      	cmp	r10d, 0xfc0013
  3a2ea2:      	jne	0x3a3484 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa84>
  3a2ea8:      	lea	rax, [rdi + r9]
  3a2eac:      	mov	r10, rdi
  3a2eaf:      	and	r10, r9
  3a2eb2:      	or	rdi, r9
  3a2eb5:      	mov	r9, rax
  3a2eb8:      	not	r9
  3a2ebb:      	and	r9, rdi
  3a2ebe:      	or	r9, r10
  3a2ec1:      	shr	r9, 0x3f
  3a2ec5:      	add	rdx, rsi
  3a2ec8:      	add	rdx, r9
  3a2ecb:      	jmp	0x3a34a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa9>
  3a2ed0:      	add	edx, -0x2
  3a2ed3:      	cmp	dx, 0x21
  3a2ed7:      	jae	0x3a30c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6c9>
  3a2edd:      	lea	rax, [r8 + 0x1c003e10]
  3a2ee4:      	movzx	ecx, dx
  3a2ee7:      	bt	rax, rcx
  3a2eeb:      	jae	0x3a30c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6c9>
  3a2ef1:      	mov	edi, dword ptr [r15 + 0xc]
  3a2ef5:      	mov	eax, 0xffffffff
  3a2efa:      	cmp	rdi, r8
  3a2efd:      	je	0x3a3524 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb24>
  3a2f03:      	cmp	r9, rdi
  3a2f06:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2f0c:      	shl	rdi, 0x4
  3a2f10:      	mov	r8, qword ptr [r10 + rdi]
  3a2f14:      	mov	edi, dword ptr [r15 + 0x10]
  3a2f18:      	cmp	rdi, rax
  3a2f1b:      	je	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb34>
  3a2f21:      	cmp	r9, rdi
  3a2f24:      	jbe	0x3a4556 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a2f2a:      	shl	rdi, 0x4
  3a2f2e:      	mov	rax, qword ptr [r10 + rdi]
  3a2f32:      	mov	r9, qword ptr [r10 + rdi + 0x8]
  3a2f37:      	mov	r10d, 0x0
  3a2f3d:      	mov	r11d, 0x0
  3a2f43:      	cmp	dx, 0xc
  3a2f47:      	je	0x3a354b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb4b>
  3a2f4d:      	jmp	0x3a358b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb8b>
  3a2f52:      	cmp	edx, 0xe7
  3a2f58:      	jb	0x3a4309 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1909>
  3a2f5e:      	nop
  3a2f60:      	mov	edi, dword ptr [r15 + 0xc]
  3a2f64:      	cmp	rdi, r8
  3a2f67:      	je	0x3a2ff0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5f0>
  3a2f6d:      	cmp	r9, rdi
  3a2f70:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2f76:      	shl	rdi, 0x4
  3a2f7a:      	mov	r12, qword ptr [r10 + rdi]
  3a2f7e:      	mov	rdx, qword ptr [r10 + rdi + 0x8]
  3a2f83:      	mov	edi, dword ptr [r15 + 0x10]
  3a2f87:      	cmp	rdi, r8
  3a2f8a:      	je	0x3a2ffe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5fe>
  3a2f8c:      	cmp	r9, rdi
  3a2f8f:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a2f95:      	shl	rdi, 0x4
  3a2f99:      	mov	rax, r10
  3a2f9c:      	mov	r10, qword ptr [r10 + rdi]
  3a2fa0:      	mov	rax, qword ptr [rax + rdi + 0x8]
  3a2fa5:      	mov	r9d, dword ptr [r15]
  3a2fa8:      	lea	ecx, [r9 - 0x45]
  3a2fac:      	cmp	ecx, 0x7f
  3a2faf:      	ja	0x3a300f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x60f>
  3a2fb1:      	ja	0x3a3d64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1364>
  3a2fb7:      	jmp	qword ptr [8*rcx + 0x203aa8]
  3a2fbe:      	mov	edi, r9d
  3a2fc1:      	mov	rsi, r12
  3a2fc4:      	mov	rdx, r10
  3a2fc7:      	call	0x3d0a60 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a2fcc:      	test	dl, dl
  3a2fce:      	jne	0x3a4423 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a23>
  3a2fd4:      	xor	edx, edx
  3a2fd6:      	mov	r11, qword ptr [rbp - 0x40]
  3a2fda:      	mov	r8d, 0xffffffff
  3a2fe0:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2fe4:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a2fe9:      	nop	dword ptr [rax]
  3a2ff0:      	xor	r12d, r12d
  3a2ff3:      	xor	edx, edx
  3a2ff5:      	mov	edi, dword ptr [r15 + 0x10]
  3a2ff9:      	cmp	rdi, r8
  3a2ffc:      	jne	0x3a2f8c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x58c>
  3a2ffe:      	xor	r10d, r10d
  3a3001:      	xor	eax, eax
  3a3003:      	mov	r9d, dword ptr [r15]
  3a3006:      	lea	ecx, [r9 - 0x45]
  3a300a:      	cmp	ecx, 0x7f
  3a300d:      	jbe	0x3a2fb1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5b1>
  3a300f:      	lea	ecx, [r9 - 0x1]
  3a3013:      	cmp	ecx, 0xd3
  3a3019:      	ja	0x3a3106 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x706>
  3a301f:      	jmp	qword ptr [8*rcx + 0x2033c8]
  3a3026:      	mov	rcx, qword ptr [rbp - 0x58]
  3a302a:      	mov	rsi, qword ptr [rcx + 0x80]
  3a3031:      	mov	edi, dword ptr [r15 + 0x4]
  3a3035:      	cmp	rsi, rdi
  3a3038:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a303e:      	imul	rax, rdi, 0xe0
  3a3045:      	add	rax, qword ptr [rcx + 0x78]
  3a3049:      	mov	rcx, qword ptr [rbp - 0x60]
  3a304d:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3053:      	mov	rsi, qword ptr [rcx + 0x78]
  3a3057:      	mov	edi, dword ptr [rax + 0x18]
  3a305a:      	cmp	rsi, rdi
  3a305d:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3063:      	shl	rdi, 0x3
  3a3067:      	add	rdi, qword ptr [rcx + 0x70]
  3a306b:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3071:      	mov	rcx, qword ptr [rdi]
  3a3074:      	test	rcx, rcx
  3a3077:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a307d:      	mov	rsi, qword ptr [rcx + 0x10]
  3a3081:      	mov	ecx, dword ptr [rax + 0x14]
  3a3084:      	mov	r8d, dword ptr [rax + 0x1c]
  3a3088:      	shl	r8, 0x20
  3a308c:      	or	r8, rcx
  3a308f:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a3094:      	je	0x3a30a1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6a1>
  3a3096:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a309b:      	je	0x3a344b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa4b>
  3a30a1:      	xorps	xmm0, xmm0
  3a30a4:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a30ab:      	movups	xmmword ptr [rsp], xmm0
  3a30af:      	mov	r13, qword ptr [rbp - 0x68]
  3a30b3:      	mov	rdi, r13
  3a30b6:      	mov	edx, r9d
  3a30b9:      	mov	rcx, r12
  3a30bc:      	xor	r9d, r9d
  3a30bf:      	call	0x3c6b80 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a30c4:      	jmp	0x3a346e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa6e>
  3a30c9:      	mov	r15, qword ptr [rbp - 0x70]
  3a30cd:      	mov	dword ptr [r15 - 0x38], r14d
  3a30d1:      	mov	rdi, r11
  3a30d4:      	call	0x3a48b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a30d9:      	mov	rsi, qword ptr [rbp - 0x48]
  3a30dd:      	mov	r10, qword ptr [rbp - 0x50]
  3a30e1:      	mov	r9, qword ptr [rbp - 0x38]
  3a30e5:      	mov	r8d, 0xffffffff
  3a30eb:      	mov	r11, qword ptr [rbp - 0x40]
  3a30ef:      	test	al, al
  3a30f1:      	jne	0x3a44ff <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a30f7:      	inc	r14d
  3a30fa:      	mov	dword ptr [r15 - 0x38], r14d
  3a30fe:      	mov	r13d, r14d
  3a3101:      	jmp	0x3a2b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10b>
  3a3106:      	lea	ecx, [r9 - 0xfc0000]
  3a310d:      	cmp	ecx, 0x8
  3a3110:      	jae	0x3a3304 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x904>
  3a3116:      	mov	edi, r9d
  3a3119:      	mov	rsi, r12
  3a311c:      	mov	rdx, r10
  3a311f:      	call	0x3d0a60 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a3124:      	mov	r8d, 0xffffffff
  3a312a:      	mov	r11, qword ptr [rbp - 0x40]
  3a312e:      	mov	r13d, edx
  3a3131:      	xor	edx, edx
  3a3133:      	jmp	0x3a3dbd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13bd>
  3a3138:      	xor	eax, eax
  3a313a:      	xor	edx, edx
  3a313c:      	test	rax, rax
  3a313f:      	jne	0x3a2c05 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x205>
  3a3145:      	jmp	0x3a2ccc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2cc>
  3a314a:      	mov	rdx, qword ptr [rbp - 0x58]
  3a314e:      	mov	rsi, qword ptr [rdx + 0x80]
  3a3155:      	mov	edi, dword ptr [r15 + 0x4]
  3a3159:      	cmp	rsi, rdi
  3a315c:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3162:      	imul	rcx, rdi, 0xe0
  3a3169:      	add	rcx, qword ptr [rdx + 0x78]
  3a316d:      	mov	rdx, qword ptr [rbp - 0x60]
  3a3171:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3177:      	mov	rsi, qword ptr [rdx + 0x78]
  3a317b:      	mov	edi, dword ptr [rcx + 0x18]
  3a317e:      	cmp	rsi, rdi
  3a3181:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3187:      	shl	rdi, 0x3
  3a318b:      	add	rdi, qword ptr [rdx + 0x70]
  3a318f:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3195:      	mov	rdx, qword ptr [rdi]
  3a3198:      	test	rdx, rdx
  3a319b:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a31a1:      	mov	rdi, qword ptr [rdx + 0x10]
  3a31a5:      	mov	edx, dword ptr [rcx + 0x14]
  3a31a8:      	mov	ecx, dword ptr [rcx + 0x1c]
  3a31ab:      	shl	rcx, 0x20
  3a31af:      	or	rcx, rdx
  3a31b2:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a31b7:      	je	0x3a31c4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7c4>
  3a31b9:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a31be:      	je	0x3a369e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc9e>
  3a31c4:      	mov	qword ptr [rbp - 0x90], r10
  3a31cb:      	mov	qword ptr [rbp - 0x88], rax
  3a31d2:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a31d9:      	movups	xmmword ptr [rsp], xmm0
  3a31dd:      	mov	esi, r9d
  3a31e0:      	mov	rdx, r12
  3a31e3:      	xor	r8d, r8d
  3a31e6:      	call	0x3c8070 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a31eb:      	jmp	0x3a3dac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13ac>
  3a31f0:      	xor	r10d, r10d
  3a31f3:      	xor	ecx, ecx
  3a31f5:      	mov	r12, qword ptr [rbp - 0x58]
  3a31f9:      	mov	edi, dword ptr [r15 + 0x10]
  3a31fd:      	mov	esi, 0xffffffff
  3a3202:      	cmp	rdi, r8
  3a3205:      	jne	0x3a2d53 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x353>
  3a320b:      	xor	eax, eax
  3a320d:      	xor	r8d, r8d
  3a3210:      	mov	edi, dword ptr [r15 + 0x14]
  3a3214:      	cmp	rdi, rsi
  3a3217:      	jne	0x3a2d7a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x37a>
  3a321d:      	xor	r9d, r9d
  3a3220:      	xor	r11d, r11d
  3a3223:      	mov	rsi, qword ptr [r12 + 0x80]
  3a322b:      	mov	edi, dword ptr [r15 + 0x4]
  3a322f:      	cmp	rsi, rdi
  3a3232:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3238:      	imul	rdi, rdi, 0xe0
  3a323f:      	add	rdi, qword ptr [r12 + 0x78]
  3a3244:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a324a:      	movzx	esi, dx
  3a324d:      	cmp	esi, 0xb
  3a3250:      	setb	r12b
  3a3254:      	mov	esi, edx
  3a3256:      	and	esi, 0xfffc
  3a325c:      	cmp	esi, 0x54
  3a325f:      	sete	r13b
  3a3263:      	or	r13b, r12b
  3a3266:      	jne	0x3a327b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x87b>
  3a3268:      	movzx	r12d, dx
  3a326c:      	lea	r13d, [r12 - 0x5c]
  3a3271:      	cmp	r13d, 0x2
  3a3275:      	jae	0x3a3380 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x980>
  3a327b:      	mov	r10, qword ptr [rbp - 0x60]
  3a327f:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3283:      	mov	r9d, dword ptr [rdi + 0x18]
  3a3287:      	cmp	rsi, r9
  3a328a:      	jbe	0x3a455f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5f>
  3a3290:      	shl	r9, 0x3
  3a3294:      	add	r9, qword ptr [r10 + 0x70]
  3a3298:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a329e:      	mov	rsi, qword ptr [r9]
  3a32a1:      	test	rsi, rsi
  3a32a4:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a32aa:      	mov	rsi, qword ptr [rsi + 0x10]
  3a32ae:      	mov	r9d, dword ptr [rdi + 0x14]
  3a32b2:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a32b6:      	shl	r10, 0x20
  3a32ba:      	or	r10, r9
  3a32bd:      	movzx	edi, byte ptr [rdi + 0x40]
  3a32c1:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a32c6:      	je	0x3a32d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8d3>
  3a32c8:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a32cd:      	je	0x3a3419 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa19>
  3a32d3:      	mov	qword ptr [rbp - 0xa0], rax
  3a32da:      	mov	qword ptr [rbp - 0x98], r8
  3a32e1:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a32e8:      	movups	xmmword ptr [rsp], xmm0
  3a32ec:      	movzx	r9d, dil
  3a32f0:      	mov	r13, qword ptr [rbp - 0x68]
  3a32f4:      	mov	rdi, r13
  3a32f7:      	mov	r8, r10
  3a32fa:      	call	0x3c6b80 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a32ff:      	jmp	0x3a346e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa6e>
  3a3304:      	lea	ecx, [r9 - 0x1000001]
  3a330b:      	cmp	ecx, 0x8
  3a330e:      	ja	0x3a3f6e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x156e>
  3a3314:      	mov	edx, 0x117
  3a3319:      	bt	edx, ecx
  3a331c:      	jae	0x3a3f5c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x155c>
  3a3322:      	movzx	eax, r9b
  3a3326:      	xor	edi, edi
  3a3328:      	add	eax, -0x2
  3a332b:      	cmp	eax, 0x7
  3a332e:      	ja	0x3a3dda <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13da>
  3a3334:      	mov	r9, qword ptr [rbp - 0x38]
  3a3338:      	mov	r10, qword ptr [rbp - 0x50]
  3a333c:      	jmp	qword ptr [8*rax + 0x203a68]
  3a3343:      	xor	edi, edi
  3a3345:      	test	r12d, r12d
  3a3348:      	sete	dil
  3a334c:      	jmp	0x3a3de2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13e2>
  3a3351:      	mov	rax, rsi
  3a3354:      	mul	r9
  3a3357:      	cmp	dword ptr [r15], 0xfc0015
  3a335e:      	jne	0x3a34a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa9>
  3a3364:      	mov	rdi, r9
  3a3367:      	sar	rdi, 0x3f
  3a336b:      	and	rdi, rsi
  3a336e:      	sar	rsi, 0x3f
  3a3372:      	and	rsi, r9
  3a3375:      	add	rsi, rdi
  3a3378:      	sub	rdx, rsi
  3a337b:      	jmp	0x3a34a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa9>
  3a3380:      	cmp	r12d, 0xb
  3a3384:      	je	0x3a338f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x98f>
  3a3386:      	cmp	esi, 0x58
  3a3389:      	jne	0x3a3703 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd03>
  3a338f:      	mov	r10, qword ptr [rbp - 0x60]
  3a3393:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3397:      	mov	r9d, dword ptr [rdi + 0x18]
  3a339b:      	cmp	rsi, r9
  3a339e:      	jbe	0x3a455f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5f>
  3a33a4:      	shl	r9, 0x3
  3a33a8:      	add	r9, qword ptr [r10 + 0x70]
  3a33ac:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a33b2:      	mov	rsi, qword ptr [r9]
  3a33b5:      	test	rsi, rsi
  3a33b8:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a33be:      	mov	rsi, qword ptr [rsi + 0x10]
  3a33c2:      	mov	r10d, dword ptr [rdi + 0x14]
  3a33c6:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a33ca:      	shl	r9, 0x20
  3a33ce:      	or	r9, r10
  3a33d1:      	movzx	edi, byte ptr [rdi + 0x40]
  3a33d5:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a33da:      	je	0x3a33e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9e7>
  3a33dc:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a33e1:      	je	0x3a3d7f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x137f>
  3a33e7:      	mov	qword ptr [rbp - 0x90], rax
  3a33ee:      	mov	qword ptr [rbp - 0x88], r8
  3a33f5:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a33fc:      	movups	xmmword ptr [rsp], xmm0
  3a3400:      	movzx	r8d, dil
  3a3404:      	mov	rdi, rsi
  3a3407:      	mov	esi, edx
  3a3409:      	mov	rdx, rcx
  3a340c:      	mov	rcx, r9
  3a340f:      	call	0x3c8070 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a3414:      	jmp	0x3a3dac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13ac>
  3a3419:      	mov	qword ptr [rbp - 0xa0], rax
  3a3420:      	mov	qword ptr [rbp - 0x98], r8
  3a3427:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a342e:      	movups	xmmword ptr [rsp], xmm0
  3a3432:      	movzx	r9d, dil
  3a3436:      	mov	r13, qword ptr [rbp - 0x68]
  3a343a:      	mov	rdi, r13
  3a343d:      	mov	r8, r10
  3a3440:      	jmp	0x3a3469 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa69>
  3a3442:      	mov	rax, qword ptr [r15 + 0x18]
  3a3446:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a344b:      	xorps	xmm0, xmm0
  3a344e:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a3455:      	movups	xmmword ptr [rsp], xmm0
  3a3459:      	mov	r13, qword ptr [rbp - 0x68]
  3a345d:      	mov	rdi, r13
  3a3460:      	mov	edx, r9d
  3a3463:      	mov	rcx, r12
  3a3466:      	xor	r9d, r9d
  3a3469:      	call	0x3c6cc0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a346e:      	mov	rax, qword ptr [r13]
  3a3472:      	mov	rdx, qword ptr [r13 + 0x8]
  3a3476:      	mov	qword ptr [rbp - 0x68], r13
  3a347a:      	movzx	r13d, byte ptr [r13 + 0x10]
  3a347f:      	jmp	0x3a3db3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13b3>
  3a3484:      	mov	rax, r9
  3a3487:      	sub	rax, rdi
  3a348a:      	not	r9
  3a348d:      	mov	r10, rdi
  3a3490:      	and	r10, r9
  3a3493:      	xor	r9, rdi
  3a3496:      	and	r9, rax
  3a3499:      	or	r9, r10
  3a349c:      	shr	r9, 0x3f
  3a34a0:      	add	rdx, r9
  3a34a3:      	sub	rsi, rdx
  3a34a6:      	mov	rdx, rsi
  3a34a9:      	mov	edi, dword ptr [r15 + 0x1c]
  3a34ad:      	cmp	ecx, edi
  3a34af:      	mov	r9, qword ptr [rbp - 0x38]
  3a34b3:      	jbe	0x3a459f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b9f>
  3a34b9:      	lea	esi, [rdi + 0x1]
  3a34bc:      	cmp	ecx, esi
  3a34be:      	mov	r10, qword ptr [rbp - 0x70]
  3a34c2:      	jbe	0x3a45a7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ba7>
  3a34c8:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a34cc:      	cmp	r9, rcx
  3a34cf:      	jbe	0x3a45b2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bb2>
  3a34d5:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a34d9:      	shl	rcx, 0x4
  3a34dd:      	mov	rsi, qword ptr [rbp - 0x50]
  3a34e1:      	mov	qword ptr [rsi + rcx], rax
  3a34e5:      	mov	qword ptr [rsi + rcx + 0x8], 0x0
  3a34ee:      	cmp	r9, rdi
  3a34f1:      	jbe	0x3a45b5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bb5>
  3a34f7:      	shl	rdi, 0x4
  3a34fb:      	mov	qword ptr [rsi + rdi], rdx
  3a34ff:      	mov	qword ptr [rsi + rdi + 0x8], 0x0
  3a3508:      	inc	r14d
  3a350b:      	mov	dword ptr [r10 - 0x38], r14d
  3a350f:      	mov	r13d, r14d
  3a3512:      	mov	r8d, 0xffffffff
  3a3518:      	mov	r10, rsi
  3a351b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a351f:      	jmp	0x3a2b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10b>
  3a3524:      	xor	r8d, r8d
  3a3527:      	mov	edi, dword ptr [r15 + 0x10]
  3a352b:      	cmp	rdi, rax
  3a352e:      	jne	0x3a2f21 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x521>
  3a3534:      	xor	eax, eax
  3a3536:      	xor	r9d, r9d
  3a3539:      	mov	r10d, 0x0
  3a353f:      	mov	r11d, 0x0
  3a3545:      	cmp	dx, 0xc
  3a3549:      	jne	0x3a358b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb8b>
  3a354b:      	mov	rcx, qword ptr [rbp - 0x58]
  3a354f:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a3556:      	mov	edi, dword ptr [r15 + 0x18]
  3a355a:      	add	edi, 0x2
  3a355d:      	cmp	rsi, rdi
  3a3560:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3566:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a356d:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a3570:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3574:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3578:      	jbe	0x3a4556 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a357e:      	shl	rdi, 0x4
  3a3582:      	mov	r10, qword ptr [rcx + rdi]
  3a3586:      	mov	r11, qword ptr [rcx + rdi + 0x8]
  3a358b:      	cmp	qword ptr [rbp - 0x40], 0x0
  3a3590:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3596:      	mov	rdx, qword ptr [rbp - 0x58]
  3a359a:      	mov	rsi, qword ptr [rdx + 0x80]
  3a35a1:      	mov	edi, dword ptr [r15 + 0x4]
  3a35a5:      	cmp	rsi, rdi
  3a35a8:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a35ae:      	imul	rcx, rdi, 0xe0
  3a35b5:      	add	rcx, qword ptr [rdx + 0x78]
  3a35b9:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a35bf:      	mov	rdx, qword ptr [rbp - 0x40]
  3a35c3:      	mov	rsi, qword ptr [rdx]
  3a35c6:      	mov	qword ptr [rbp - 0xd0], rax
  3a35cd:      	mov	qword ptr [rbp - 0xc8], r9
  3a35d4:      	mov	qword ptr [rbp - 0xc0], r10
  3a35db:      	mov	qword ptr [rbp - 0xb8], r11
  3a35e2:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a35e9:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a35ee:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a35f5:      	movups	xmmword ptr [rsp], xmm0
  3a35f9:      	lea	rdi, [rbp - 0x150]
  3a3600:      	mov	rdx, qword ptr [rbp - 0x60]
  3a3604:      	call	0x3b6980 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a3609:      	movaps	xmm0, xmmword ptr [rbp - 0x150]
  3a3610:      	mov	r13, qword ptr [rbp - 0x140]
  3a3617:      	test	r13, r13
  3a361a:      	je	0x3a36c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcc5>
  3a3620:      	cmp	qword ptr [r13 + 0x8], 0x0
  3a3625:      	mov	r8d, 0xffffffff
  3a362b:      	mov	r9, qword ptr [rbp - 0x38]
  3a362f:      	mov	r10, qword ptr [rbp - 0x50]
  3a3633:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3637:      	je	0x3a36d7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcd7>
  3a363d:      	jmp	0x3a4428 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a28>
  3a3642:      	mov	edi, dword ptr [r15 + 0x18]
  3a3646:      	cmp	qword ptr [rbp - 0x38], rdi
  3a364a:      	jbe	0x3a4556 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a3650:      	shl	rdi, 0x4
  3a3654:      	mov	rax, qword ptr [rbp - 0x50]
  3a3658:      	mov	qword ptr [rax + rdi], r12
  3a365c:      	mov	qword ptr [rax + rdi + 0x8], rdx
  3a3661:      	mov	rax, r12
  3a3664:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3669:      	mov	edi, dword ptr [r15 + 0x14]
  3a366d:      	cmp	rdi, r8
  3a3670:      	je	0x3a3dcf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13cf>
  3a3676:      	cmp	qword ptr [rbp - 0x38], rdi
  3a367a:      	jbe	0x3a4556 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a3680:      	shl	rdi, 0x4
  3a3684:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3688:      	mov	rcx, qword ptr [rcx + rdi]
  3a368c:      	test	ecx, ecx
  3a368e:      	cmove	rdx, rax
  3a3692:      	cmove	r12, r10
  3a3696:      	mov	rax, r12
  3a3699:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a369e:      	mov	qword ptr [rbp - 0x90], r10
  3a36a5:      	mov	qword ptr [rbp - 0x88], rax
  3a36ac:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a36b3:      	movups	xmmword ptr [rsp], xmm0
  3a36b7:      	mov	esi, r9d
  3a36ba:      	mov	rdx, r12
  3a36bd:      	xor	r8d, r8d
  3a36c0:      	jmp	0x3a3da7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13a7>
  3a36c5:      	mov	r8d, 0xffffffff
  3a36cb:      	mov	r9, qword ptr [rbp - 0x38]
  3a36cf:      	mov	r10, qword ptr [rbp - 0x50]
  3a36d3:      	mov	rsi, qword ptr [rbp - 0x48]
  3a36d7:      	mov	edi, dword ptr [r15 + 0x8]
  3a36db:      	cmp	rdi, r8
  3a36de:      	mov	r11, qword ptr [rbp - 0x40]
  3a36e2:      	mov	r15, qword ptr [rbp - 0x70]
  3a36e6:      	je	0x3a30f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6f7>
  3a36ec:      	cmp	r9, rdi
  3a36ef:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a36f5:      	shl	rdi, 0x4
  3a36f9:      	movups	xmmword ptr [r10 + rdi], xmm0
  3a36fe:      	jmp	0x3a30f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6f7>
  3a3703:      	mov	qword ptr [rsp], r11
  3a3707:      	mov	rsi, rcx
  3a370a:      	mov	rdx, r10
  3a370d:      	mov	rcx, rax
  3a3710:      	call	0x3cec20 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a3715:      	jmp	0x3a2fd6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d6>
  3a371a:      	mov	edi, dword ptr [r15 + 0x1c]
  3a371e:      	dec	edi
  3a3720:      	cmp	edi, r12d
  3a3723:      	cmovae	edi, r12d
  3a3727:      	jmp	0x3a3de2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13e2>
  3a372c:      	xor	edi, edi
  3a372e:      	test	r12, r12
  3a3731:      	setne	dil
  3a3735:      	jmp	0x3a3de2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13e2>
  3a373a:      	xor	eax, eax
  3a373c:      	cmp	r12, r10
  3a373f:      	sete	al
  3a3742:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3747:      	mov	rax, qword ptr [rbp - 0x60]
  3a374b:      	mov	rsi, qword ptr [rax + 0x78]
  3a374f:      	mov	edi, dword ptr [r15 + 0x18]
  3a3753:      	cmp	rsi, rdi
  3a3756:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a375c:      	shl	rdi, 0x3
  3a3760:      	add	rdi, qword ptr [rax + 0x70]
  3a3764:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a376a:      	test	r11, r11
  3a376d:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3773:      	mov	r13, qword ptr [rdi]
  3a3776:      	test	r13, r13
  3a3779:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a377f:      	mov	rdi, qword ptr [r11]
  3a3782:      	mov	rax, qword ptr [r13 + 0x10]
  3a3786:      	movzx	eax, byte ptr [rax + 0x31]
  3a378a:      	mov	byte ptr [rbp - 0x71], al
  3a378d:      	mov	edx, r12d
  3a3790:      	test	al, 0x1
  3a3792:      	cmovne	rdx, r12
  3a3796:      	mov	rsi, r13
  3a3799:      	call	0x3c0cf0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a379e:      	test	dl, 0x1
  3a37a1:      	jne	0x3a2fd4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d4>
  3a37a7:      	mov	rax, qword ptr [r13 + 0x10]
  3a37ab:      	movzx	eax, byte ptr [rax + 0x31]
  3a37af:      	mov	byte ptr [rbp - 0x72], al
  3a37b2:      	and	al, 0x1
  3a37b4:      	neg	al
  3a37b6:      	mov	eax, 0x0
  3a37bb:      	sbb	rax, rax
  3a37be:      	mov	r8d, 0xffffffff
  3a37c4:      	or	rax, r8
  3a37c7:      	xor	edx, edx
  3a37c9:      	mov	r11, qword ptr [rbp - 0x40]
  3a37cd:      	mov	rsi, qword ptr [rbp - 0x48]
  3a37d1:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a37d6:      	mov	rax, qword ptr [rbp - 0x60]
  3a37da:      	mov	rsi, qword ptr [rax + 0x60]
  3a37de:      	mov	edi, dword ptr [r15 + 0x18]
  3a37e2:      	cmp	rsi, rdi
  3a37e5:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a37eb:      	mov	rax, qword ptr [rax + 0x58]
  3a37ef:      	lea	rcx, [rax + 8*rdi]
  3a37f3:      	test	rcx, rcx
  3a37f6:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a37fc:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3800:      	test	rax, rax
  3a3803:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3809:      	cmp	byte ptr [rax + 0x29], 0x0
  3a380d:      	mov	ecx, r12d
  3a3810:      	cmovne	rcx, r12
  3a3814:      	cmp	rcx, qword ptr [rax + 0x40]
  3a3818:      	jae	0x3a447f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7f>
  3a381e:      	mov	rax, qword ptr [rax + 0x38]
  3a3822:      	mov	rax, qword ptr [rax + 8*rcx]
  3a3826:      	jmp	0x3a42cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cd>
  3a382b:      	mov	rax, qword ptr [rbp - 0x60]
  3a382f:      	mov	rsi, qword ptr [rax + 0x48]
  3a3833:      	mov	edi, dword ptr [r15 + 0x18]
  3a3837:      	cmp	rsi, rdi
  3a383a:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3840:      	mov	rax, qword ptr [rax + 0x40]
  3a3844:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3848:      	jmp	0x3a42cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cd>
  3a384d:      	mov	rax, r12
  3a3850:      	test	r12, r12
  3a3853:      	jne	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3859:      	jmp	0x3a4489 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a89>
  3a385e:      	mov	rax, qword ptr [rbp - 0x60]
  3a3862:      	mov	rsi, qword ptr [rax + 0x90]
  3a3869:      	mov	edi, dword ptr [r15 + 0x18]
  3a386d:      	cmp	rsi, rdi
  3a3870:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3876:      	mov	rax, qword ptr [rax + 0x88]
  3a387d:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3881:      	mov	qword ptr [rax + 0x20], r12
  3a3885:      	mov	qword ptr [rax + 0x28], rdx
  3a3889:      	jmp	0x3a3912 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf12>
  3a388e:      	mov	rax, qword ptr [rbp - 0x60]
  3a3892:      	mov	rsi, qword ptr [rax + 0x90]
  3a3899:      	mov	edi, dword ptr [r15 + 0x18]
  3a389d:      	cmp	rsi, rdi
  3a38a0:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a38a6:      	mov	rax, qword ptr [rax + 0x88]
  3a38ad:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a38b1:      	mov	rax, qword ptr [rcx + 0x20]
  3a38b5:      	mov	rdx, qword ptr [rcx + 0x28]
  3a38b9:      	mov	rsi, qword ptr [rbp - 0x48]
  3a38bd:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a38c2:      	mov	rax, qword ptr [rbp - 0x60]
  3a38c6:      	mov	rsi, qword ptr [rax + 0x60]
  3a38ca:      	mov	edi, dword ptr [r15 + 0x18]
  3a38ce:      	cmp	rsi, rdi
  3a38d1:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a38d7:      	mov	rax, qword ptr [rax + 0x58]
  3a38db:      	lea	rcx, [rax + 8*rdi]
  3a38df:      	test	rcx, rcx
  3a38e2:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a38e8:      	mov	rax, qword ptr [rax + 8*rdi]
  3a38ec:      	test	rax, rax
  3a38ef:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a38f5:      	cmp	byte ptr [rax + 0x29], 0x0
  3a38f9:      	mov	ecx, r12d
  3a38fc:      	cmovne	rcx, r12
  3a3900:      	cmp	rcx, qword ptr [rax + 0x40]
  3a3904:      	jae	0x3a447f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7f>
  3a390a:      	mov	rax, qword ptr [rax + 0x38]
  3a390e:      	mov	qword ptr [rax + 8*rcx], r10
  3a3912:      	xor	edx, edx
  3a3914:      	xor	eax, eax
  3a3916:      	mov	rsi, qword ptr [rbp - 0x48]
  3a391a:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a391f:      	mov	edi, dword ptr [r15 + 0x18]
  3a3923:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3927:      	jbe	0x3a4556 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a392d:      	shl	rdi, 0x4
  3a3931:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3935:      	mov	rax, qword ptr [rcx + rdi]
  3a3939:      	mov	rdx, qword ptr [rcx + rdi + 0x8]
  3a393e:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3943:      	xor	eax, eax
  3a3945:      	test	r12, r12
  3a3948:      	sete	al
  3a394b:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3950:      	mov	rax, qword ptr [rbp - 0x60]
  3a3954:      	mov	rsi, qword ptr [rax + 0x78]
  3a3958:      	mov	edi, dword ptr [r15 + 0x18]
  3a395c:      	cmp	rsi, rdi
  3a395f:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3965:      	shl	rdi, 0x3
  3a3969:      	add	rdi, qword ptr [rax + 0x70]
  3a396d:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3973:      	mov	rax, qword ptr [rdi]
  3a3976:      	test	rax, rax
  3a3979:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a397f:      	mov	rdi, qword ptr [rax + 0x10]
  3a3983:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a3988:      	mov	rax, qword ptr [rdi + 0x60]
  3a398c:      	je	0x3a4161 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1761>
  3a3992:      	test	rax, rax
  3a3995:      	jne	0x3a416a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x176a>
  3a399b:      	call	0x3c4780 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a39a0:      	mov	r8d, 0xffffffff
  3a39a6:      	mov	r11, qword ptr [rbp - 0x40]
  3a39aa:      	jmp	0x3a42cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cd>
  3a39af:      	mov	eax, dword ptr [r15 + 0x18]
  3a39b3:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a39b8:      	xor	r10, r12
  3a39bb:      	jmp	0x3a3d1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a39c0:      	mov	eax, r12d
  3a39c3:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a39c8:      	mov	ecx, r10d
  3a39cb:      	shr	r12, cl
  3a39ce:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a39d3:      	mov	ecx, r10d
  3a39d6:      	shr	r12d, cl
  3a39d9:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a39de:      	or	r10, r12
  3a39e1:      	jmp	0x3a3d1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a39e6:      	mov	ecx, r10d
  3a39e9:      	sar	r12d, cl
  3a39ec:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a39f1:      	xor	r10d, r12d
  3a39f4:      	jmp	0x3a3d1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a39f9:      	mov	rax, r12
  3a39fc:      	shr	rax
  3a39ff:      	movabs	rcx, 0x5555555555555555
  3a3a09:      	and	rax, rcx
  3a3a0c:      	and	r12, rcx
  3a3a0f:      	add	r12, rax
  3a3a12:      	mov	rax, r12
  3a3a15:      	shr	rax, 0x2
  3a3a19:      	movabs	rcx, 0x3333333333333333
  3a3a23:      	and	rax, rcx
  3a3a26:      	and	r12, rcx
  3a3a29:      	add	r12, rax
  3a3a2c:      	mov	rax, r12
  3a3a2f:      	shr	rax, 0x4
  3a3a33:      	add	rax, r12
  3a3a36:      	movabs	rcx, 0xf0f0f0f0f0f0f0f
  3a3a40:      	and	rcx, rax
  3a3a43:      	mov	rax, rcx
  3a3a46:      	shr	rax, 0x8
  3a3a4a:      	add	rax, rcx
  3a3a4d:      	mov	rcx, rax
  3a3a50:      	shr	rcx, 0x10
  3a3a54:      	add	rcx, rax
  3a3a57:      	mov	rax, rcx
  3a3a5a:      	shr	rax, 0x20
  3a3a5e:      	add	eax, ecx
  3a3a60:      	and	eax, 0x7f
  3a3a63:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3a68:      	xor	eax, eax
  3a3a6a:      	cmp	r12, r10
  3a3a6d:      	setb	al
  3a3a70:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3a75:      	xor	eax, eax
  3a3a77:      	cmp	r12, r10
  3a3a7a:      	setbe	al
  3a3a7d:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3a82:      	xor	eax, eax
  3a3a84:      	cmp	r12, r10
  3a3a87:      	seta	al
  3a3a8a:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3a8f:      	imul	r10d, r12d
  3a3a93:      	jmp	0x3a3d1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a3a98:      	mov	ecx, r12d
  3a3a9b:      	shr	ecx, 0x10
  3a3a9e:      	xor	eax, eax
  3a3aa0:      	cmp	r12d, 0x10000
  3a3aa7:      	setb	al
  3a3aaa:      	cmovb	ecx, r12d
  3a3aae:      	shl	eax, 0x4
  3a3ab1:      	mov	edx, ecx
  3a3ab3:      	shr	edx, 0x8
  3a3ab6:      	xor	esi, esi
  3a3ab8:      	cmp	ecx, 0x100
  3a3abe:      	setb	sil
  3a3ac2:      	lea	eax, [rax + 8*rsi]
  3a3ac5:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3ac9:      	cmovb	edx, ecx
  3a3acc:      	movzx	ecx, dl
  3a3acf:      	movzx	ecx, byte ptr [rcx + 0x3284c9]
  3a3ad6:      	sub	rax, rcx
  3a3ad9:      	add	rax, 0x8
  3a3add:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3ae2:      	mov	ecx, r10d
  3a3ae5:      	rol	r12d, cl
  3a3ae8:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3aed:      	xor	eax, eax
  3a3aef:      	cmp	r12, r10
  3a3af2:      	setg	al
  3a3af5:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3afa:      	xor	eax, eax
  3a3afc:      	cmp	r12d, r10d
  3a3aff:      	setl	al
  3a3b02:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b07:      	xor	eax, eax
  3a3b09:      	cmp	r12, r10
  3a3b0c:      	setne	al
  3a3b0f:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b14:      	test	r12d, r12d
  3a3b17:      	je	0x3a41b2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17b2>
  3a3b1d:      	tzcnt	eax, r12d
  3a3b22:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b27:      	add	r10d, r12d
  3a3b2a:      	jmp	0x3a3d1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a3b2f:      	xor	eax, eax
  3a3b31:      	cmp	r12d, r10d
  3a3b34:      	setne	al
  3a3b37:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b3c:      	xor	eax, eax
  3a3b3e:      	cmp	r12, r10
  3a3b41:      	setle	al
  3a3b44:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b49:      	mov	ecx, r10d
  3a3b4c:      	ror	r12d, cl
  3a3b4f:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3b54:      	xor	eax, eax
  3a3b56:      	cmp	r12d, r10d
  3a3b59:      	setbe	al
  3a3b5c:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b61:      	xor	eax, eax
  3a3b63:      	cmp	r12d, r10d
  3a3b66:      	setge	al
  3a3b69:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b6e:      	xor	eax, eax
  3a3b70:      	test	r12d, r12d
  3a3b73:      	sete	al
  3a3b76:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b7b:      	xor	eax, eax
  3a3b7d:      	cmp	r12d, r10d
  3a3b80:      	setb	al
  3a3b83:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3b88:      	mov	rdx, r12
  3a3b8b:      	mov	eax, edx
  3a3b8d:      	shr	eax, 0x18
  3a3b90:      	mov	ecx, edx
  3a3b92:      	shr	ecx, 0x10
  3a3b95:      	movzx	ecx, cl
  3a3b98:      	movzx	ecx, byte ptr [rcx + 0x3285c9]
  3a3b9f:      	add	cl, byte ptr [rax + 0x3285c9]
  3a3ba5:      	movzx	eax, dh
  3a3ba8:      	add	cl, byte ptr [rax + 0x3285c9]
  3a3bae:      	movzx	eax, dl
  3a3bb1:      	add	cl, byte ptr [rax + 0x3285c9]
  3a3bb7:      	movzx	eax, cl
  3a3bba:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3bbf:      	add	r10, r12
  3a3bc2:      	jmp	0x3a3d1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a3bc7:      	mov	ecx, r10d
  3a3bca:      	shl	r12d, cl
  3a3bcd:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3bd2:      	xor	eax, eax
  3a3bd4:      	test	r12, r12
  3a3bd7:      	sete	al
  3a3bda:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3bdf:      	and	r10, r12
  3a3be2:      	jmp	0x3a3d1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a3be7:      	sub	r12, r10
  3a3bea:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3bef:      	xor	eax, eax
  3a3bf1:      	cmp	r12d, r10d
  3a3bf4:      	sete	al
  3a3bf7:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3bfc:      	mov	ecx, r10d
  3a3bff:      	rol	r12, cl
  3a3c02:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3c07:      	xor	eax, eax
  3a3c09:      	cmp	r12d, r10d
  3a3c0c:      	setae	al
  3a3c0f:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3c14:      	and	r12d, r10d
  3a3c17:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3c1c:      	xor	eax, eax
  3a3c1e:      	cmp	r12d, r10d
  3a3c21:      	setle	al
  3a3c24:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3c29:      	xor	eax, eax
  3a3c2b:      	cmp	r12d, r10d
  3a3c2e:      	setg	al
  3a3c31:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3c36:      	xor	eax, eax
  3a3c38:      	mov	rcx, r12
  3a3c3b:      	shr	rcx, 0x20
  3a3c3f:      	cmove	rcx, r12
  3a3c43:      	setne	al
  3a3c46:      	shl	eax, 0x5
  3a3c49:      	mov	rdx, rcx
  3a3c4c:      	shr	rdx, 0x10
  3a3c50:      	lea	rsi, [rax + 0x10]
  3a3c54:      	cmp	rcx, 0x10000
  3a3c5b:      	cmovb	rdx, rcx
  3a3c5f:      	cmovb	rsi, rax
  3a3c63:      	mov	eax, edx
  3a3c65:      	shr	eax, 0x8
  3a3c68:      	lea	rcx, [rsi + 0x8]
  3a3c6c:      	cmp	rdx, 0x100
  3a3c73:      	cmovb	rax, rdx
  3a3c77:      	cmovb	rcx, rsi
  3a3c7b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3c7f:      	movzx	eax, al
  3a3c82:      	movzx	edx, byte ptr [rax + 0x3284c9]
  3a3c89:      	add	rdx, rcx
  3a3c8c:      	mov	eax, 0x40
  3a3c91:      	sub	rax, rdx
  3a3c94:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3c99:      	xor	eax, eax
  3a3c9b:      	cmp	r12, r10
  3a3c9e:      	setl	al
  3a3ca1:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3ca6:      	mov	ecx, r10d
  3a3ca9:      	ror	r12, cl
  3a3cac:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3cb1:      	xor	eax, eax
  3a3cb3:      	cmp	r12d, r10d
  3a3cb6:      	seta	al
  3a3cb9:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3cbe:      	mov	eax, r12d
  3a3cc1:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3cc6:      	movsx	rax, r12w
  3a3cca:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3ccf:      	movsxd	rax, r12d
  3a3cd2:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3cd7:      	xor	eax, eax
  3a3cd9:      	cmp	r12, r10
  3a3cdc:      	sete	al
  3a3cdf:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3ce4:      	movsxd	rax, r12d
  3a3ce7:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3cec:      	xor	eax, eax
  3a3cee:      	cmp	r12, r10
  3a3cf1:      	setae	al
  3a3cf4:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3cf9:      	xor	eax, eax
  3a3cfb:      	cmp	r12, r10
  3a3cfe:      	setge	al
  3a3d01:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3d06:      	mov	ecx, r10d
  3a3d09:      	sar	r12, cl
  3a3d0c:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3d0e:      	or	r10d, r12d
  3a3d11:      	jmp	0x3a3d1c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131c>
  3a3d13:      	sub	r12d, r10d
  3a3d16:      	jmp	0x3a3d46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1346>
  3a3d18:      	imul	r10, r12
  3a3d1c:      	xor	edx, edx
  3a3d1e:      	mov	rax, r10
  3a3d21:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3d26:      	movsx	rax, r12w
  3a3d2a:      	mov	eax, eax
  3a3d2c:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3d31:      	mov	eax, 0x40
  3a3d36:      	tzcnt	rax, r12
  3a3d3b:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3d40:      	mov	ecx, r10d
  3a3d43:      	shl	r12, cl
  3a3d46:      	xor	edx, edx
  3a3d48:      	mov	rax, r12
  3a3d4b:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3d50:      	movsx	rax, r12b
  3a3d54:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3d59:      	movsx	rax, r12b
  3a3d5d:      	mov	eax, eax
  3a3d5f:      	jmp	0x3a41c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c0>
  3a3d64:      	movzx	ecx, r9b
  3a3d68:      	add	ecx, -0x61
  3a3d6b:      	cmp	ecx, 0x5e
  3a3d6e:      	ja	0x3a3fe5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15e5>
  3a3d74:      	xor	edx, edx
  3a3d76:      	xor	eax, eax
  3a3d78:      	jmp	qword ptr [8*rcx + 0x203ea8]
  3a3d7f:      	mov	qword ptr [rbp - 0x90], rax
  3a3d86:      	mov	qword ptr [rbp - 0x88], r8
  3a3d8d:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a3d94:      	movups	xmmword ptr [rsp], xmm0
  3a3d98:      	movzx	r8d, dil
  3a3d9c:      	mov	rdi, rsi
  3a3d9f:      	mov	esi, edx
  3a3da1:      	mov	rdx, rcx
  3a3da4:      	mov	rcx, r9
  3a3da7:      	call	0x3c8350 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a3dac:      	mov	r13d, eax
  3a3daf:      	xor	edx, edx
  3a3db1:      	xor	eax, eax
  3a3db3:      	mov	r11, qword ptr [rbp - 0x40]
  3a3db7:      	mov	r8d, 0xffffffff
  3a3dbd:      	test	r13b, r13b
  3a3dc0:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3dc4:      	je	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3dca:      	jmp	0x3a448c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8c>
  3a3dcf:      	mov	rdx, rax
  3a3dd2:      	mov	rax, r10
  3a3dd5:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3dda:      	mov	r9, qword ptr [rbp - 0x38]
  3a3dde:      	mov	r10, qword ptr [rbp - 0x50]
  3a3de2:      	mov	rax, qword ptr [rbp - 0x58]
  3a3de6:      	mov	rsi, qword ptr [rax + 0xe8]
  3a3ded:      	add	edi, dword ptr [r15 + 0x18]
  3a3df1:      	cmp	rsi, rdi
  3a3df4:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a3dfa:      	mov	rax, qword ptr [rax + 0xe0]
  3a3e01:      	lea	rcx, [rdi + 2*rdi]
  3a3e05:      	mov	r13d, dword ptr [rax + 4*rcx]
  3a3e09:      	mov	r15d, dword ptr [rax + 4*rcx + 0x4]
  3a3e0e:      	mov	esi, dword ptr [rax + 4*rcx + 0x8]
  3a3e12:      	cmp	r13d, r14d
  3a3e15:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a3e1a:      	ja	0x3a3ec5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c5>
  3a3e20:      	mov	rax, qword ptr [r11 + 0x160]
  3a3e27:      	test	rax, rax
  3a3e2a:      	je	0x3a3e3b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x143b>
  3a3e2c:      	test	r11, r11
  3a3e2f:      	je	0x3a3e9e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x149e>
  3a3e31:      	mov	eax, dword ptr [rax]
  3a3e33:      	test	eax, eax
  3a3e35:      	je	0x3a41a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17a0>
  3a3e3b:      	mov	rax, qword ptr [r11 + 0x8]
  3a3e3f:      	test	rax, rax
  3a3e42:      	je	0x3a463e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3e>
  3a3e48:      	mov	rdi, qword ptr [r11 + 0x10]
  3a3e4c:      	mov	rcx, qword ptr [rax + 0x8]
  3a3e50:      	test	byte ptr [rcx + 0x17], 0x20
  3a3e54:      	jne	0x3a3e5b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x145b>
  3a3e56:      	mov	r14, rsi
  3a3e59:      	jmp	0x3a3e75 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1475>
  3a3e5b:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a3e5f:      	and	cl, 0x1f
  3a3e62:      	mov	r14, rsi
  3a3e65:      	cmp	cl, 0x16
  3a3e68:      	je	0x3a3e75 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1475>
  3a3e6a:      	mov	qword ptr [rbp - 0x90], rdi
  3a3e71:      	mov	rdi, qword ptr [rbp - 0x68]
  3a3e75:      	call	qword ptr [rax + 0x28]
  3a3e78:      	mov	r8d, 0xffffffff
  3a3e7e:      	mov	r9, qword ptr [rbp - 0x38]
  3a3e82:      	mov	r10, qword ptr [rbp - 0x50]
  3a3e86:      	mov	rsi, r14
  3a3e89:      	test	rax, rax
  3a3e8c:      	je	0x3a3ec1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c1>
  3a3e8e:      	cmp	qword ptr [rax + 0x8], 0x0
  3a3e93:      	mov	r11, qword ptr [rbp - 0x40]
  3a3e97:      	je	0x3a3ec5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c5>
  3a3e99:      	jmp	0x3a4418 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a18>
  3a3e9e:      	mov	edi, 0x1
  3a3ea3:      	mov	r14, rsi
  3a3ea6:      	call	0x5e6230 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a3eab:      	mov	rsi, r14
  3a3eae:      	mov	r10, qword ptr [rbp - 0x50]
  3a3eb2:      	mov	r9, qword ptr [rbp - 0x38]
  3a3eb6:      	mov	r8d, 0xffffffff
  3a3ebc:      	test	rax, rax
  3a3ebf:      	jne	0x3a3e8e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x148e>
  3a3ec1:      	mov	r11, qword ptr [rbp - 0x40]
  3a3ec5:      	mov	rax, qword ptr [rbp - 0x58]
  3a3ec9:      	mov	rax, qword ptr [rax + 0x108]
  3a3ed0:      	lea	edx, [rsi + r15]
  3a3ed4:      	cmp	rax, rdx
  3a3ed7:      	jb	0x3a462c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c2c>
  3a3edd:      	cmp	r15d, edx
  3a3ee0:      	ja	0x3a461d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c1d>
  3a3ee6:      	sub	rax, r15
  3a3ee9:      	lea	rcx, [8*r15]
  3a3ef1:      	test	rax, rax
  3a3ef4:      	mov	eax, 0x0
  3a3ef9:      	cmovle	rcx, rax
  3a3efd:      	test	esi, esi
  3a3eff:      	je	0x3a3f53 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1553>
  3a3f01:      	mov	rax, qword ptr [rbp - 0x58]
  3a3f05:      	mov	rax, qword ptr [rax + 0xf8]
  3a3f0c:      	sub	rdx, r15
  3a3f0f:      	add	rcx, rax
  3a3f12:      	add	rcx, 0x4
  3a3f16:      	xor	esi, esi
  3a3f18:      	nop	dword ptr [rax + rax]
  3a3f20:      	mov	edi, dword ptr [rcx + 8*rsi]
  3a3f23:      	cmp	r9, rdi
  3a3f26:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a3f2c:      	mov	eax, dword ptr [rcx + 8*rsi - 0x4]
  3a3f30:      	cmp	r9, rax
  3a3f33:      	jbe	0x3a454b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b4b>
  3a3f39:      	shl	rdi, 0x4
  3a3f3d:      	shl	rax, 0x4
  3a3f41:      	movups	xmm0, xmmword ptr [r10 + rdi]
  3a3f46:      	movups	xmmword ptr [r10 + rax], xmm0
  3a3f4b:      	inc	rsi
  3a3f4e:      	cmp	rdx, rsi
  3a3f51:      	jne	0x3a3f20 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1520>
  3a3f53:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3f57:      	jmp	0x3a2b0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10b>
  3a3f5c:      	cmp	ecx, 0x3
  3a3f5f:      	je	0x3a434f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x194f>
  3a3f65:      	cmp	ecx, 0x6
  3a3f68:      	je	0x3a4370 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1970>
  3a3f6e:      	lea	ecx, [r9 - 0x12]
  3a3f72:      	cmp	ecx, 0x4
  3a3f75:      	jb	0x3a4325 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1925>
  3a3f7b:      	mov	edi, dword ptr [r15 + 0x14]
  3a3f7f:      	cmp	rdi, r8
  3a3f82:      	je	0x3a3fee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15ee>
  3a3f84:      	cmp	qword ptr [rbp - 0x38], rdi
  3a3f88:      	jbe	0x3a4556 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b56>
  3a3f8e:      	shl	rdi, 0x4
  3a3f92:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3f96:      	movups	xmm0, xmmword ptr [rcx + rdi]
  3a3f9a:      	mov	ecx, r9d
  3a3f9d:      	and	ecx, 0xffff0000
  3a3fa3:      	cmp	ecx, 0xfe0000
  3a3fa9:      	jne	0x3a4002 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1602>
  3a3fab:      	movzx	r9d, r9w
  3a3faf:      	cmp	r9d, 0x3
  3a3fb3:      	jne	0x3a41f2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17f2>
  3a3fb9:      	test	r11, r11
  3a3fbc:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3fc2:      	mov	rax, qword ptr [r11]
  3a3fc5:      	test	rax, rax
  3a3fc8:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a3fce:      	mov	rdi, qword ptr [rax + 0x8]
  3a3fd2:      	call	0x3c3d20 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a3fd7:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3fdb:      	mov	r8d, 0xffffffff
  3a3fe1:      	mov	r11, qword ptr [rbp - 0x40]
  3a3fe5:      	xor	edx, edx
  3a3fe7:      	xor	eax, eax
  3a3fe9:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a3fee:      	xorps	xmm0, xmm0
  3a3ff1:      	mov	ecx, r9d
  3a3ff4:      	and	ecx, 0xffff0000
  3a3ffa:      	cmp	ecx, 0xfe0000
  3a4000:      	je	0x3a3fab <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15ab>
  3a4002:      	add	r9d, 0xff03fff8
  3a4009:      	cmp	r9d, 0xa
  3a400d:      	jae	0x3a464d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c4d>
  3a4013:      	mov	rdx, qword ptr [rbp - 0x58]
  3a4017:      	mov	rsi, qword ptr [rdx + 0x80]
  3a401e:      	mov	edi, dword ptr [r15 + 0x4]
  3a4022:      	cmp	rsi, rdi
  3a4025:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a402b:      	imul	rcx, rdi, 0xe0
  3a4032:      	add	rcx, qword ptr [rdx + 0x78]
  3a4036:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a403c:      	movups	xmm1, xmmword ptr [rcx + 0x98]
  3a4043:      	movups	xmm2, xmmword ptr [rcx + 0x88]
  3a404a:      	movups	xmm3, xmmword ptr [rcx + 0x78]
  3a404e:      	movups	xmm4, xmmword ptr [rcx + 0x68]
  3a4052:      	movups	xmm5, xmmword ptr [rcx + 0x58]
  3a4056:      	movups	xmm6, xmmword ptr [rcx + 0x48]
  3a405a:      	movzx	edx, byte ptr [rcx + 0x40]
  3a405e:      	movups	xmm7, xmmword ptr [rcx + 0x30]
  3a4062:      	movups	xmm8, xmmword ptr [rcx + 0x20]
  3a4067:      	movups	xmm9, xmmword ptr [rcx + 0x10]
  3a406c:      	mov	rsi, qword ptr [rcx + 0x8]
  3a4070:      	mov	ecx, dword ptr [rcx]
  3a4072:      	mov	dword ptr [rbp - 0x208], ecx
  3a4078:      	mov	qword ptr [rbp - 0x200], rsi
  3a407f:      	movups	xmmword ptr [rbp - 0x1f8], xmm9
  3a4087:      	movups	xmmword ptr [rbp - 0x1e8], xmm8
  3a408f:      	movups	xmmword ptr [rbp - 0x1d8], xmm7
  3a4096:      	mov	byte ptr [rbp - 0x1c8], dl
  3a409c:      	movups	xmmword ptr [rbp - 0x1c0], xmm6
  3a40a3:      	movups	xmmword ptr [rbp - 0x1b0], xmm5
  3a40aa:      	movups	xmmword ptr [rbp - 0x1a0], xmm4
  3a40b1:      	movups	xmmword ptr [rbp - 0x190], xmm3
  3a40b8:      	movups	xmmword ptr [rbp - 0x180], xmm2
  3a40bf:      	movups	xmmword ptr [rbp - 0x170], xmm1
  3a40c6:      	mov	qword ptr [rbp - 0xe0], r10
  3a40cd:      	mov	qword ptr [rbp - 0xd8], rax
  3a40d4:      	movaps	xmmword ptr [rbp - 0x160], xmm0
  3a40db:      	mov	ecx, 0x15
  3a40e0:      	mov	rdi, rsp
  3a40e3:      	lea	rsi, [rbp - 0x208]
  3a40ea:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a40ed:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a40f4:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a40fc:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a4103:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a410b:      	lea	rdi, [rbp - 0x108]
  3a4112:      	mov	rsi, r11
  3a4115:      	mov	rcx, r12
  3a4118:      	mov	rdx, qword ptr [rbp - 0x60]
  3a411c:      	call	0x3a95f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a4121:      	mov	rax, qword ptr [rbp - 0x108]
  3a4128:      	mov	rdx, qword ptr [rbp - 0x100]
  3a412f:      	movzx	r13d, byte ptr [rbp - 0xf8]
  3a4137:      	mov	rsi, qword ptr [rbp - 0xf0]
  3a413e:      	test	rsi, rsi
  3a4141:      	je	0x3a42c4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18c4>
  3a4147:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a414c:      	mov	r11, qword ptr [rbp - 0x40]
  3a4150:      	mov	r8d, 0xffffffff
  3a4156:      	je	0x3a3dbd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13bd>
  3a415c:      	jmp	0x3a4514 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b14>
  3a4161:      	test	rax, rax
  3a4164:      	je	0x3a42d8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18d8>
  3a416a:      	mov	rax, qword ptr [rax]
  3a416d:      	test	rdi, rdi
  3a4170:      	je	0x3a42e5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18e5>
  3a4176:      	mov	rcx, qword ptr [rdi + 0x38]
  3a417a:      	test	rcx, rcx
  3a417d:      	mov	edx, 0x10000
  3a4182:      	cmove	rcx, rdx
  3a4186:      	mov	rdx, rax
  3a4189:      	or	rdx, rcx
  3a418c:      	shr	rdx, 0x20
  3a4190:      	je	0x3a42c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18c9>
  3a4196:      	xor	edx, edx
  3a4198:      	div	rcx
  3a419b:      	jmp	0x3a42cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cd>
  3a41a0:      	xor	eax, eax
  3a41a2:      	xor	edx, edx
  3a41a4:      	test	rax, rax
  3a41a7:      	jne	0x3a3e8e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x148e>
  3a41ad:      	jmp	0x3a3ec1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c1>
  3a41b2:      	mov	eax, 0x20
  3a41b7:      	nop	word ptr [rax + rax]
  3a41c0:      	xor	edx, edx
  3a41c2:      	mov	edi, dword ptr [r15 + 0x8]
  3a41c6:      	cmp	rdi, r8
  3a41c9:      	mov	r9, qword ptr [rbp - 0x38]
  3a41cd:      	mov	r10, qword ptr [rbp - 0x50]
  3a41d1:      	je	0x3a2b00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a41d7:      	cmp	r9, rdi
  3a41da:      	jbe	0x3a452d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2d>
  3a41e0:      	shl	rdi, 0x4
  3a41e4:      	mov	qword ptr [r10 + rdi], rax
  3a41e8:      	mov	qword ptr [r10 + rdi + 0x8], rdx
  3a41ed:      	jmp	0x3a2b00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a41f2:      	mov	rcx, qword ptr [rbp - 0x58]
  3a41f6:      	mov	rsi, qword ptr [rcx + 0x80]
  3a41fd:      	mov	edi, dword ptr [r15 + 0x4]
  3a4201:      	cmp	rsi, rdi
  3a4204:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a420a:      	imul	rax, rdi, 0xe0
  3a4211:      	add	rax, qword ptr [rcx + 0x78]
  3a4215:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a421b:      	mov	rcx, qword ptr [rbp - 0x60]
  3a421f:      	mov	rsi, qword ptr [rcx + 0x78]
  3a4223:      	mov	edi, dword ptr [rax + 0x18]
  3a4226:      	cmp	rsi, rdi
  3a4229:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a422f:      	shl	rdi, 0x3
  3a4233:      	add	rdi, qword ptr [rcx + 0x70]
  3a4237:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a423d:      	mov	rsi, qword ptr [rdi]
  3a4240:      	test	rsi, rsi
  3a4243:      	je	0x3a43b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b9>
  3a4249:      	mov	rcx, qword ptr [r11]
  3a424c:      	mov	rdx, qword ptr [r11 + 0x8]
  3a4250:      	mov	r8, qword ptr [rcx + 0x8]
  3a4254:      	mov	rcx, qword ptr [r11 + 0x10]
  3a4258:      	mov	rsi, qword ptr [rsi + 0x10]
  3a425c:      	mov	edi, dword ptr [rax + 0x14]
  3a425f:      	mov	eax, dword ptr [rax + 0x1c]
  3a4262:      	shl	rax, 0x20
  3a4266:      	or	rax, rdi
  3a4269:      	movlps	qword ptr [rsp + 0x18], xmm0
  3a426e:      	mov	qword ptr [rsp + 0x10], r10
  3a4273:      	mov	qword ptr [rsp + 0x8], rax
  3a4278:      	mov	qword ptr [rsp], r12
  3a427c:      	lea	rdi, [rbp - 0x130]
  3a4283:      	call	0x3c5030 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a4288:      	mov	rax, qword ptr [rbp - 0x130]
  3a428f:      	mov	rdx, qword ptr [rbp - 0x128]
  3a4296:      	movzx	r13d, byte ptr [rbp - 0x120]
  3a429e:      	mov	rsi, qword ptr [rbp - 0x118]
  3a42a5:      	test	rsi, rsi
  3a42a8:      	je	0x3a4304 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1904>
  3a42aa:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a42af:      	mov	r11, qword ptr [rbp - 0x40]
  3a42b3:      	mov	r8d, 0xffffffff
  3a42b9:      	je	0x3a3dbd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13bd>
  3a42bf:      	jmp	0x3a451d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b1d>
  3a42c4:      	jmp	0x3a3db3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13b3>
  3a42c9:      	xor	edx, edx
  3a42cb:      	div	ecx
  3a42cd:      	xor	edx, edx
  3a42cf:      	mov	rsi, qword ptr [rbp - 0x48]
  3a42d3:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a42d8:      	mov	rax, qword ptr [rdi + 0x8]
  3a42dc:      	test	rdi, rdi
  3a42df:      	jne	0x3a4176 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1776>
  3a42e5:      	mov	edi, 0x1
  3a42ea:      	call	0x5e6230 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a42ef:      	mov	r8d, 0xffffffff
  3a42f5:      	mov	r11, qword ptr [rbp - 0x40]
  3a42f9:      	xor	edx, edx
  3a42fb:      	mov	rsi, qword ptr [rbp - 0x48]
  3a42ff:      	jmp	0x3a41c2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17c2>
  3a4304:      	jmp	0x3a3db3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13b3>
  3a4309:      	mov	rax, qword ptr [rbp - 0x70]
  3a430d:      	mov	dword ptr [rax - 0x38], r14d
  3a4311:      	mov	rdi, r11
  3a4314:      	mov	r14, r11
  3a4317:      	call	0x3a6a30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a431c:      	test	al, al
  3a431e:      	je	0x3a4390 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1990>
  3a4320:      	jmp	0x3a44ff <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a4325:      	mov	rax, qword ptr [rbp - 0x70]
  3a4329:      	mov	dword ptr [rax - 0x38], r14d
  3a432d:      	mov	rdi, r11
  3a4330:      	mov	r14, r11
  3a4333:      	call	0x3a57d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a4338:      	jmp	0x3a4383 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1983>
  3a433a:      	mov	rax, qword ptr [rbp - 0x70]
  3a433e:      	mov	dword ptr [rax - 0x38], r14d
  3a4342:      	mov	rdi, r11
  3a4345:      	mov	r14, r11
  3a4348:      	call	0x3a6530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a434d:      	jmp	0x3a4383 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1983>
  3a434f:      	mov	rax, qword ptr [rbp - 0x70]
  3a4353:      	mov	dword ptr [rax - 0x38], r14d
  3a4357:      	mov	rdi, r11
  3a435a:      	mov	r14, r11
  3a435d:      	call	0x3a5340 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a4362:      	test	al, al
  3a4364:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a4369:      	je	0x3a4390 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1990>
  3a436b:      	jmp	0x3a4529 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b29>
  3a4370:      	mov	rax, qword ptr [rbp - 0x70]
  3a4374:      	mov	dword ptr [rax - 0x38], r14d
  3a4378:      	mov	rdi, r11
  3a437b:      	mov	r14, r11
  3a437e:      	call	0x3a6a30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a4383:      	test	al, al
  3a4385:      	movzx	r12d, byte ptr [rbp - 0x29]
  3a438a:      	jne	0x3a44ff <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a4390:      	mov	rcx, qword ptr [r14 + 0x48]
  3a4394:      	test	rcx, rcx
  3a4397:      	mov	r11, r14
  3a439a:      	mov	r8d, 0xffffffff
  3a43a0:      	mov	r14, qword ptr [rbp - 0x68]
  3a43a4:      	jg	0x3a2a41 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x41>
  3a43aa:      	mov	edi, 0x333a58
  3a43af:      	mov	esi, 0x29a060
  3a43b4:      	call	0x5e1be0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a43b9:      	mov	edi, 0x1
  3a43be:      	call	0x5e6230 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a43c3:      	jmp	0x3a44ff <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a43c8:      	mov	edi, 0x28
  3a43cd:      	call	0x5fc800 <GC_malloc@plt>
  3a43d2:      	test	rax, rax
  3a43d5:      	je	0x3a465c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a43db:      	mov	r14, rax
  3a43de:      	add	r13, r12
  3a43e1:      	mov	edi, 0x28
  3a43e6:      	call	0x5e1ec0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a43eb:      	mov	rax, qword ptr [r13 + 0x8]
  3a43ef:      	movaps	xmm0, xmmword ptr [rip + 0x31543a] # 0x6b9830 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a43f6:      	movups	xmmword ptr [r14], xmm0
  3a43fa:      	mov	qword ptr [r14 + 0x10], rax
  3a43fe:      	mov	qword ptr [r14 + 0x18], 0x29a070
  3a4406:      	mov	qword ptr [r14 + 0x20], 0xe
  3a440e:      	mov	esi, 0x2baa70
  3a4413:      	jmp	0x3a44f1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af1>
  3a4418:      	mov	rdi, r11
  3a441b:      	mov	rsi, rax
  3a441e:      	jmp	0x3a44f8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af8>
  3a4423:      	mov	r13d, edx
  3a4426:      	jmp	0x3a448c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8c>
  3a4428:      	mov	r14, qword ptr [rbp - 0x138]
  3a442f:      	mov	edi, 0x8
  3a4434:      	call	0x5fc800 <GC_malloc@plt>
  3a4439:      	test	rax, rax
  3a443c:      	je	0x3a465c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a4442:      	mov	r15, rax
  3a4445:      	mov	edi, 0x8
  3a444a:      	call	0x5e1ec0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a444f:      	mov	edx, 0x299fe0
  3a4454:      	mov	rdi, r13
  3a4457:      	mov	rsi, r14
  3a445a:      	mov	rcx, r15
  3a445d:      	call	0x5b8af0 <errors.As>
  3a4462:      	test	al, 0x1
  3a4464:      	je	0x3a4476 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a76>
  3a4466:      	mov	rax, qword ptr [r15]
  3a4469:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a4470:      	mov	ecx, dword ptr [rcx + 0x50]
  3a4473:      	mov	dword ptr [rax + 0x4], ecx
  3a4476:      	mov	rdi, qword ptr [rbp - 0x40]
  3a447a:      	mov	rsi, r13
  3a447d:      	jmp	0x3a44f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af5>
  3a447f:      	mov	r13b, 0x3
  3a4482:      	jmp	0x3a448c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8c>
  3a4484:      	mov	r13b, 0x1
  3a4487:      	jmp	0x3a448c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8c>
  3a4489:      	mov	r13b, 0x9
  3a448c:      	mov	edi, 0x10
  3a4491:      	call	0x5fc8b0 <GC_malloc_atomic@plt>
  3a4496:      	test	rax, rax
  3a4499:      	je	0x3a465c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a449f:      	mov	r14, rax
  3a44a2:      	mov	edi, 0x10
  3a44a7:      	call	0x5e1ec0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a44ac:      	xorps	xmm0, xmm0
  3a44af:      	movups	xmmword ptr [r14], xmm0
  3a44b3:      	mov	rax, qword ptr [rbp - 0x58]
  3a44b7:      	mov	rsi, qword ptr [rax + 0x80]
  3a44be:      	mov	edi, dword ptr [r15 + 0x4]
  3a44c2:      	cmp	rsi, rdi
  3a44c5:      	jbe	0x3a453b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3b>
  3a44c7:      	mov	rax, qword ptr [rax + 0x78]
  3a44cb:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a44d2:      	mov	ecx, dword ptr [rcx + 0x50]
  3a44d5:      	imul	rdx, rdi, 0xe0
  3a44dc:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a44e1:      	mov	byte ptr [r14], r13b
  3a44e4:      	mov	dword ptr [r14 + 0x4], ecx
  3a44e8:      	mov	qword ptr [r14 + 0x8], rax
  3a44ec:      	mov	esi, 0x29a3f8
  3a44f1:      	mov	rdi, qword ptr [rbp - 0x40]
  3a44f5:      	mov	rdx, r14
  3a44f8:      	call	0x3a23d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a44fd:      	mov	al, 0x3
  3a44ff:      	mov	rdx, rbx
  3a4502:      	add	rsp, 0x2a8
  3a4509:      	pop	rbx
  3a450a:      	pop	r12
  3a450c:      	pop	r13
  3a450e:      	pop	r14
  3a4510:      	pop	r15
  3a4512:      	pop	rbp
  3a4513:      	ret
  3a4514:      	mov	rdx, qword ptr [rbp - 0xe8]
  3a451b:      	jmp	0x3a4524 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b24>
  3a451d:      	mov	rdx, qword ptr [rbp - 0x110]
  3a4524:      	mov	rdi, r11
  3a4527:      	jmp	0x3a44f8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af8>
  3a4529:      	mov	al, 0x1
  3a452b:      	jmp	0x3a44ff <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aff>
  3a452d:      	mov	rsi, r9
  3a4530:      	call	0x5e1fa0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4535:      	mov	rdi, r14
  3a4538:      	mov	rsi, r13
  3a453b:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4540:      	mov	rdi, r15
  3a4543:      	mov	rsi, r13
  3a4546:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a454b:      	mov	rdi, rax
  3a454e:      	mov	rsi, r9
  3a4551:      	call	0x5e1fa0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4556:      	mov	rsi, qword ptr [rbp - 0x38]
  3a455a:      	call	0x5e1fa0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a455f:      	mov	rdi, r9
  3a4562:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4567:      	mov	rsi, qword ptr [rbp - 0x38]
  3a456b:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4570:      	mov	rdi, rax
  3a4573:      	mov	esi, 0x1
  3a4578:      	mov	ecx, 0x2
  3a457d:      	call	0x5e2ce0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4582:      	mov	esi, 0x1
  3a4587:      	mov	rdx, rax
  3a458a:      	mov	ecx, 0x3
  3a458f:      	call	0x5e2ce0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4594:      	mov	rdi, rax
  3a4597:      	mov	rsi, r9
  3a459a:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a459f:      	mov	rsi, rcx
  3a45a2:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45a7:      	mov	rdi, rsi
  3a45aa:      	mov	rsi, rcx
  3a45ad:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45b2:      	mov	rdi, rcx
  3a45b5:      	mov	rsi, r9
  3a45b8:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45bd:      	mov	edi, 0x1
  3a45c2:      	mov	esi, 0x1
  3a45c7:      	call	0x5e1fa0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a45cc:      	xor	edi, edi
  3a45ce:      	xor	esi, esi
  3a45d0:      	call	0x5e1fa0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a45d5:      	xor	esi, esi
  3a45d7:      	mov	rdx, rcx
  3a45da:      	mov	ecx, 0x3
  3a45df:      	call	0x5e2ce0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a45e4:      	mov	rdi, rcx
  3a45e7:      	xor	esi, esi
  3a45e9:      	mov	ecx, 0x2
  3a45ee:      	call	0x5e2ce0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a45f3:      	mov	rdi, rax
  3a45f6:      	mov	rsi, qword ptr [rbp - 0x38]
  3a45fa:      	call	0x5e34d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45ff:      	mov	edi, 0x3
  3a4604:      	mov	esi, 0x3
  3a4609:      	call	0x5e1fa0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a460e:      	mov	edi, 0x2
  3a4613:      	mov	esi, 0x2
  3a4618:      	call	0x5e1fa0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a461d:      	mov	rdi, r15
  3a4620:      	xor	esi, esi
  3a4622:      	mov	ecx, 0x3
  3a4627:      	call	0x5e2ce0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a462c:      	mov	rdi, rdx
  3a462f:      	xor	esi, esi
  3a4631:      	mov	rdx, rax
  3a4634:      	mov	ecx, 0x2
  3a4639:      	call	0x5e2ce0 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a463e:      	mov	edi, 0x3295a8
  3a4643:      	mov	esi, 0x3298d8
  3a4648:      	call	0x5e1be0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a464d:      	mov	edi, 0x333a58
  3a4652:      	mov	esi, 0x29a080
  3a4657:      	call	0x5e1be0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a465c:      	mov	edi, 0x333a58
  3a4661:      	mov	esi, 0x329580
  3a4666:      	call	0x5e1be0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a466b:      	int3
  3a466c:      	int3
  3a466d:      	int3
  3a466e:      	int3
  3a466f:      	int3
