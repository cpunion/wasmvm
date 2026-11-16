
/home/runner/work/wasmvm/wasmvm/main-regression/scalar-base:	file format elf64-x86-64

Disassembly of section .text:

00000000003a29f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a29f0:      	push	rbp
  3a29f1:      	mov	rbp, rsp
  3a29f4:      	push	r15
  3a29f6:      	push	r14
  3a29f8:      	push	r13
  3a29fa:      	push	r12
  3a29fc:      	push	rbx
  3a29fd:      	sub	rsp, 0x2a8
  3a2a04:      	mov	rcx, qword ptr [rdi + 0x48]
  3a2a08:      	test	rcx, rcx
  3a2a0b:      	jle	0x3a465b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6b>
  3a2a11:      	mov	rbx, rsi
  3a2a14:      	mov	r11, rdi
  3a2a17:      	movzx	eax, byte ptr [rdi + 0x20]
  3a2a1b:      	mov	byte ptr [rbp - 0x39], al
  3a2a1e:      	lea	rax, [rbp - 0x90]
  3a2a25:      	mov	qword ptr [rbp - 0x68], rax
  3a2a29:      	mov	qword ptr [rbp - 0x38], rdi
  3a2a2d:      	mov	rdi, qword ptr [r11 + 0x40]
  3a2a31:      	shl	rcx, 0x6
  3a2a35:      	mov	byte ptr [rdi + rcx - 0x30], 0x0
  3a2a3a:      	mov	rsi, qword ptr [rdi + rcx - 0x40]
  3a2a3f:      	mov	r10, qword ptr [rsi + 0x48]
  3a2a43:      	mov	r13, qword ptr [r10 + 0x98]
  3a2a4a:      	mov	r9, qword ptr [rdi + rcx - 0x28]
  3a2a4f:      	mov	rdx, qword ptr [r11 + 0x88]
  3a2a56:      	lea	rax, [r9 + r13]
  3a2a5a:      	cmp	rax, rdx
  3a2a5d:      	ja	0x3a4573 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b83>
  3a2a63:      	cmp	r9, rax
  3a2a66:      	ja	0x3a4585 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b95>
  3a2a6c:      	mov	r15, qword ptr [r10 + 0xb8]
  3a2a73:      	mov	r12d, dword ptr [rdi + rcx - 0x38]
  3a2a78:      	sub	rdx, r9
  3a2a7b:      	shl	r9, 0x4
  3a2a7f:      	test	rdx, rdx
  3a2a82:      	mov	eax, 0x0
  3a2a87:      	cmovle	r9, rax
  3a2a8b:      	cmp	r15, r12
  3a2a8e:      	jbe	0x3a4537 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a2a94:      	add	rdi, rcx
  3a2a97:      	mov	qword ptr [rbp - 0x60], rdi
  3a2a9b:      	mov	qword ptr [rbp - 0xa8], rsi
  3a2aa2:      	mov	rax, qword ptr [rsi + 0x40]
  3a2aa6:      	mov	qword ptr [rbp - 0x50], rax
  3a2aaa:      	mov	rcx, qword ptr [r10 + 0xb0]
  3a2ab1:      	add	r9, qword ptr [r11 + 0x78]
  3a2ab5:      	mov	r14d, r12d
  3a2ab8:      	mov	qword ptr [rbp - 0x78], r13
  3a2abc:      	mov	qword ptr [rbp - 0x30], r10
  3a2ac0:      	mov	qword ptr [rbp - 0x48], r9
  3a2ac4:      	mov	qword ptr [rbp - 0xb0], r15
  3a2acb:      	mov	qword ptr [rbp - 0x58], rcx
  3a2acf:      	mov	esi, 0xffffffff
  3a2ad4:      	jmp	0x3a2b03 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x113>
  3a2ad6:      	nop	word ptr cs:[rax + rax]
  3a2ae0:      	inc	r14d
  3a2ae3:      	mov	r15d, r14d
  3a2ae6:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2aea:      	mov	r10, r8
  3a2aed:      	mov	r12d, r15d
  3a2af0:      	mov	r14d, r15d
  3a2af3:      	mov	r15, qword ptr [rbp - 0xb0]
  3a2afa:      	cmp	r15, r12
  3a2afd:      	jbe	0x3a4537 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a2b03:      	shl	r12, 0x5
  3a2b07:      	add	r12, rcx
  3a2b0a:      	test	byte ptr [rbp - 0x39], 0x1
  3a2b0e:      	je	0x3a2b70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x180>
  3a2b10:      	mov	rsi, qword ptr [r10 + 0x80]
  3a2b17:      	mov	edi, dword ptr [r12 + 0x4]
  3a2b1c:      	cmp	rsi, rdi
  3a2b1f:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a2b25:      	mov	r15, qword ptr [r10 + 0x78]
  3a2b29:      	imul	r13, rdi, 0xe0
  3a2b30:      	mov	eax, dword ptr [r15 + r13]
  3a2b34:      	cmp	eax, 0x5
  3a2b37:      	setne	cl
  3a2b3a:      	cmp	eax, 0xb
  3a2b3d:      	setne	al
  3a2b40:      	and	al, cl
  3a2b42:      	movzx	ecx, al
  3a2b45:      	mov	rax, qword ptr [r11 + 0x150]
  3a2b4c:      	sub	rax, rcx
  3a2b4f:      	jb	0x3a43bc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19cc>
  3a2b55:      	mov	qword ptr [r11 + 0x150], rax
  3a2b5c:      	mov	r13, qword ptr [rbp - 0x78]
  3a2b60:      	mov	r15, qword ptr [rbp - 0xb0]
  3a2b67:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2b6b:      	mov	esi, 0xffffffff
  3a2b70:      	inc	rbx
  3a2b73:      	test	bl, 0x3f
  3a2b76:      	jne	0x3a2bf5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x205>
  3a2b78:      	mov	rax, qword ptr [r11 + 0x160]
  3a2b7f:      	test	rax, rax
  3a2b82:      	je	0x3a2b97 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7>
  3a2b84:      	test	r11, r11
  3a2b87:      	je	0x3a2c9a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2aa>
  3a2b8d:      	mov	eax, dword ptr [rax]
  3a2b8f:      	test	eax, eax
  3a2b91:      	je	0x3a3103 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x713>
  3a2b97:      	mov	rax, qword ptr [r11 + 0x8]
  3a2b9b:      	test	rax, rax
  3a2b9e:      	je	0x3a464c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a2ba4:      	mov	rdi, qword ptr [r11 + 0x10]
  3a2ba8:      	mov	rcx, qword ptr [rax + 0x8]
  3a2bac:      	test	byte ptr [rcx + 0x17], 0x20
  3a2bb0:      	je	0x3a2bc9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d9>
  3a2bb2:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a2bb6:      	and	cl, 0x1f
  3a2bb9:      	cmp	cl, 0x16
  3a2bbc:      	je	0x3a2bc9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d9>
  3a2bbe:      	mov	qword ptr [rbp - 0x90], rdi
  3a2bc5:      	mov	rdi, qword ptr [rbp - 0x68]
  3a2bc9:      	call	qword ptr [rax + 0x28]
  3a2bcc:      	mov	esi, 0xffffffff
  3a2bd1:      	mov	r10, qword ptr [rbp - 0x30]
  3a2bd5:      	mov	r9, qword ptr [rbp - 0x48]
  3a2bd9:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2bdd:      	test	rax, rax
  3a2be0:      	je	0x3a2cbe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2ce>
  3a2be6:      	cmp	qword ptr [rax + 0x8], 0x0
  3a2beb:      	mov	r11, qword ptr [rbp - 0x38]
  3a2bef:      	jne	0x3a440c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a1c>
  3a2bf5:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a2bfb:      	je	0x3a2cd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c01:      	test	byte ptr [rbp - 0x39], 0x1
  3a2c05:      	je	0x3a2c30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2c07:      	mov	edx, dword ptr [r12]
  3a2c0b:      	cmp	edx, 0x20
  3a2c0e:      	je	0x3a2cd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c14:      	cmp	edx, 0x41
  3a2c17:      	je	0x3a2cd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c1d:      	jmp	0x3a2cd4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e4>
  3a2c22:      	nop	word ptr cs:[rax + rax]
  3a2c30:      	mov	edx, dword ptr [r12]
  3a2c34:      	cmp	edx, 0x20
  3a2c37:      	je	0x3a2c42 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x252>
  3a2c39:      	cmp	edx, 0x41
  3a2c3c:      	jne	0x3a2cd4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e4>
  3a2c42:      	mov	eax, ebx
  3a2c44:      	not	eax
  3a2c46:      	test	al, 0x3f
  3a2c48:      	je	0x3a2cd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c4e:      	cmp	edx, 0x41
  3a2c51:      	jne	0x3a2c77 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x287>
  3a2c53:      	mov	edi, dword ptr [r12 + 0x8]
  3a2c58:      	cmp	r13, rdi
  3a2c5b:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2c61:      	mov	eax, dword ptr [r12 + 0x18]
  3a2c66:      	shl	rdi, 0x4
  3a2c6a:      	mov	qword ptr [r9 + rdi], rax
  3a2c6e:      	mov	qword ptr [r9 + rdi + 0x8], 0x0
  3a2c77:      	inc	r14d
  3a2c7a:      	cmp	r15, r14
  3a2c7d:      	jbe	0x3a451f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2f>
  3a2c83:      	mov	rax, r14
  3a2c86:      	shl	rax, 0x5
  3a2c8a:      	lea	r12, [rcx + rax]
  3a2c8e:      	inc	rbx
  3a2c91:      	cmp	dword ptr [rcx + rax + 0x1c], 0x0
  3a2c96:      	jne	0x3a2c30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2c98:      	jmp	0x3a2cd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2c9a:      	mov	edi, 0x1
  3a2c9f:      	call	0x5e6250 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a2ca4:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2ca8:      	mov	r9, qword ptr [rbp - 0x48]
  3a2cac:      	mov	r10, qword ptr [rbp - 0x30]
  3a2cb0:      	mov	esi, 0xffffffff
  3a2cb5:      	test	rax, rax
  3a2cb8:      	jne	0x3a2be6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f6>
  3a2cbe:      	mov	r11, qword ptr [rbp - 0x38]
  3a2cc2:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a2cc8:      	jne	0x3a2c01 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x211>
  3a2cce:      	nop
  3a2cd0:      	mov	edx, dword ptr [r12]
  3a2cd4:      	lea	eax, [rdx - 0xe0]
  3a2cda:      	cmp	eax, 0xffff20
  3a2cdf:      	jae	0x3a2f40 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x550>
  3a2ce5:      	mov	eax, edx
  3a2ce7:      	shr	eax, 0x10
  3a2cea:      	cmp	eax, 0xfb
  3a2cef:      	je	0x3a2ea2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4b2>
  3a2cf5:      	cmp	eax, 0xfc
  3a2cfa:      	je	0x3a2d84 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x394>
  3a2d00:      	cmp	eax, 0xfd
  3a2d05:      	jne	0x3a2f26 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x536>
  3a2d0b:      	mov	edi, dword ptr [r12 + 0xc]
  3a2d10:      	cmp	rdi, rsi
  3a2d13:      	je	0x3a31bc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7cc>
  3a2d19:      	cmp	r13, rdi
  3a2d1c:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2d22:      	shl	rdi, 0x4
  3a2d26:      	mov	rcx, qword ptr [r9 + rdi]
  3a2d2a:      	mov	r10, qword ptr [r9 + rdi + 0x8]
  3a2d2f:      	mov	edi, dword ptr [r12 + 0x10]
  3a2d34:      	cmp	rdi, rsi
  3a2d37:      	je	0x3a31cf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7df>
  3a2d3d:      	mov	esi, 0xffffffff
  3a2d42:      	cmp	r13, rdi
  3a2d45:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2d4b:      	shl	rdi, 0x4
  3a2d4f:      	mov	rax, qword ptr [r9 + rdi]
  3a2d53:      	mov	r8, qword ptr [r9 + rdi + 0x8]
  3a2d58:      	mov	edi, dword ptr [r12 + 0x14]
  3a2d5d:      	cmp	rdi, rsi
  3a2d60:      	je	0x3a31e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7f7>
  3a2d66:      	cmp	r13, rdi
  3a2d69:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2d6f:      	shl	rdi, 0x4
  3a2d73:      	mov	rsi, r9
  3a2d76:      	mov	r9, qword ptr [r9 + rdi]
  3a2d7a:      	mov	r11, qword ptr [rsi + rdi + 0x8]
  3a2d7f:      	jmp	0x3a31ed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7fd>
  3a2d84:      	add	edx, 0xff03ffed
  3a2d8a:      	cmp	edx, 0x4
  3a2d8d:      	jae	0x3a2f40 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x550>
  3a2d93:      	mov	rdx, qword ptr [r10 + 0xd8]
  3a2d9a:      	mov	edi, dword ptr [r12 + 0x18]
  3a2d9f:      	mov	eax, dword ptr [r12 + 0x1c]
  3a2da4:      	lea	ecx, [rdi + rax]
  3a2da7:      	add	ecx, 0x2
  3a2daa:      	cmp	rdx, rcx
  3a2dad:      	jb	0x3a45df <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bef>
  3a2db3:      	cmp	edi, ecx
  3a2db5:      	ja	0x3a45d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1be0>
  3a2dbb:      	sub	rdx, rdi
  3a2dbe:      	lea	r8, [4*rdi]
  3a2dc6:      	test	rdx, rdx
  3a2dc9:      	mov	eax, 0x0
  3a2dce:      	cmovle	r8, rax
  3a2dd2:      	cmp	ecx, edi
  3a2dd4:      	je	0x3a45c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd7>
  3a2dda:      	add	r8, qword ptr [r10 + 0xc8]
  3a2de1:      	mov	eax, dword ptr [r8]
  3a2de4:      	cmp	r13, rax
  3a2de7:      	jbe	0x3a4568 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b78>
  3a2ded:      	sub	rcx, rdi
  3a2df0:      	cmp	ecx, 0x1
  3a2df3:      	je	0x3a45b8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a2df9:      	mov	edi, dword ptr [r8 + 0x4]
  3a2dfd:      	cmp	r13, rdi
  3a2e00:      	jbe	0x3a4558 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b68>
  3a2e06:      	shl	rax, 0x4
  3a2e0a:      	mov	rdx, r9
  3a2e0d:      	mov	r9, qword ptr [r9 + rax]
  3a2e11:      	shl	rdi, 0x4
  3a2e15:      	mov	rsi, qword ptr [rdx + rdi]
  3a2e19:      	mov	r10d, dword ptr [r12]
  3a2e1d:      	lea	eax, [r10 - 0xfc0013]
  3a2e24:      	cmp	eax, 0x2
  3a2e27:      	jae	0x3a332e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x93e>
  3a2e2d:      	cmp	ecx, 0x3
  3a2e30:      	jb	0x3a45fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c0d>
  3a2e36:      	mov	edi, dword ptr [r8 + 0x8]
  3a2e3a:      	cmp	r13, rdi
  3a2e3d:      	jbe	0x3a4558 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b68>
  3a2e43:      	cmp	ecx, 0x3
  3a2e46:      	je	0x3a45ee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bfe>
  3a2e4c:      	mov	eax, dword ptr [r8 + 0xc]
  3a2e50:      	cmp	r13, rax
  3a2e53:      	jbe	0x3a4568 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b78>
  3a2e59:      	shl	rdi, 0x4
  3a2e5d:      	mov	rdx, qword ptr [rbp - 0x48]
  3a2e61:      	mov	rdi, qword ptr [rdx + rdi]
  3a2e65:      	shl	rax, 0x4
  3a2e69:      	mov	rdx, qword ptr [rdx + rax]
  3a2e6d:      	cmp	r10d, 0xfc0013
  3a2e74:      	jne	0x3a346b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa7b>
  3a2e7a:      	lea	rax, [rdi + r9]
  3a2e7e:      	mov	r10, rdi
  3a2e81:      	and	r10, r9
  3a2e84:      	or	rdi, r9
  3a2e87:      	mov	r9, rax
  3a2e8a:      	not	r9
  3a2e8d:      	and	r9, rdi
  3a2e90:      	or	r9, r10
  3a2e93:      	shr	r9, 0x3f
  3a2e97:      	add	rdx, rsi
  3a2e9a:      	add	rdx, r9
  3a2e9d:      	jmp	0x3a3490 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa0>
  3a2ea2:      	add	edx, -0x2
  3a2ea5:      	cmp	dx, 0x21
  3a2ea9:      	jae	0x3a309a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6aa>
  3a2eaf:      	lea	rax, [rsi + 0x1c003e10]
  3a2eb6:      	movzx	ecx, dx
  3a2eb9:      	bt	rax, rcx
  3a2ebd:      	jae	0x3a309a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6aa>
  3a2ec3:      	mov	edi, dword ptr [r12 + 0xc]
  3a2ec8:      	mov	eax, 0xffffffff
  3a2ecd:      	cmp	rdi, rsi
  3a2ed0:      	je	0x3a350b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb1b>
  3a2ed6:      	cmp	r13, rdi
  3a2ed9:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2edf:      	shl	rdi, 0x4
  3a2ee3:      	mov	r8, qword ptr [r9 + rdi]
  3a2ee7:      	mov	edi, dword ptr [r12 + 0x10]
  3a2eec:      	cmp	rdi, rax
  3a2eef:      	je	0x3a351c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb2c>
  3a2ef5:      	cmp	r13, rdi
  3a2ef8:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2efe:      	shl	rdi, 0x4
  3a2f02:      	mov	rax, qword ptr [r9 + rdi]
  3a2f06:      	mov	r9, qword ptr [r9 + rdi + 0x8]
  3a2f0b:      	mov	r10d, 0x0
  3a2f11:      	mov	r11d, 0x0
  3a2f17:      	cmp	dx, 0xc
  3a2f1b:      	je	0x3a3533 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb43>
  3a2f21:      	jmp	0x3a3573 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb83>
  3a2f26:      	cmp	edx, 0xe7
  3a2f2c:      	jb	0x3a4335 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1945>
  3a2f32:      	nop	word ptr cs:[rax + rax]
  3a2f40:      	mov	edi, dword ptr [r12 + 0xc]
  3a2f45:      	cmp	rdi, rsi
  3a2f48:      	je	0x3a2fc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d0>
  3a2f4a:      	cmp	r13, rdi
  3a2f4d:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2f53:      	mov	r8, r10
  3a2f56:      	mov	rax, rdi
  3a2f59:      	shl	rax, 0x4
  3a2f5d:      	mov	r10, qword ptr [r9 + rax]
  3a2f61:      	mov	ecx, dword ptr [r12 + 0x10]
  3a2f66:      	cmp	rcx, rsi
  3a2f69:      	je	0x3a2fd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5e0>
  3a2f6b:      	cmp	r13, rcx
  3a2f6e:      	jbe	0x3a454d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5d>
  3a2f74:      	mov	rax, rcx
  3a2f77:      	shl	rax, 0x4
  3a2f7b:      	mov	rdx, qword ptr [r9 + rax]
  3a2f7f:      	mov	r9d, dword ptr [r12]
  3a2f83:      	lea	eax, [r9 - 0x45]
  3a2f87:      	cmp	eax, 0x7f
  3a2f8a:      	ja	0x3a2fdf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5ef>
  3a2f8c:      	ja	0x3a3e04 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1414>
  3a2f92:      	jmp	qword ptr [8*rax + 0x203aa8]
  3a2f99:      	mov	edi, r9d
  3a2f9c:      	mov	rsi, r10
  3a2f9f:      	call	0x3d0a80 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a2fa4:      	test	dl, dl
  3a2fa6:      	jne	0x3a4417 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a27>
  3a2fac:      	xor	ecx, ecx
  3a2fae:      	mov	r11, qword ptr [rbp - 0x38]
  3a2fb2:      	mov	r8, qword ptr [rbp - 0x30]
  3a2fb6:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a2fbb:      	nop	dword ptr [rax + rax]
  3a2fc0:      	mov	r8, r10
  3a2fc3:      	xor	r10d, r10d
  3a2fc6:      	mov	ecx, dword ptr [r12 + 0x10]
  3a2fcb:      	cmp	rcx, rsi
  3a2fce:      	jne	0x3a2f6b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x57b>
  3a2fd0:      	xor	edx, edx
  3a2fd2:      	mov	r9d, dword ptr [r12]
  3a2fd6:      	lea	eax, [r9 - 0x45]
  3a2fda:      	cmp	eax, 0x7f
  3a2fdd:      	jbe	0x3a2f8c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x59c>
  3a2fdf:      	lea	eax, [r9 - 0x1]
  3a2fe3:      	cmp	eax, 0xd3
  3a2fe8:      	mov	esi, 0xffffffff
  3a2fed:      	ja	0x3a30d6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6e6>
  3a2ff3:      	jmp	qword ptr [8*rax + 0x2033c8]
  3a2ffa:      	mov	rsi, qword ptr [r8 + 0x80]
  3a3001:      	mov	edi, dword ptr [r12 + 0x4]
  3a3006:      	cmp	rsi, rdi
  3a3009:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a300f:      	imul	rax, rdi, 0xe0
  3a3016:      	add	rax, qword ptr [r8 + 0x78]
  3a301a:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3020:      	mov	rcx, qword ptr [rbp - 0x50]
  3a3024:      	mov	rsi, qword ptr [rcx + 0x78]
  3a3028:      	mov	edi, dword ptr [rax + 0x18]
  3a302b:      	cmp	rsi, rdi
  3a302e:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3034:      	shl	rdi, 0x3
  3a3038:      	add	rdi, qword ptr [rcx + 0x70]
  3a303c:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3042:      	mov	rcx, qword ptr [rdi]
  3a3045:      	test	rcx, rcx
  3a3048:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a304e:      	mov	rsi, qword ptr [rcx + 0x10]
  3a3052:      	mov	ecx, dword ptr [rax + 0x14]
  3a3055:      	mov	r8d, dword ptr [rax + 0x1c]
  3a3059:      	shl	r8, 0x20
  3a305d:      	or	r8, rcx
  3a3060:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a3065:      	je	0x3a3072 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x682>
  3a3067:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a306c:      	je	0x3a3446 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa56>
  3a3072:      	xorps	xmm0, xmm0
  3a3075:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a307c:      	movups	xmmword ptr [rsp], xmm0
  3a3080:      	mov	r15, qword ptr [rbp - 0x68]
  3a3084:      	mov	rdi, r15
  3a3087:      	mov	edx, r9d
  3a308a:      	mov	rcx, r10
  3a308d:      	xor	r9d, r9d
  3a3090:      	call	0x3c6ba0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a3095:      	jmp	0x3a3427 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa37>
  3a309a:      	mov	r15, qword ptr [rbp - 0x60]
  3a309e:      	mov	dword ptr [r15 - 0x38], r14d
  3a30a2:      	mov	rdi, r11
  3a30a5:      	call	0x3a48d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a30aa:      	mov	r9, qword ptr [rbp - 0x48]
  3a30ae:      	mov	r10, qword ptr [rbp - 0x30]
  3a30b2:      	mov	esi, 0xffffffff
  3a30b7:      	mov	r11, qword ptr [rbp - 0x38]
  3a30bb:      	test	al, al
  3a30bd:      	mov	rcx, qword ptr [rbp - 0x58]
  3a30c1:      	jne	0x3a44f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b05>
  3a30c7:      	inc	r14d
  3a30ca:      	mov	dword ptr [r15 - 0x38], r14d
  3a30ce:      	mov	r15d, r14d
  3a30d1:      	jmp	0x3a2aed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a30d6:      	lea	eax, [r9 - 0xfc0000]
  3a30dd:      	cmp	eax, 0x8
  3a30e0:      	jae	0x3a32d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8e0>
  3a30e6:      	mov	edi, r9d
  3a30e9:      	mov	rsi, r10
  3a30ec:      	call	0x3d0a80 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a30f1:      	mov	r8, qword ptr [rbp - 0x30]
  3a30f5:      	mov	r11, qword ptr [rbp - 0x38]
  3a30f9:      	mov	r15d, edx
  3a30fc:      	xor	ecx, ecx
  3a30fe:      	jmp	0x3a3e5b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x146b>
  3a3103:      	xor	eax, eax
  3a3105:      	xor	edx, edx
  3a3107:      	test	rax, rax
  3a310a:      	jne	0x3a2be6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f6>
  3a3110:      	jmp	0x3a2cbe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2ce>
  3a3115:      	mov	rsi, qword ptr [r8 + 0x80]
  3a311c:      	mov	edi, dword ptr [r12 + 0x4]
  3a3121:      	cmp	rsi, rdi
  3a3124:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a312a:      	imul	rax, rdi, 0xe0
  3a3131:      	add	rax, qword ptr [r8 + 0x78]
  3a3135:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a313b:      	mov	rcx, qword ptr [rbp - 0x50]
  3a313f:      	mov	rsi, qword ptr [rcx + 0x78]
  3a3143:      	mov	edi, dword ptr [rax + 0x18]
  3a3146:      	cmp	rsi, rdi
  3a3149:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a314f:      	shl	rdi, 0x3
  3a3153:      	add	rdi, qword ptr [rcx + 0x70]
  3a3157:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a315d:      	mov	rcx, qword ptr [rdi]
  3a3160:      	test	rcx, rcx
  3a3163:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3169:      	mov	rdi, qword ptr [rcx + 0x10]
  3a316d:      	mov	esi, dword ptr [rax + 0x14]
  3a3170:      	mov	ecx, dword ptr [rax + 0x1c]
  3a3173:      	shl	rcx, 0x20
  3a3177:      	or	rcx, rsi
  3a317a:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a317f:      	je	0x3a318c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x79c>
  3a3181:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a3186:      	je	0x3a36b1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcc1>
  3a318c:      	mov	qword ptr [rbp - 0x90], rdx
  3a3193:      	mov	qword ptr [rbp - 0x88], 0x0
  3a319e:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a31a5:      	movups	xmmword ptr [rsp], xmm0
  3a31a9:      	mov	esi, r9d
  3a31ac:      	mov	rdx, r10
  3a31af:      	xor	r8d, r8d
  3a31b2:      	call	0x3c8090 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a31b7:      	jmp	0x3a3e4c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x145c>
  3a31bc:      	xor	r10d, r10d
  3a31bf:      	xor	ecx, ecx
  3a31c1:      	mov	edi, dword ptr [r12 + 0x10]
  3a31c6:      	cmp	rdi, rsi
  3a31c9:      	jne	0x3a2d3d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x34d>
  3a31cf:      	mov	esi, 0xffffffff
  3a31d4:      	xor	eax, eax
  3a31d6:      	xor	r8d, r8d
  3a31d9:      	mov	edi, dword ptr [r12 + 0x14]
  3a31de:      	cmp	rdi, rsi
  3a31e1:      	jne	0x3a2d66 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x376>
  3a31e7:      	xor	r9d, r9d
  3a31ea:      	xor	r11d, r11d
  3a31ed:      	mov	r15, qword ptr [rbp - 0x30]
  3a31f1:      	mov	rsi, qword ptr [r15 + 0x80]
  3a31f8:      	mov	edi, dword ptr [r12 + 0x4]
  3a31fd:      	cmp	rsi, rdi
  3a3200:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3206:      	imul	rdi, rdi, 0xe0
  3a320d:      	add	rdi, qword ptr [r15 + 0x78]
  3a3211:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3217:      	movzx	esi, dx
  3a321a:      	cmp	esi, 0xb
  3a321d:      	setb	r15b
  3a3221:      	mov	esi, edx
  3a3223:      	and	esi, 0xfffc
  3a3229:      	cmp	esi, 0x54
  3a322c:      	sete	r13b
  3a3230:      	or	r13b, r15b
  3a3233:      	jne	0x3a3247 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x857>
  3a3235:      	movzx	r15d, dx
  3a3239:      	lea	r13d, [r15 - 0x5c]
  3a323d:      	cmp	r13d, 0x2
  3a3241:      	jae	0x3a335e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x96e>
  3a3247:      	mov	r10, qword ptr [rbp - 0x50]
  3a324b:      	mov	rsi, qword ptr [r10 + 0x78]
  3a324f:      	mov	r9d, dword ptr [rdi + 0x18]
  3a3253:      	cmp	rsi, r9
  3a3256:      	jbe	0x3a4560 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b70>
  3a325c:      	shl	r9, 0x3
  3a3260:      	add	r9, qword ptr [r10 + 0x70]
  3a3264:      	mov	r15, qword ptr [rbp - 0x68]
  3a3268:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a326e:      	mov	rsi, qword ptr [r9]
  3a3271:      	test	rsi, rsi
  3a3274:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a327a:      	mov	rsi, qword ptr [rsi + 0x10]
  3a327e:      	mov	r9d, dword ptr [rdi + 0x14]
  3a3282:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a3286:      	shl	r10, 0x20
  3a328a:      	or	r10, r9
  3a328d:      	movzx	edi, byte ptr [rdi + 0x40]
  3a3291:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a3296:      	je	0x3a32a3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8b3>
  3a3298:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a329d:      	je	0x3a33fb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa0b>
  3a32a3:      	mov	qword ptr [rbp - 0xa0], rax
  3a32aa:      	mov	qword ptr [rbp - 0x98], r8
  3a32b1:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a32b8:      	movups	xmmword ptr [rsp], xmm0
  3a32bc:      	movzx	r9d, dil
  3a32c0:      	mov	rdi, r15
  3a32c3:      	mov	r8, r10
  3a32c6:      	call	0x3c6ba0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a32cb:      	jmp	0x3a3423 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa33>
  3a32d0:      	lea	eax, [r9 - 0x1000001]
  3a32d7:      	cmp	eax, 0x8
  3a32da:      	ja	0x3a3fd3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15e3>
  3a32e0:      	mov	ecx, 0x117
  3a32e5:      	bt	ecx, eax
  3a32e8:      	jae	0x3a3fc1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15d1>
  3a32ee:      	movzx	eax, r9b
  3a32f2:      	xor	edi, edi
  3a32f4:      	add	eax, -0x2
  3a32f7:      	cmp	eax, 0x7
  3a32fa:      	ja	0x3a3eef <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14ff>
  3a3300:      	mov	r9, qword ptr [rbp - 0x48]
  3a3304:      	jmp	qword ptr [8*rax + 0x203a68]
  3a330b:      	xor	edi, edi
  3a330d:      	test	r10d, r10d
  3a3310:      	sete	dil
  3a3314:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a331b:      	add	edi, dword ptr [r12 + 0x18]
  3a3320:      	cmp	rsi, rdi
  3a3323:      	ja	0x3a3782 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd92>
  3a3329:      	jmp	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a332e:      	mov	rax, rsi
  3a3331:      	mul	r9
  3a3334:      	cmp	dword ptr [r12], 0xfc0015
  3a333c:      	jne	0x3a3490 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa0>
  3a3342:      	mov	rdi, r9
  3a3345:      	sar	rdi, 0x3f
  3a3349:      	and	rdi, rsi
  3a334c:      	sar	rsi, 0x3f
  3a3350:      	and	rsi, r9
  3a3353:      	add	rsi, rdi
  3a3356:      	sub	rdx, rsi
  3a3359:      	jmp	0x3a3490 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa0>
  3a335e:      	cmp	r15d, 0xb
  3a3362:      	je	0x3a336d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x97d>
  3a3364:      	cmp	esi, 0x58
  3a3367:      	jne	0x3a371a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd2a>
  3a336d:      	mov	r10, qword ptr [rbp - 0x50]
  3a3371:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3375:      	mov	r9d, dword ptr [rdi + 0x18]
  3a3379:      	cmp	rsi, r9
  3a337c:      	mov	r13, qword ptr [rbp - 0x78]
  3a3380:      	jbe	0x3a4560 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b70>
  3a3386:      	shl	r9, 0x3
  3a338a:      	add	r9, qword ptr [r10 + 0x70]
  3a338e:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3394:      	mov	rsi, qword ptr [r9]
  3a3397:      	test	rsi, rsi
  3a339a:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a33a0:      	mov	rsi, qword ptr [rsi + 0x10]
  3a33a4:      	mov	r10d, dword ptr [rdi + 0x14]
  3a33a8:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a33ac:      	shl	r9, 0x20
  3a33b0:      	or	r9, r10
  3a33b3:      	movzx	edi, byte ptr [rdi + 0x40]
  3a33b7:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a33bc:      	je	0x3a33c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9d9>
  3a33be:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a33c3:      	je	0x3a3e1f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x142f>
  3a33c9:      	mov	qword ptr [rbp - 0x90], rax
  3a33d0:      	mov	qword ptr [rbp - 0x88], r8
  3a33d7:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a33de:      	movups	xmmword ptr [rsp], xmm0
  3a33e2:      	movzx	r8d, dil
  3a33e6:      	mov	rdi, rsi
  3a33e9:      	mov	esi, edx
  3a33eb:      	mov	rdx, rcx
  3a33ee:      	mov	rcx, r9
  3a33f1:      	call	0x3c8090 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a33f6:      	jmp	0x3a3e4c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x145c>
  3a33fb:      	mov	qword ptr [rbp - 0xa0], rax
  3a3402:      	mov	qword ptr [rbp - 0x98], r8
  3a3409:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a3410:      	movups	xmmword ptr [rsp], xmm0
  3a3414:      	movzx	r9d, dil
  3a3418:      	mov	rdi, r15
  3a341b:      	mov	r8, r10
  3a341e:      	call	0x3c6ce0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a3423:      	mov	r13, qword ptr [rbp - 0x78]
  3a3427:      	mov	rax, qword ptr [r15]
  3a342a:      	mov	rcx, qword ptr [r15 + 0x8]
  3a342e:      	mov	qword ptr [rbp - 0x68], r15
  3a3432:      	movzx	r15d, byte ptr [r15 + 0x10]
  3a3437:      	jmp	0x3a3e53 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1463>
  3a343c:      	mov	rax, qword ptr [r12 + 0x18]
  3a3441:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3446:      	xorps	xmm0, xmm0
  3a3449:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a3450:      	movups	xmmword ptr [rsp], xmm0
  3a3454:      	mov	r15, qword ptr [rbp - 0x68]
  3a3458:      	mov	rdi, r15
  3a345b:      	mov	edx, r9d
  3a345e:      	mov	rcx, r10
  3a3461:      	xor	r9d, r9d
  3a3464:      	call	0x3c6ce0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a3469:      	jmp	0x3a3427 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa37>
  3a346b:      	mov	rax, r9
  3a346e:      	sub	rax, rdi
  3a3471:      	not	r9
  3a3474:      	mov	r10, rdi
  3a3477:      	and	r10, r9
  3a347a:      	xor	r9, rdi
  3a347d:      	and	r9, rax
  3a3480:      	or	r9, r10
  3a3483:      	shr	r9, 0x3f
  3a3487:      	add	rdx, r9
  3a348a:      	sub	rsi, rdx
  3a348d:      	mov	rdx, rsi
  3a3490:      	mov	r10, qword ptr [rbp - 0x30]
  3a3494:      	mov	edi, dword ptr [r12 + 0x1c]
  3a3499:      	cmp	ecx, edi
  3a349b:      	jbe	0x3a459a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1baa>
  3a34a1:      	lea	esi, [rdi + 0x1]
  3a34a4:      	cmp	ecx, esi
  3a34a6:      	mov	r9, qword ptr [rbp - 0x60]
  3a34aa:      	jbe	0x3a45a2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bb2>
  3a34b0:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a34b4:      	cmp	r13, rcx
  3a34b7:      	jbe	0x3a45ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bbd>
  3a34bd:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a34c1:      	shl	rcx, 0x4
  3a34c5:      	mov	rsi, qword ptr [rbp - 0x48]
  3a34c9:      	mov	qword ptr [rsi + rcx], rax
  3a34cd:      	mov	qword ptr [rsi + rcx + 0x8], 0x0
  3a34d6:      	cmp	r13, rdi
  3a34d9:      	jbe	0x3a4558 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b68>
  3a34df:      	shl	rdi, 0x4
  3a34e3:      	mov	qword ptr [rsi + rdi], rdx
  3a34e7:      	mov	qword ptr [rsi + rdi + 0x8], 0x0
  3a34f0:      	inc	r14d
  3a34f3:      	mov	dword ptr [r9 - 0x38], r14d
  3a34f7:      	mov	r15d, r14d
  3a34fa:      	mov	r9, rsi
  3a34fd:      	mov	esi, 0xffffffff
  3a3502:      	mov	rcx, qword ptr [rbp - 0x58]
  3a3506:      	jmp	0x3a2aed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a350b:      	xor	r8d, r8d
  3a350e:      	mov	edi, dword ptr [r12 + 0x10]
  3a3513:      	cmp	rdi, rax
  3a3516:      	jne	0x3a2ef5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x505>
  3a351c:      	xor	eax, eax
  3a351e:      	xor	r9d, r9d
  3a3521:      	mov	r10d, 0x0
  3a3527:      	mov	r11d, 0x0
  3a352d:      	cmp	dx, 0xc
  3a3531:      	jne	0x3a3573 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb83>
  3a3533:      	mov	rcx, qword ptr [rbp - 0x30]
  3a3537:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a353e:      	mov	edi, dword ptr [r12 + 0x18]
  3a3543:      	add	edi, 0x2
  3a3546:      	cmp	rsi, rdi
  3a3549:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a354f:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a3556:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a3559:      	cmp	r13, rdi
  3a355c:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3562:      	shl	rdi, 0x4
  3a3566:      	mov	rcx, qword ptr [rbp - 0x48]
  3a356a:      	mov	r10, qword ptr [rcx + rdi]
  3a356e:      	mov	r11, qword ptr [rcx + rdi + 0x8]
  3a3573:      	cmp	qword ptr [rbp - 0x38], 0x0
  3a3578:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a357e:      	mov	rdx, qword ptr [rbp - 0x30]
  3a3582:      	mov	rsi, qword ptr [rdx + 0x80]
  3a3589:      	mov	edi, dword ptr [r12 + 0x4]
  3a358e:      	cmp	rsi, rdi
  3a3591:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3597:      	imul	rcx, rdi, 0xe0
  3a359e:      	add	rcx, qword ptr [rdx + 0x78]
  3a35a2:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a35a8:      	mov	rdx, qword ptr [rbp - 0x38]
  3a35ac:      	mov	rsi, qword ptr [rdx]
  3a35af:      	mov	qword ptr [rbp - 0xd0], rax
  3a35b6:      	mov	qword ptr [rbp - 0xc8], r9
  3a35bd:      	mov	qword ptr [rbp - 0xc0], r10
  3a35c4:      	mov	qword ptr [rbp - 0xb8], r11
  3a35cb:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a35d2:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a35d7:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a35de:      	movups	xmmword ptr [rsp], xmm0
  3a35e2:      	lea	rdi, [rbp - 0x160]
  3a35e9:      	mov	rdx, qword ptr [rbp - 0x50]
  3a35ed:      	call	0x3b69a0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a35f2:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a35f9:      	mov	r15, qword ptr [rbp - 0x150]
  3a3600:      	test	r15, r15
  3a3603:      	je	0x3a36dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcec>
  3a3609:      	cmp	qword ptr [r15 + 0x8], 0x0
  3a360e:      	mov	esi, 0xffffffff
  3a3613:      	mov	r10, qword ptr [rbp - 0x30]
  3a3617:      	mov	r9, qword ptr [rbp - 0x48]
  3a361b:      	mov	rcx, qword ptr [rbp - 0x58]
  3a361f:      	je	0x3a36ed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcfd>
  3a3625:      	jmp	0x3a441c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a2c>
  3a362a:      	cmp	r13, rdi
  3a362d:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3633:      	mov	edx, dword ptr [r12 + 0x18]
  3a3638:      	cmp	r13, rdx
  3a363b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a363f:      	jbe	0x3a462d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3d>
  3a3645:      	shl	rdi, 0x4
  3a3649:      	mov	rax, qword ptr [rsi + rdi]
  3a364d:      	mov	rcx, qword ptr [rsi + rdi + 0x8]
  3a3652:      	shl	rdx, 0x4
  3a3656:      	mov	qword ptr [rsi + rdx], rax
  3a365a:      	mov	qword ptr [rsi + rdx + 0x8], rcx
  3a365f:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3664:      	mov	eax, dword ptr [r12 + 0x14]
  3a3669:      	cmp	rax, rsi
  3a366c:      	je	0x3a3e96 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14a6>
  3a3672:      	cmp	r13, rax
  3a3675:      	mov	rdx, qword ptr [rbp - 0x48]
  3a3679:      	jbe	0x3a4542 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b52>
  3a367f:      	cmp	r13, rcx
  3a3682:      	jbe	0x3a4638 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c48>
  3a3688:      	shl	rax, 0x4
  3a368c:      	cmp	dword ptr [rdx + rax], 0x0
  3a3690:      	je	0x3a3ea3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14b3>
  3a3696:      	cmp	r13, rdi
  3a3699:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a369f:      	shl	rdi, 0x4
  3a36a3:      	mov	rax, qword ptr [rdx + rdi]
  3a36a7:      	mov	rcx, qword ptr [rdx + rdi + 0x8]
  3a36ac:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a36b1:      	mov	qword ptr [rbp - 0x90], rdx
  3a36b8:      	mov	qword ptr [rbp - 0x88], 0x0
  3a36c3:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a36ca:      	movups	xmmword ptr [rsp], xmm0
  3a36ce:      	mov	esi, r9d
  3a36d1:      	mov	rdx, r10
  3a36d4:      	xor	r8d, r8d
  3a36d7:      	jmp	0x3a3e47 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1457>
  3a36dc:      	mov	esi, 0xffffffff
  3a36e1:      	mov	r10, qword ptr [rbp - 0x30]
  3a36e5:      	mov	r9, qword ptr [rbp - 0x48]
  3a36e9:      	mov	rcx, qword ptr [rbp - 0x58]
  3a36ed:      	mov	edi, dword ptr [r12 + 0x8]
  3a36f2:      	cmp	rdi, rsi
  3a36f5:      	mov	r11, qword ptr [rbp - 0x38]
  3a36f9:      	mov	r15, qword ptr [rbp - 0x60]
  3a36fd:      	je	0x3a30c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6d7>
  3a3703:      	cmp	r13, rdi
  3a3706:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a370c:      	shl	rdi, 0x4
  3a3710:      	movups	xmmword ptr [r9 + rdi], xmm0
  3a3715:      	jmp	0x3a30c7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6d7>
  3a371a:      	mov	qword ptr [rsp], r11
  3a371e:      	mov	rsi, rcx
  3a3721:      	mov	rdx, r10
  3a3724:      	mov	rcx, rax
  3a3727:      	call	0x3cec40 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a372c:      	mov	rcx, rdx
  3a372f:      	mov	r11, qword ptr [rbp - 0x38]
  3a3733:      	mov	r13, qword ptr [rbp - 0x78]
  3a3737:      	mov	r8, qword ptr [rbp - 0x30]
  3a373b:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3740:      	mov	edi, dword ptr [r12 + 0x1c]
  3a3745:      	dec	edi
  3a3747:      	cmp	edi, r10d
  3a374a:      	cmovae	edi, r10d
  3a374e:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a3755:      	add	edi, dword ptr [r12 + 0x18]
  3a375a:      	cmp	rsi, rdi
  3a375d:      	ja	0x3a3782 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd92>
  3a375f:      	jmp	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3764:      	xor	edi, edi
  3a3766:      	test	r10, r10
  3a3769:      	setne	dil
  3a376d:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a3774:      	add	edi, dword ptr [r12 + 0x18]
  3a3779:      	cmp	rsi, rdi
  3a377c:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3782:      	mov	r10, r8
  3a3785:      	mov	rax, qword ptr [r8 + 0xe0]
  3a378c:      	lea	rcx, [rdi + 2*rdi]
  3a3790:      	mov	r15d, dword ptr [rax + 4*rcx]
  3a3794:      	mov	r12d, dword ptr [rax + 4*rcx + 0x4]
  3a3799:      	mov	esi, dword ptr [rax + 4*rcx + 0x8]
  3a379d:      	cmp	r15d, r14d
  3a37a0:      	ja	0x3a3f2e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x153e>
  3a37a6:      	mov	rax, qword ptr [r11 + 0x160]
  3a37ad:      	test	rax, rax
  3a37b0:      	je	0x3a37c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd5>
  3a37b2:      	test	r11, r11
  3a37b5:      	je	0x3a3f0d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x151d>
  3a37bb:      	mov	eax, dword ptr [rax]
  3a37bd:      	test	eax, eax
  3a37bf:      	je	0x3a4207 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1817>
  3a37c5:      	mov	rax, qword ptr [r11 + 0x8]
  3a37c9:      	test	rax, rax
  3a37cc:      	je	0x3a464c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a37d2:      	mov	rdi, qword ptr [r11 + 0x10]
  3a37d6:      	mov	rcx, qword ptr [rax + 0x8]
  3a37da:      	test	byte ptr [rcx + 0x17], 0x20
  3a37de:      	jne	0x3a3eb2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c2>
  3a37e4:      	mov	r14, rsi
  3a37e7:      	jmp	0x3a3ecc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dc>
  3a37ec:      	xor	eax, eax
  3a37ee:      	cmp	r10, rdx
  3a37f1:      	sete	al
  3a37f4:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a37f9:      	mov	rax, qword ptr [rbp - 0x50]
  3a37fd:      	mov	rsi, qword ptr [rax + 0x78]
  3a3801:      	mov	edi, dword ptr [r12 + 0x18]
  3a3806:      	cmp	rsi, rdi
  3a3809:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a380f:      	shl	rdi, 0x3
  3a3813:      	add	rdi, qword ptr [rax + 0x70]
  3a3817:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a381d:      	test	r11, r11
  3a3820:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3826:      	mov	r15, qword ptr [rdi]
  3a3829:      	test	r15, r15
  3a382c:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3832:      	mov	rdi, qword ptr [r11]
  3a3835:      	mov	rax, qword ptr [r15 + 0x10]
  3a3839:      	movzx	eax, byte ptr [rax + 0x31]
  3a383d:      	mov	byte ptr [rbp - 0x69], al
  3a3840:      	mov	edx, r10d
  3a3843:      	test	al, 0x1
  3a3845:      	cmovne	rdx, r10
  3a3849:      	mov	rsi, r15
  3a384c:      	call	0x3c0d10 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a3851:      	test	dl, 0x1
  3a3854:      	jne	0x3a2fac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5bc>
  3a385a:      	mov	rax, qword ptr [r15 + 0x10]
  3a385e:      	movzx	eax, byte ptr [rax + 0x31]
  3a3862:      	mov	byte ptr [rbp - 0x6a], al
  3a3865:      	and	al, 0x1
  3a3867:      	neg	al
  3a3869:      	mov	eax, 0x0
  3a386e:      	sbb	rax, rax
  3a3871:      	mov	ecx, 0xffffffff
  3a3876:      	or	rax, rcx
  3a3879:      	jmp	0x3a2fac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5bc>
  3a387e:      	mov	rax, qword ptr [rbp - 0x50]
  3a3882:      	mov	rsi, qword ptr [rax + 0x60]
  3a3886:      	mov	edi, dword ptr [r12 + 0x18]
  3a388b:      	cmp	rsi, rdi
  3a388e:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3894:      	mov	rax, qword ptr [rax + 0x58]
  3a3898:      	lea	rcx, [rax + 8*rdi]
  3a389c:      	test	rcx, rcx
  3a389f:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a38a5:      	mov	rax, qword ptr [rax + 8*rdi]
  3a38a9:      	test	rax, rax
  3a38ac:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a38b2:      	cmp	byte ptr [rax + 0x29], 0x0
  3a38b6:      	mov	ecx, r10d
  3a38b9:      	cmovne	rcx, r10
  3a38bd:      	cmp	rcx, qword ptr [rax + 0x40]
  3a38c1:      	jae	0x3a4474 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a84>
  3a38c7:      	mov	rax, qword ptr [rax + 0x38]
  3a38cb:      	mov	rax, qword ptr [rax + 8*rcx]
  3a38cf:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a38d4:      	mov	rax, qword ptr [rbp - 0x50]
  3a38d8:      	mov	rsi, qword ptr [rax + 0x48]
  3a38dc:      	mov	edi, dword ptr [r12 + 0x18]
  3a38e1:      	cmp	rsi, rdi
  3a38e4:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a38ea:      	mov	rax, qword ptr [rax + 0x40]
  3a38ee:      	mov	rax, qword ptr [rax + 8*rdi]
  3a38f2:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a38f7:      	test	r10, r10
  3a38fa:      	jne	0x3a39dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfec>
  3a3900:      	jmp	0x3a447e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8e>
  3a3905:      	mov	rax, qword ptr [rbp - 0x50]
  3a3909:      	mov	rsi, qword ptr [rax + 0x90]
  3a3910:      	mov	eax, dword ptr [r12 + 0x18]
  3a3915:      	cmp	rsi, rax
  3a3918:      	jbe	0x3a4644 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c54>
  3a391e:      	cmp	r13, rdi
  3a3921:      	mov	rdx, qword ptr [rbp - 0x48]
  3a3925:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a392b:      	mov	rcx, qword ptr [rbp - 0x50]
  3a392f:      	mov	rcx, qword ptr [rcx + 0x88]
  3a3936:      	mov	rax, qword ptr [rcx + 8*rax]
  3a393a:      	shl	rdi, 0x4
  3a393e:      	movups	xmm0, xmmword ptr [rdx + rdi]
  3a3942:      	movups	xmmword ptr [rax + 0x20], xmm0
  3a3946:      	jmp	0x3a39cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfdd>
  3a394b:      	mov	rax, qword ptr [rbp - 0x50]
  3a394f:      	mov	rsi, qword ptr [rax + 0x90]
  3a3956:      	mov	edi, dword ptr [r12 + 0x18]
  3a395b:      	cmp	rsi, rdi
  3a395e:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3964:      	mov	rax, qword ptr [rax + 0x88]
  3a396b:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a396f:      	mov	rax, qword ptr [rcx + 0x20]
  3a3973:      	mov	rcx, qword ptr [rcx + 0x28]
  3a3977:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a397c:      	mov	rax, qword ptr [rbp - 0x50]
  3a3980:      	mov	rsi, qword ptr [rax + 0x60]
  3a3984:      	mov	edi, dword ptr [r12 + 0x18]
  3a3989:      	cmp	rsi, rdi
  3a398c:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3992:      	mov	rax, qword ptr [rax + 0x58]
  3a3996:      	lea	rcx, [rax + 8*rdi]
  3a399a:      	test	rcx, rcx
  3a399d:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a39a3:      	mov	rax, qword ptr [rax + 8*rdi]
  3a39a7:      	test	rax, rax
  3a39aa:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a39b0:      	cmp	byte ptr [rax + 0x29], 0x0
  3a39b4:      	mov	ecx, r10d
  3a39b7:      	cmovne	rcx, r10
  3a39bb:      	cmp	rcx, qword ptr [rax + 0x40]
  3a39bf:      	jae	0x3a4474 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a84>
  3a39c5:      	mov	rax, qword ptr [rax + 0x38]
  3a39c9:      	mov	qword ptr [rax + 8*rcx], rdx
  3a39cd:      	xor	eax, eax
  3a39cf:      	nop
  3a39d0:      	xor	ecx, ecx
  3a39d2:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a39d7:      	mov	edi, dword ptr [r12 + 0x18]
  3a39dc:      	cmp	r13, rdi
  3a39df:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a39e5:      	shl	rdi, 0x4
  3a39e9:      	mov	rcx, qword ptr [rbp - 0x48]
  3a39ed:      	mov	rax, qword ptr [rcx + rdi]
  3a39f1:      	mov	rcx, qword ptr [rcx + rdi + 0x8]
  3a39f6:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a39fb:      	xor	eax, eax
  3a39fd:      	test	r10, r10
  3a3a00:      	sete	al
  3a3a03:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a05:      	mov	rax, qword ptr [rbp - 0x50]
  3a3a09:      	mov	rsi, qword ptr [rax + 0x78]
  3a3a0d:      	mov	edi, dword ptr [r12 + 0x18]
  3a3a12:      	cmp	rsi, rdi
  3a3a15:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3a1b:      	shl	rdi, 0x3
  3a3a1f:      	add	rdi, qword ptr [rax + 0x70]
  3a3a23:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3a29:      	mov	rax, qword ptr [rdi]
  3a3a2c:      	test	rax, rax
  3a3a2f:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3a35:      	mov	rdi, qword ptr [rax + 0x10]
  3a3a39:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a3a3e:      	mov	rax, qword ptr [rdi + 0x60]
  3a3a42:      	je	0x3a41c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17d8>
  3a3a48:      	test	rax, rax
  3a3a4b:      	jne	0x3a41d1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17e1>
  3a3a51:      	call	0x3c47a0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a3a56:      	mov	r8, qword ptr [rbp - 0x30]
  3a3a5a:      	mov	r11, qword ptr [rbp - 0x38]
  3a3a5e:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a63:      	mov	eax, dword ptr [r12 + 0x18]
  3a3a68:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a6d:      	xor	rdx, r10
  3a3a70:      	jmp	0x3a3dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3a75:      	mov	eax, r10d
  3a3a78:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a7d:      	mov	ecx, edx
  3a3a7f:      	shr	r10, cl
  3a3a82:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3a87:      	mov	ecx, edx
  3a3a89:      	shr	r10d, cl
  3a3a8c:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3a91:      	or	rdx, r10
  3a3a94:      	jmp	0x3a3dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3a99:      	mov	ecx, edx
  3a3a9b:      	sar	r10d, cl
  3a3a9e:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3aa3:      	xor	edx, r10d
  3a3aa6:      	jmp	0x3a3dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3aab:      	mov	rax, r10
  3a3aae:      	shr	rax
  3a3ab1:      	movabs	rcx, 0x5555555555555555
  3a3abb:      	and	rax, rcx
  3a3abe:      	and	r10, rcx
  3a3ac1:      	add	r10, rax
  3a3ac4:      	mov	rax, r10
  3a3ac7:      	shr	rax, 0x2
  3a3acb:      	movabs	rcx, 0x3333333333333333
  3a3ad5:      	and	rax, rcx
  3a3ad8:      	and	r10, rcx
  3a3adb:      	add	r10, rax
  3a3ade:      	mov	rax, r10
  3a3ae1:      	shr	rax, 0x4
  3a3ae5:      	add	rax, r10
  3a3ae8:      	movabs	rcx, 0xf0f0f0f0f0f0f0f
  3a3af2:      	and	rcx, rax
  3a3af5:      	mov	rax, rcx
  3a3af8:      	shr	rax, 0x8
  3a3afc:      	add	rax, rcx
  3a3aff:      	mov	rcx, rax
  3a3b02:      	shr	rcx, 0x10
  3a3b06:      	add	rcx, rax
  3a3b09:      	mov	rax, rcx
  3a3b0c:      	shr	rax, 0x20
  3a3b10:      	add	eax, ecx
  3a3b12:      	and	eax, 0x7f
  3a3b15:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b1a:      	xor	eax, eax
  3a3b1c:      	cmp	r10, rdx
  3a3b1f:      	setb	al
  3a3b22:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b27:      	xor	eax, eax
  3a3b29:      	cmp	r10, rdx
  3a3b2c:      	setbe	al
  3a3b2f:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b34:      	xor	eax, eax
  3a3b36:      	cmp	r10, rdx
  3a3b39:      	seta	al
  3a3b3c:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b41:      	imul	edx, r10d
  3a3b45:      	jmp	0x3a3dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3b4a:      	mov	ecx, r10d
  3a3b4d:      	shr	ecx, 0x10
  3a3b50:      	xor	eax, eax
  3a3b52:      	cmp	r10d, 0x10000
  3a3b59:      	setb	al
  3a3b5c:      	cmovb	ecx, r10d
  3a3b60:      	shl	eax, 0x4
  3a3b63:      	mov	edx, ecx
  3a3b65:      	shr	edx, 0x8
  3a3b68:      	xor	esi, esi
  3a3b6a:      	cmp	ecx, 0x100
  3a3b70:      	setb	sil
  3a3b74:      	lea	eax, [rax + 8*rsi]
  3a3b77:      	cmovb	edx, ecx
  3a3b7a:      	movzx	ecx, dl
  3a3b7d:      	movzx	ecx, byte ptr [rcx + 0x3284b9]
  3a3b84:      	sub	rax, rcx
  3a3b87:      	add	rax, 0x8
  3a3b8b:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b90:      	mov	ecx, edx
  3a3b92:      	rol	r10d, cl
  3a3b95:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3b9a:      	xor	eax, eax
  3a3b9c:      	cmp	r10, rdx
  3a3b9f:      	setg	al
  3a3ba2:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3ba7:      	xor	eax, eax
  3a3ba9:      	cmp	r10d, edx
  3a3bac:      	setl	al
  3a3baf:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3bb4:      	xor	eax, eax
  3a3bb6:      	cmp	r10, rdx
  3a3bb9:      	setne	al
  3a3bbc:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3bc1:      	test	r10d, r10d
  3a3bc4:      	je	0x3a4219 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1829>
  3a3bca:      	tzcnt	eax, r10d
  3a3bcf:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3bd4:      	add	edx, r10d
  3a3bd7:      	jmp	0x3a3dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3bdc:      	xor	eax, eax
  3a3bde:      	cmp	r10d, edx
  3a3be1:      	setne	al
  3a3be4:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3be9:      	xor	eax, eax
  3a3beb:      	cmp	r10, rdx
  3a3bee:      	setle	al
  3a3bf1:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3bf6:      	mov	ecx, edx
  3a3bf8:      	ror	r10d, cl
  3a3bfb:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3c00:      	xor	eax, eax
  3a3c02:      	cmp	r10d, edx
  3a3c05:      	setbe	al
  3a3c08:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3c0d:      	xor	eax, eax
  3a3c0f:      	cmp	r10d, edx
  3a3c12:      	setge	al
  3a3c15:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3c1a:      	xor	eax, eax
  3a3c1c:      	test	r10d, r10d
  3a3c1f:      	sete	al
  3a3c22:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3c27:      	xor	eax, eax
  3a3c29:      	cmp	r10d, edx
  3a3c2c:      	setb	al
  3a3c2f:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3c34:      	mov	rdx, r10
  3a3c37:      	mov	eax, edx
  3a3c39:      	shr	eax, 0x18
  3a3c3c:      	mov	ecx, edx
  3a3c3e:      	shr	ecx, 0x10
  3a3c41:      	movzx	ecx, cl
  3a3c44:      	movzx	ecx, byte ptr [rcx + 0x3285b9]
  3a3c4b:      	add	cl, byte ptr [rax + 0x3285b9]
  3a3c51:      	movzx	eax, dh
  3a3c54:      	add	cl, byte ptr [rax + 0x3285b9]
  3a3c5a:      	movzx	eax, dl
  3a3c5d:      	add	cl, byte ptr [rax + 0x3285b9]
  3a3c63:      	movzx	eax, cl
  3a3c66:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3c6b:      	add	rdx, r10
  3a3c6e:      	jmp	0x3a3dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3c73:      	mov	ecx, edx
  3a3c75:      	shl	r10d, cl
  3a3c78:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3c7d:      	xor	eax, eax
  3a3c7f:      	test	r10, r10
  3a3c82:      	sete	al
  3a3c85:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3c8a:      	and	rdx, r10
  3a3c8d:      	jmp	0x3a3dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3c92:      	sub	r10, rdx
  3a3c95:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3c9a:      	xor	eax, eax
  3a3c9c:      	cmp	r10d, edx
  3a3c9f:      	sete	al
  3a3ca2:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3ca7:      	mov	ecx, edx
  3a3ca9:      	rol	r10, cl
  3a3cac:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3cb1:      	xor	eax, eax
  3a3cb3:      	cmp	r10d, edx
  3a3cb6:      	setae	al
  3a3cb9:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3cbe:      	and	r10d, edx
  3a3cc1:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3cc6:      	xor	eax, eax
  3a3cc8:      	cmp	r10d, edx
  3a3ccb:      	setle	al
  3a3cce:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3cd3:      	xor	eax, eax
  3a3cd5:      	cmp	r10d, edx
  3a3cd8:      	setg	al
  3a3cdb:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3ce0:      	xor	eax, eax
  3a3ce2:      	mov	rcx, r10
  3a3ce5:      	shr	rcx, 0x20
  3a3ce9:      	cmove	rcx, r10
  3a3ced:      	setne	al
  3a3cf0:      	shl	eax, 0x5
  3a3cf3:      	mov	rdx, rcx
  3a3cf6:      	shr	rdx, 0x10
  3a3cfa:      	lea	rsi, [rax + 0x10]
  3a3cfe:      	cmp	rcx, 0x10000
  3a3d05:      	cmovb	rdx, rcx
  3a3d09:      	cmovb	rsi, rax
  3a3d0d:      	mov	eax, edx
  3a3d0f:      	shr	eax, 0x8
  3a3d12:      	lea	rcx, [rsi + 0x8]
  3a3d16:      	cmp	rdx, 0x100
  3a3d1d:      	cmovb	rax, rdx
  3a3d21:      	cmovb	rcx, rsi
  3a3d25:      	movzx	eax, al
  3a3d28:      	movzx	edx, byte ptr [rax + 0x3284b9]
  3a3d2f:      	add	rdx, rcx
  3a3d32:      	mov	eax, 0x40
  3a3d37:      	sub	rax, rdx
  3a3d3a:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d3f:      	xor	eax, eax
  3a3d41:      	cmp	r10, rdx
  3a3d44:      	setl	al
  3a3d47:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d4c:      	mov	ecx, edx
  3a3d4e:      	ror	r10, cl
  3a3d51:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3d56:      	xor	eax, eax
  3a3d58:      	cmp	r10d, edx
  3a3d5b:      	seta	al
  3a3d5e:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d63:      	mov	eax, r10d
  3a3d66:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d6b:      	movsx	rax, r10w
  3a3d6f:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d74:      	movsxd	rax, r10d
  3a3d77:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d7c:      	xor	eax, eax
  3a3d7e:      	cmp	r10, rdx
  3a3d81:      	sete	al
  3a3d84:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d89:      	movsxd	rax, r10d
  3a3d8c:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d91:      	xor	eax, eax
  3a3d93:      	cmp	r10, rdx
  3a3d96:      	setae	al
  3a3d99:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3d9e:      	xor	eax, eax
  3a3da0:      	cmp	r10, rdx
  3a3da3:      	setge	al
  3a3da6:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3dab:      	mov	ecx, edx
  3a3dad:      	sar	r10, cl
  3a3db0:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3db2:      	or	edx, r10d
  3a3db5:      	jmp	0x3a3dc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3db7:      	sub	r10d, edx
  3a3dba:      	jmp	0x3a3de9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3dbc:      	imul	rdx, r10
  3a3dc0:      	xor	ecx, ecx
  3a3dc2:      	mov	rax, rdx
  3a3dc5:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3dca:      	movsx	rax, r10w
  3a3dce:      	mov	eax, eax
  3a3dd0:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3dd5:      	mov	eax, 0x40
  3a3dda:      	tzcnt	rax, r10
  3a3ddf:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3de4:      	mov	ecx, edx
  3a3de6:      	shl	r10, cl
  3a3de9:      	xor	ecx, ecx
  3a3deb:      	mov	rax, r10
  3a3dee:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3df0:      	movsx	rax, r10b
  3a3df4:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3df9:      	movsx	rax, r10b
  3a3dfd:      	mov	eax, eax
  3a3dff:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3e04:      	movzx	esi, r9b
  3a3e08:      	add	esi, -0x61
  3a3e0b:      	cmp	esi, 0x5e
  3a3e0e:      	ja	0x3a39cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfdd>
  3a3e14:      	xor	eax, eax
  3a3e16:      	xor	ecx, ecx
  3a3e18:      	jmp	qword ptr [8*rsi + 0x203ea8]
  3a3e1f:      	mov	qword ptr [rbp - 0x90], rax
  3a3e26:      	mov	qword ptr [rbp - 0x88], r8
  3a3e2d:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a3e34:      	movups	xmmword ptr [rsp], xmm0
  3a3e38:      	movzx	r8d, dil
  3a3e3c:      	mov	rdi, rsi
  3a3e3f:      	mov	esi, edx
  3a3e41:      	mov	rdx, rcx
  3a3e44:      	mov	rcx, r9
  3a3e47:      	call	0x3c8370 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a3e4c:      	mov	r15d, eax
  3a3e4f:      	xor	eax, eax
  3a3e51:      	xor	ecx, ecx
  3a3e53:      	mov	r11, qword ptr [rbp - 0x38]
  3a3e57:      	mov	r8, qword ptr [rbp - 0x30]
  3a3e5b:      	test	r15b, r15b
  3a3e5e:      	jne	0x3a4481 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3e64:      	mov	edi, dword ptr [r12 + 0x8]
  3a3e69:      	mov	esi, 0xffffffff
  3a3e6e:      	cmp	rdi, rsi
  3a3e71:      	mov	r9, qword ptr [rbp - 0x48]
  3a3e75:      	je	0x3a2ae0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0>
  3a3e7b:      	cmp	r13, rdi
  3a3e7e:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3e84:      	shl	rdi, 0x4
  3a3e88:      	mov	qword ptr [r9 + rdi], rax
  3a3e8c:      	mov	qword ptr [r9 + rdi + 0x8], rcx
  3a3e91:      	jmp	0x3a2ae0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0>
  3a3e96:      	cmp	r13, rcx
  3a3e99:      	mov	rdx, qword ptr [rbp - 0x48]
  3a3e9d:      	jbe	0x3a4638 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c48>
  3a3ea3:      	shl	rcx, 0x4
  3a3ea7:      	mov	rax, qword ptr [rdx + rcx]
  3a3eab:      	mov	rcx, qword ptr [rdx + rcx + 0x8]
  3a3eb0:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3eb2:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a3eb6:      	and	cl, 0x1f
  3a3eb9:      	mov	r14, rsi
  3a3ebc:      	cmp	cl, 0x16
  3a3ebf:      	je	0x3a3ecc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dc>
  3a3ec1:      	mov	qword ptr [rbp - 0x90], rdi
  3a3ec8:      	mov	rdi, qword ptr [rbp - 0x68]
  3a3ecc:      	call	qword ptr [rax + 0x28]
  3a3ecf:      	mov	r10, qword ptr [rbp - 0x30]
  3a3ed3:      	mov	r9, qword ptr [rbp - 0x48]
  3a3ed7:      	mov	rsi, r14
  3a3eda:      	test	rax, rax
  3a3edd:      	je	0x3a3f2a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x153a>
  3a3edf:      	cmp	qword ptr [rax + 0x8], 0x0
  3a3ee4:      	mov	r11, qword ptr [rbp - 0x38]
  3a3ee8:      	je	0x3a3f2e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x153e>
  3a3eea:      	jmp	0x3a440c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a1c>
  3a3eef:      	mov	r9, qword ptr [rbp - 0x48]
  3a3ef3:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a3efa:      	add	edi, dword ptr [r12 + 0x18]
  3a3eff:      	cmp	rsi, rdi
  3a3f02:      	ja	0x3a3782 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd92>
  3a3f08:      	jmp	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3f0d:      	mov	edi, 0x1
  3a3f12:      	mov	r14, rsi
  3a3f15:      	call	0x5e6250 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a3f1a:      	mov	rsi, r14
  3a3f1d:      	mov	r9, qword ptr [rbp - 0x48]
  3a3f21:      	mov	r10, qword ptr [rbp - 0x30]
  3a3f25:      	test	rax, rax
  3a3f28:      	jne	0x3a3edf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14ef>
  3a3f2a:      	mov	r11, qword ptr [rbp - 0x38]
  3a3f2e:      	mov	rax, qword ptr [r10 + 0x108]
  3a3f35:      	lea	edx, [rsi + r12]
  3a3f39:      	cmp	rax, rdx
  3a3f3c:      	jb	0x3a461b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c2b>
  3a3f42:      	cmp	r12d, edx
  3a3f45:      	ja	0x3a460c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c1c>
  3a3f4b:      	sub	rax, r12
  3a3f4e:      	lea	rcx, [8*r12]
  3a3f56:      	test	rax, rax
  3a3f59:      	mov	eax, 0x0
  3a3f5e:      	cmovle	rcx, rax
  3a3f62:      	test	esi, esi
  3a3f64:      	je	0x3a3fb3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15c3>
  3a3f66:      	mov	rax, qword ptr [r10 + 0xf8]
  3a3f6d:      	sub	rdx, r12
  3a3f70:      	add	rcx, rax
  3a3f73:      	add	rcx, 0x4
  3a3f77:      	xor	esi, esi
  3a3f79:      	nop	dword ptr [rax]
  3a3f80:      	mov	edi, dword ptr [rcx + 8*rsi]
  3a3f83:      	cmp	r13, rdi
  3a3f86:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3f8c:      	mov	eax, dword ptr [rcx + 8*rsi - 0x4]
  3a3f90:      	cmp	r13, rax
  3a3f93:      	jbe	0x3a4542 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b52>
  3a3f99:      	shl	rdi, 0x4
  3a3f9d:      	shl	rax, 0x4
  3a3fa1:      	movups	xmm0, xmmword ptr [r9 + rdi]
  3a3fa6:      	movups	xmmword ptr [r9 + rax], xmm0
  3a3fab:      	inc	rsi
  3a3fae:      	cmp	rdx, rsi
  3a3fb1:      	jne	0x3a3f80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1590>
  3a3fb3:      	mov	rcx, qword ptr [rbp - 0x58]
  3a3fb7:      	mov	esi, 0xffffffff
  3a3fbc:      	jmp	0x3a2aed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a3fc1:      	cmp	eax, 0x3
  3a3fc4:      	je	0x3a438f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x199f>
  3a3fca:      	cmp	eax, 0x6
  3a3fcd:      	je	0x3a4335 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1945>
  3a3fd3:      	lea	eax, [r9 - 0x12]
  3a3fd7:      	cmp	eax, 0x4
  3a3fda:      	jb	0x3a4365 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1975>
  3a3fe0:      	mov	edi, dword ptr [r12 + 0x14]
  3a3fe5:      	cmp	rdi, rsi
  3a3fe8:      	je	0x3a404b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x165b>
  3a3fea:      	cmp	r13, rdi
  3a3fed:      	jbe	0x3a452a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3ff3:      	shl	rdi, 0x4
  3a3ff7:      	mov	rax, qword ptr [rbp - 0x48]
  3a3ffb:      	mov	r15, qword ptr [rax + rdi]
  3a3fff:      	mov	eax, r9d
  3a4002:      	and	eax, 0xffff0000
  3a4007:      	cmp	eax, 0xfe0000
  3a400c:      	jne	0x3a405d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x166d>
  3a400e:      	movzx	r9d, r9w
  3a4012:      	cmp	r9d, 0x3
  3a4016:      	jne	0x3a4223 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1833>
  3a401c:      	test	r11, r11
  3a401f:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a4025:      	mov	rax, qword ptr [r11]
  3a4028:      	test	rax, rax
  3a402b:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a4031:      	mov	rdi, qword ptr [rax + 0x8]
  3a4035:      	call	0x3c3d40 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a403a:      	mov	r11, qword ptr [rbp - 0x38]
  3a403e:      	xor	eax, eax
  3a4040:      	xor	ecx, ecx
  3a4042:      	mov	r8, qword ptr [rbp - 0x30]
  3a4046:      	jmp	0x3a3e64 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a404b:      	xor	r15d, r15d
  3a404e:      	mov	eax, r9d
  3a4051:      	and	eax, 0xffff0000
  3a4056:      	cmp	eax, 0xfe0000
  3a405b:      	je	0x3a400e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x161e>
  3a405d:      	add	r9d, 0xff03fff8
  3a4064:      	cmp	r9d, 0xa
  3a4068:      	jae	0x3a466a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7a>
  3a406e:      	mov	rcx, qword ptr [rbp - 0x30]
  3a4072:      	mov	rsi, qword ptr [rcx + 0x80]
  3a4079:      	mov	edi, dword ptr [r12 + 0x4]
  3a407e:      	cmp	rsi, rdi
  3a4081:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a4087:      	imul	rax, rdi, 0xe0
  3a408e:      	add	rax, qword ptr [rcx + 0x78]
  3a4092:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a4098:      	movups	xmm0, xmmword ptr [rax + 0x98]
  3a409f:      	movups	xmm1, xmmword ptr [rax + 0x88]
  3a40a6:      	movups	xmm2, xmmword ptr [rax + 0x78]
  3a40aa:      	movups	xmm3, xmmword ptr [rax + 0x68]
  3a40ae:      	movups	xmm4, xmmword ptr [rax + 0x58]
  3a40b2:      	movups	xmm5, xmmword ptr [rax + 0x48]
  3a40b6:      	movzx	ecx, byte ptr [rax + 0x40]
  3a40ba:      	movups	xmm6, xmmword ptr [rax + 0x30]
  3a40be:      	movups	xmm7, xmmword ptr [rax + 0x20]
  3a40c2:      	movups	xmm8, xmmword ptr [rax + 0x10]
  3a40c7:      	mov	rsi, qword ptr [rax + 0x8]
  3a40cb:      	mov	eax, dword ptr [rax]
  3a40cd:      	mov	dword ptr [rbp - 0x208], eax
  3a40d3:      	mov	qword ptr [rbp - 0x200], rsi
  3a40da:      	movups	xmmword ptr [rbp - 0x1f8], xmm8
  3a40e2:      	movups	xmmword ptr [rbp - 0x1e8], xmm7
  3a40e9:      	movups	xmmword ptr [rbp - 0x1d8], xmm6
  3a40f0:      	mov	byte ptr [rbp - 0x1c8], cl
  3a40f6:      	movups	xmmword ptr [rbp - 0x1c0], xmm5
  3a40fd:      	movups	xmmword ptr [rbp - 0x1b0], xmm4
  3a4104:      	movups	xmmword ptr [rbp - 0x1a0], xmm3
  3a410b:      	movups	xmmword ptr [rbp - 0x190], xmm2
  3a4112:      	movups	xmmword ptr [rbp - 0x180], xmm1
  3a4119:      	movups	xmmword ptr [rbp - 0x170], xmm0
  3a4120:      	mov	qword ptr [rbp - 0xf0], rdx
  3a4127:      	mov	qword ptr [rbp - 0xe8], 0x0
  3a4132:      	mov	qword ptr [rbp - 0xe0], r15
  3a4139:      	mov	qword ptr [rbp - 0xd8], 0x0
  3a4144:      	mov	ecx, 0x15
  3a4149:      	mov	rdi, rsp
  3a414c:      	lea	rsi, [rbp - 0x208]
  3a4153:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a4156:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a415d:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a4165:      	movups	xmm0, xmmword ptr [rbp - 0xf0]
  3a416c:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a4174:      	lea	rdi, [rbp - 0x118]
  3a417b:      	mov	rsi, r11
  3a417e:      	mov	rcx, r10
  3a4181:      	mov	rdx, qword ptr [rbp - 0x50]
  3a4185:      	call	0x3a9610 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a418a:      	mov	rax, qword ptr [rbp - 0x118]
  3a4191:      	mov	rcx, qword ptr [rbp - 0x110]
  3a4198:      	movzx	r15d, byte ptr [rbp - 0x108]
  3a41a0:      	mov	rsi, qword ptr [rbp - 0x100]
  3a41a7:      	test	rsi, rsi
  3a41aa:      	je	0x3a42fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x190e>
  3a41b0:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a41b5:      	mov	r11, qword ptr [rbp - 0x38]
  3a41b9:      	mov	r8, qword ptr [rbp - 0x30]
  3a41bd:      	je	0x3a3e5b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x146b>
  3a41c3:      	jmp	0x3a450a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b1a>
  3a41c8:      	test	rax, rax
  3a41cb:      	je	0x3a430c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x191c>
  3a41d1:      	mov	rax, qword ptr [rax]
  3a41d4:      	test	rdi, rdi
  3a41d7:      	je	0x3a4319 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1929>
  3a41dd:      	mov	rcx, qword ptr [rdi + 0x38]
  3a41e1:      	test	rcx, rcx
  3a41e4:      	mov	edx, 0x10000
  3a41e9:      	cmove	rcx, rdx
  3a41ed:      	mov	rdx, rax
  3a41f0:      	or	rdx, rcx
  3a41f3:      	shr	rdx, 0x20
  3a41f7:      	je	0x3a4303 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1913>
  3a41fd:      	xor	edx, edx
  3a41ff:      	div	rcx
  3a4202:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a4207:      	xor	eax, eax
  3a4209:      	xor	edx, edx
  3a420b:      	test	rax, rax
  3a420e:      	jne	0x3a3edf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14ef>
  3a4214:      	jmp	0x3a3f2a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x153a>
  3a4219:      	mov	eax, 0x20
  3a421e:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a4223:      	mov	rax, qword ptr [rbp - 0x30]
  3a4227:      	mov	rsi, qword ptr [rax + 0x80]
  3a422e:      	mov	edi, dword ptr [r12 + 0x4]
  3a4233:      	cmp	rsi, rdi
  3a4236:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a423c:      	imul	r11, rdi, 0xe0
  3a4243:      	add	r11, qword ptr [rax + 0x78]
  3a4247:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a424d:      	mov	rax, qword ptr [rbp - 0x50]
  3a4251:      	mov	rsi, qword ptr [rax + 0x78]
  3a4255:      	mov	edi, dword ptr [r11 + 0x18]
  3a4259:      	cmp	rsi, rdi
  3a425c:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a4262:      	shl	rdi, 0x3
  3a4266:      	add	rdi, qword ptr [rax + 0x70]
  3a426a:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a4270:      	mov	rsi, qword ptr [rdi]
  3a4273:      	test	rsi, rsi
  3a4276:      	je	0x3a43ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a427c:      	mov	rdi, qword ptr [rbp - 0x38]
  3a4280:      	mov	rcx, qword ptr [rdi]
  3a4283:      	mov	rax, qword ptr [rdi + 0x8]
  3a4287:      	mov	r8, qword ptr [rcx + 0x8]
  3a428b:      	mov	rcx, qword ptr [rdi + 0x10]
  3a428f:      	mov	rsi, qword ptr [rsi + 0x10]
  3a4293:      	mov	edi, dword ptr [r11 + 0x14]
  3a4297:      	mov	r11d, dword ptr [r11 + 0x1c]
  3a429b:      	shl	r11, 0x20
  3a429f:      	or	r11, rdi
  3a42a2:      	mov	qword ptr [rsp + 0x18], r15
  3a42a7:      	mov	qword ptr [rsp + 0x10], rdx
  3a42ac:      	mov	qword ptr [rsp + 0x8], r11
  3a42b1:      	mov	qword ptr [rsp], r10
  3a42b5:      	lea	rdi, [rbp - 0x140]
  3a42bc:      	mov	rdx, rax
  3a42bf:      	call	0x3c5050 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a42c4:      	mov	rax, qword ptr [rbp - 0x140]
  3a42cb:      	mov	rcx, qword ptr [rbp - 0x138]
  3a42d2:      	movzx	r15d, byte ptr [rbp - 0x130]
  3a42da:      	mov	rsi, qword ptr [rbp - 0x128]
  3a42e1:      	test	rsi, rsi
  3a42e4:      	je	0x3a4330 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1940>
  3a42e6:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a42eb:      	mov	r11, qword ptr [rbp - 0x38]
  3a42ef:      	mov	r8, qword ptr [rbp - 0x30]
  3a42f3:      	je	0x3a3e5b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x146b>
  3a42f9:      	jmp	0x3a4513 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b23>
  3a42fe:      	jmp	0x3a3e53 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1463>
  3a4303:      	xor	edx, edx
  3a4305:      	div	ecx
  3a4307:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a430c:      	mov	rax, qword ptr [rdi + 0x8]
  3a4310:      	test	rdi, rdi
  3a4313:      	jne	0x3a41dd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17ed>
  3a4319:      	mov	edi, 0x1
  3a431e:      	call	0x5e6250 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a4323:      	mov	r8, qword ptr [rbp - 0x30]
  3a4327:      	mov	r11, qword ptr [rbp - 0x38]
  3a432b:      	jmp	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a4330:      	jmp	0x3a3e53 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1463>
  3a4335:      	mov	rax, qword ptr [rbp - 0x60]
  3a4339:      	mov	dword ptr [rax - 0x38], r14d
  3a433d:      	mov	rdi, r11
  3a4340:      	mov	r14, r11
  3a4343:      	call	0x3a6a50 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a4348:      	test	al, al
  3a434a:      	jne	0x3a44f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b05>
  3a4350:      	mov	rcx, qword ptr [r14 + 0x48]
  3a4354:      	test	rcx, rcx
  3a4357:      	mov	r11, r14
  3a435a:      	jg	0x3a2a2d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3d>
  3a4360:      	jmp	0x3a465b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6b>
  3a4365:      	mov	rax, qword ptr [rbp - 0x60]
  3a4369:      	mov	dword ptr [rax - 0x38], r14d
  3a436d:      	mov	rdi, r11
  3a4370:      	mov	r14, r11
  3a4373:      	call	0x3a57f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a4378:      	jmp	0x3a4348 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1958>
  3a437a:      	mov	rax, qword ptr [rbp - 0x60]
  3a437e:      	mov	dword ptr [rax - 0x38], r14d
  3a4382:      	mov	rdi, r11
  3a4385:      	mov	r14, r11
  3a4388:      	call	0x3a6550 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a438d:      	jmp	0x3a4348 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1958>
  3a438f:      	mov	rax, qword ptr [rbp - 0x60]
  3a4393:      	mov	dword ptr [rax - 0x38], r14d
  3a4397:      	mov	rdi, r11
  3a439a:      	mov	r14, r11
  3a439d:      	call	0x3a5360 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a43a2:      	test	al, al
  3a43a4:      	je	0x3a4350 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1960>
  3a43a6:      	mov	al, 0x1
  3a43a8:      	jmp	0x3a44f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b05>
  3a43ad:      	mov	edi, 0x1
  3a43b2:      	call	0x5e6250 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a43b7:      	jmp	0x3a44f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b05>
  3a43bc:      	mov	edi, 0x28
  3a43c1:      	call	0x5fc860 <GC_malloc@plt>
  3a43c6:      	test	rax, rax
  3a43c9:      	je	0x3a4679 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c89>
  3a43cf:      	mov	r14, rax
  3a43d2:      	add	r15, r13
  3a43d5:      	mov	edi, 0x28
  3a43da:      	call	0x5e1ee0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a43df:      	mov	rax, qword ptr [r15 + 0x8]
  3a43e3:      	movaps	xmm0, xmmword ptr [rip + 0x315466] # 0x6b9850 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a43ea:      	movups	xmmword ptr [r14], xmm0
  3a43ee:      	mov	qword ptr [r14 + 0x10], rax
  3a43f2:      	mov	qword ptr [r14 + 0x18], 0x29a060
  3a43fa:      	mov	qword ptr [r14 + 0x20], 0xe
  3a4402:      	mov	esi, 0x2baa60
  3a4407:      	jmp	0x3a44e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af7>
  3a440c:      	mov	rdi, r11
  3a440f:      	mov	rsi, rax
  3a4412:      	jmp	0x3a44ee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1afe>
  3a4417:      	mov	r15d, edx
  3a441a:      	jmp	0x3a4481 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a441c:      	mov	r14, qword ptr [rbp - 0x148]
  3a4423:      	mov	edi, 0x8
  3a4428:      	call	0x5fc860 <GC_malloc@plt>
  3a442d:      	test	rax, rax
  3a4430:      	je	0x3a4679 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c89>
  3a4436:      	mov	r12, rax
  3a4439:      	mov	edi, 0x8
  3a443e:      	call	0x5e1ee0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a4443:      	mov	edx, 0x299fd0
  3a4448:      	mov	rdi, r15
  3a444b:      	mov	rsi, r14
  3a444e:      	mov	rcx, r12
  3a4451:      	call	0x5b8b10 <errors.As>
  3a4456:      	test	al, 0x1
  3a4458:      	je	0x3a446b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7b>
  3a445a:      	mov	rax, qword ptr [r12]
  3a445e:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a4465:      	mov	ecx, dword ptr [rcx + 0x50]
  3a4468:      	mov	dword ptr [rax + 0x4], ecx
  3a446b:      	mov	rdi, qword ptr [rbp - 0x38]
  3a446f:      	mov	rsi, r15
  3a4472:      	jmp	0x3a44eb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1afb>
  3a4474:      	mov	r15b, 0x3
  3a4477:      	jmp	0x3a4481 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a4479:      	mov	r15b, 0x1
  3a447c:      	jmp	0x3a4481 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a447e:      	mov	r15b, 0x9
  3a4481:      	mov	edi, 0x10
  3a4486:      	call	0x5fc7f0 <GC_malloc_atomic@plt>
  3a448b:      	test	rax, rax
  3a448e:      	je	0x3a4679 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c89>
  3a4494:      	mov	r14, rax
  3a4497:      	mov	edi, 0x10
  3a449c:      	call	0x5e1ee0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a44a1:      	xorps	xmm0, xmm0
  3a44a4:      	movups	xmmword ptr [r14], xmm0
  3a44a8:      	mov	rax, qword ptr [rbp - 0x30]
  3a44ac:      	mov	rsi, qword ptr [rax + 0x80]
  3a44b3:      	mov	edi, dword ptr [r12 + 0x4]
  3a44b8:      	cmp	rsi, rdi
  3a44bb:      	jbe	0x3a4532 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a44bd:      	mov	rax, qword ptr [rax + 0x78]
  3a44c1:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a44c8:      	mov	ecx, dword ptr [rcx + 0x50]
  3a44cb:      	imul	rdx, rdi, 0xe0
  3a44d2:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a44d7:      	mov	byte ptr [r14], r15b
  3a44da:      	mov	dword ptr [r14 + 0x4], ecx
  3a44de:      	mov	qword ptr [r14 + 0x8], rax
  3a44e2:      	mov	esi, 0x29a3e8
  3a44e7:      	mov	rdi, qword ptr [rbp - 0x38]
  3a44eb:      	mov	rdx, r14
  3a44ee:      	call	0x3a23c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a44f3:      	mov	al, 0x3
  3a44f5:      	mov	rdx, rbx
  3a44f8:      	add	rsp, 0x2a8
  3a44ff:      	pop	rbx
  3a4500:      	pop	r12
  3a4502:      	pop	r13
  3a4504:      	pop	r14
  3a4506:      	pop	r15
  3a4508:      	pop	rbp
  3a4509:      	ret
  3a450a:      	mov	rdx, qword ptr [rbp - 0xf8]
  3a4511:      	jmp	0x3a451a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2a>
  3a4513:      	mov	rdx, qword ptr [rbp - 0x120]
  3a451a:      	mov	rdi, r11
  3a451d:      	jmp	0x3a44ee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1afe>
  3a451f:      	mov	rdi, r14
  3a4522:      	mov	rsi, r15
  3a4525:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a452a:      	mov	rsi, r13
  3a452d:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4532:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4537:      	mov	rdi, r12
  3a453a:      	mov	rsi, r15
  3a453d:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4542:      	mov	rdi, rax
  3a4545:      	mov	rsi, r13
  3a4548:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a454d:      	mov	rdi, rcx
  3a4550:      	mov	rsi, r13
  3a4553:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4558:      	mov	rsi, r13
  3a455b:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4560:      	mov	rdi, r9
  3a4563:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4568:      	mov	rdi, rax
  3a456b:      	mov	rsi, r13
  3a456e:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4573:      	mov	rdi, rax
  3a4576:      	mov	esi, 0x1
  3a457b:      	mov	ecx, 0x2
  3a4580:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4585:      	mov	rdi, r9
  3a4588:      	mov	esi, 0x1
  3a458d:      	mov	rdx, rax
  3a4590:      	mov	ecx, 0x3
  3a4595:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a459a:      	mov	rsi, rcx
  3a459d:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45a2:      	mov	rdi, rsi
  3a45a5:      	mov	rsi, rcx
  3a45a8:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45ad:      	mov	rdi, rcx
  3a45b0:      	mov	rsi, r13
  3a45b3:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a45b8:      	mov	edi, 0x1
  3a45bd:      	mov	esi, 0x1
  3a45c2:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a45c7:      	xor	edi, edi
  3a45c9:      	xor	esi, esi
  3a45cb:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a45d0:      	xor	esi, esi
  3a45d2:      	mov	rdx, rcx
  3a45d5:      	mov	ecx, 0x3
  3a45da:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a45df:      	mov	rdi, rcx
  3a45e2:      	xor	esi, esi
  3a45e4:      	mov	ecx, 0x2
  3a45e9:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a45ee:      	mov	edi, 0x3
  3a45f3:      	mov	esi, 0x3
  3a45f8:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a45fd:      	mov	edi, 0x2
  3a4602:      	mov	esi, 0x2
  3a4607:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a460c:      	mov	rdi, r12
  3a460f:      	xor	esi, esi
  3a4611:      	mov	ecx, 0x3
  3a4616:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a461b:      	mov	rdi, rdx
  3a461e:      	xor	esi, esi
  3a4620:      	mov	rdx, rax
  3a4623:      	mov	ecx, 0x2
  3a4628:      	call	0x5e2d00 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a462d:      	mov	rdi, rdx
  3a4630:      	mov	rsi, r13
  3a4633:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4638:      	mov	rdi, rcx
  3a463b:      	mov	rsi, qword ptr [rbp - 0x78]
  3a463f:      	call	0x5e1fc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4644:      	mov	rdi, rax
  3a4647:      	call	0x5e34f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a464c:      	mov	edi, 0x329598
  3a4651:      	mov	esi, 0x3298c8
  3a4656:      	call	0x5e1c00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a465b:      	mov	edi, 0x333a48
  3a4660:      	mov	esi, 0x29a050
  3a4665:      	call	0x5e1c00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a466a:      	mov	edi, 0x333a48
  3a466f:      	mov	esi, 0x29a070
  3a4674:      	call	0x5e1c00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a4679:      	mov	edi, 0x333a48
  3a467e:      	mov	esi, 0x329570
  3a4683:      	call	0x5e1c00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a4688:      	int3
  3a4689:      	int3
  3a468a:      	int3
  3a468b:      	int3
  3a468c:      	int3
  3a468d:      	int3
  3a468e:      	int3
  3a468f:      	int3
