
/home/runner/work/wasmvm/wasmvm/main-regression/scalar-fix:	file format elf64-x86-64

Disassembly of section .text:

00000000003a2800 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a2800:      	push	rbp
  3a2801:      	mov	rbp, rsp
  3a2804:      	push	r15
  3a2806:      	push	r14
  3a2808:      	push	r13
  3a280a:      	push	r12
  3a280c:      	push	rbx
  3a280d:      	sub	rsp, 0x2a8
  3a2814:      	mov	rcx, qword ptr [rdi + 0x48]
  3a2818:      	test	rcx, rcx
  3a281b:      	jle	0x3a446b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6b>
  3a2821:      	mov	rbx, rsi
  3a2824:      	mov	r11, rdi
  3a2827:      	movzx	eax, byte ptr [rdi + 0x20]
  3a282b:      	mov	byte ptr [rbp - 0x39], al
  3a282e:      	lea	rax, [rbp - 0x90]
  3a2835:      	mov	qword ptr [rbp - 0x68], rax
  3a2839:      	mov	qword ptr [rbp - 0x38], rdi
  3a283d:      	mov	rdi, qword ptr [r11 + 0x40]
  3a2841:      	shl	rcx, 0x6
  3a2845:      	mov	byte ptr [rdi + rcx - 0x30], 0x0
  3a284a:      	mov	rsi, qword ptr [rdi + rcx - 0x40]
  3a284f:      	mov	r10, qword ptr [rsi + 0x48]
  3a2853:      	mov	r13, qword ptr [r10 + 0x98]
  3a285a:      	mov	r9, qword ptr [rdi + rcx - 0x28]
  3a285f:      	mov	rdx, qword ptr [r11 + 0x88]
  3a2866:      	lea	rax, [r9 + r13]
  3a286a:      	cmp	rax, rdx
  3a286d:      	ja	0x3a4383 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b83>
  3a2873:      	cmp	r9, rax
  3a2876:      	ja	0x3a4395 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b95>
  3a287c:      	mov	r15, qword ptr [r10 + 0xb8]
  3a2883:      	mov	r12d, dword ptr [rdi + rcx - 0x38]
  3a2888:      	sub	rdx, r9
  3a288b:      	shl	r9, 0x4
  3a288f:      	test	rdx, rdx
  3a2892:      	mov	eax, 0x0
  3a2897:      	cmovle	r9, rax
  3a289b:      	cmp	r15, r12
  3a289e:      	jbe	0x3a4347 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a28a4:      	add	rdi, rcx
  3a28a7:      	mov	qword ptr [rbp - 0x60], rdi
  3a28ab:      	mov	qword ptr [rbp - 0xa8], rsi
  3a28b2:      	mov	rax, qword ptr [rsi + 0x40]
  3a28b6:      	mov	qword ptr [rbp - 0x50], rax
  3a28ba:      	mov	rcx, qword ptr [r10 + 0xb0]
  3a28c1:      	add	r9, qword ptr [r11 + 0x78]
  3a28c5:      	mov	r14d, r12d
  3a28c8:      	mov	qword ptr [rbp - 0x78], r13
  3a28cc:      	mov	qword ptr [rbp - 0x30], r10
  3a28d0:      	mov	qword ptr [rbp - 0x48], r9
  3a28d4:      	mov	qword ptr [rbp - 0xb0], r15
  3a28db:      	mov	qword ptr [rbp - 0x58], rcx
  3a28df:      	mov	esi, 0xffffffff
  3a28e4:      	jmp	0x3a2913 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x113>
  3a28e6:      	nop	word ptr cs:[rax + rax]
  3a28f0:      	inc	r14d
  3a28f3:      	mov	r15d, r14d
  3a28f6:      	mov	rcx, qword ptr [rbp - 0x58]
  3a28fa:      	mov	r10, r8
  3a28fd:      	mov	r12d, r15d
  3a2900:      	mov	r14d, r15d
  3a2903:      	mov	r15, qword ptr [rbp - 0xb0]
  3a290a:      	cmp	r15, r12
  3a290d:      	jbe	0x3a4347 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b47>
  3a2913:      	shl	r12, 0x5
  3a2917:      	add	r12, rcx
  3a291a:      	test	byte ptr [rbp - 0x39], 0x1
  3a291e:      	je	0x3a2980 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x180>
  3a2920:      	mov	rsi, qword ptr [r10 + 0x80]
  3a2927:      	mov	edi, dword ptr [r12 + 0x4]
  3a292c:      	cmp	rsi, rdi
  3a292f:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a2935:      	mov	r15, qword ptr [r10 + 0x78]
  3a2939:      	imul	r13, rdi, 0xe0
  3a2940:      	mov	eax, dword ptr [r15 + r13]
  3a2944:      	cmp	eax, 0x5
  3a2947:      	setne	cl
  3a294a:      	cmp	eax, 0xb
  3a294d:      	setne	al
  3a2950:      	and	al, cl
  3a2952:      	movzx	ecx, al
  3a2955:      	mov	rax, qword ptr [r11 + 0x150]
  3a295c:      	sub	rax, rcx
  3a295f:      	jb	0x3a41cc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19cc>
  3a2965:      	mov	qword ptr [r11 + 0x150], rax
  3a296c:      	mov	r13, qword ptr [rbp - 0x78]
  3a2970:      	mov	r15, qword ptr [rbp - 0xb0]
  3a2977:      	mov	rcx, qword ptr [rbp - 0x58]
  3a297b:      	mov	esi, 0xffffffff
  3a2980:      	inc	rbx
  3a2983:      	test	bl, 0x3f
  3a2986:      	jne	0x3a2a05 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x205>
  3a2988:      	mov	rax, qword ptr [r11 + 0x160]
  3a298f:      	test	rax, rax
  3a2992:      	je	0x3a29a7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7>
  3a2994:      	test	r11, r11
  3a2997:      	je	0x3a2aaa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2aa>
  3a299d:      	mov	eax, dword ptr [rax]
  3a299f:      	test	eax, eax
  3a29a1:      	je	0x3a2f13 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x713>
  3a29a7:      	mov	rax, qword ptr [r11 + 0x8]
  3a29ab:      	test	rax, rax
  3a29ae:      	je	0x3a445c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a29b4:      	mov	rdi, qword ptr [r11 + 0x10]
  3a29b8:      	mov	rcx, qword ptr [rax + 0x8]
  3a29bc:      	test	byte ptr [rcx + 0x17], 0x20
  3a29c0:      	je	0x3a29d9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d9>
  3a29c2:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a29c6:      	and	cl, 0x1f
  3a29c9:      	cmp	cl, 0x16
  3a29cc:      	je	0x3a29d9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d9>
  3a29ce:      	mov	qword ptr [rbp - 0x90], rdi
  3a29d5:      	mov	rdi, qword ptr [rbp - 0x68]
  3a29d9:      	call	qword ptr [rax + 0x28]
  3a29dc:      	mov	esi, 0xffffffff
  3a29e1:      	mov	r10, qword ptr [rbp - 0x30]
  3a29e5:      	mov	r9, qword ptr [rbp - 0x48]
  3a29e9:      	mov	rcx, qword ptr [rbp - 0x58]
  3a29ed:      	test	rax, rax
  3a29f0:      	je	0x3a2ace <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2ce>
  3a29f6:      	cmp	qword ptr [rax + 0x8], 0x0
  3a29fb:      	mov	r11, qword ptr [rbp - 0x38]
  3a29ff:      	jne	0x3a421c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a1c>
  3a2a05:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a2a0b:      	je	0x3a2ae0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2a11:      	test	byte ptr [rbp - 0x39], 0x1
  3a2a15:      	je	0x3a2a40 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2a17:      	mov	edx, dword ptr [r12]
  3a2a1b:      	cmp	edx, 0x20
  3a2a1e:      	je	0x3a2ae0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2a24:      	cmp	edx, 0x41
  3a2a27:      	je	0x3a2ae0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2a2d:      	jmp	0x3a2ae4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e4>
  3a2a32:      	nop	word ptr cs:[rax + rax]
  3a2a40:      	mov	edx, dword ptr [r12]
  3a2a44:      	cmp	edx, 0x20
  3a2a47:      	je	0x3a2a52 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x252>
  3a2a49:      	cmp	edx, 0x41
  3a2a4c:      	jne	0x3a2ae4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e4>
  3a2a52:      	mov	eax, ebx
  3a2a54:      	not	eax
  3a2a56:      	test	al, 0x3f
  3a2a58:      	je	0x3a2ae0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2a5e:      	cmp	edx, 0x41
  3a2a61:      	jne	0x3a2a87 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x287>
  3a2a63:      	mov	edi, dword ptr [r12 + 0x8]
  3a2a68:      	cmp	r13, rdi
  3a2a6b:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2a71:      	mov	eax, dword ptr [r12 + 0x18]
  3a2a76:      	shl	rdi, 0x4
  3a2a7a:      	mov	qword ptr [r9 + rdi], rax
  3a2a7e:      	mov	qword ptr [r9 + rdi + 0x8], 0x0
  3a2a87:      	inc	r14d
  3a2a8a:      	cmp	r15, r14
  3a2a8d:      	jbe	0x3a432f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2f>
  3a2a93:      	mov	rax, r14
  3a2a96:      	shl	rax, 0x5
  3a2a9a:      	lea	r12, [rcx + rax]
  3a2a9e:      	inc	rbx
  3a2aa1:      	cmp	dword ptr [rcx + rax + 0x1c], 0x0
  3a2aa6:      	jne	0x3a2a40 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a2aa8:      	jmp	0x3a2ae0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2e0>
  3a2aaa:      	mov	edi, 0x1
  3a2aaf:      	call	0x5e7260 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a2ab4:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2ab8:      	mov	r9, qword ptr [rbp - 0x48]
  3a2abc:      	mov	r10, qword ptr [rbp - 0x30]
  3a2ac0:      	mov	esi, 0xffffffff
  3a2ac5:      	test	rax, rax
  3a2ac8:      	jne	0x3a29f6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f6>
  3a2ace:      	mov	r11, qword ptr [rbp - 0x38]
  3a2ad2:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a2ad8:      	jne	0x3a2a11 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x211>
  3a2ade:      	nop
  3a2ae0:      	mov	edx, dword ptr [r12]
  3a2ae4:      	lea	eax, [rdx - 0xe0]
  3a2aea:      	cmp	eax, 0xffff20
  3a2aef:      	jae	0x3a2d50 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x550>
  3a2af5:      	mov	eax, edx
  3a2af7:      	shr	eax, 0x10
  3a2afa:      	cmp	eax, 0xfb
  3a2aff:      	je	0x3a2cb2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4b2>
  3a2b05:      	cmp	eax, 0xfc
  3a2b0a:      	je	0x3a2b94 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x394>
  3a2b10:      	cmp	eax, 0xfd
  3a2b15:      	jne	0x3a2d36 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x536>
  3a2b1b:      	mov	edi, dword ptr [r12 + 0xc]
  3a2b20:      	cmp	rdi, rsi
  3a2b23:      	je	0x3a2fcc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7cc>
  3a2b29:      	cmp	r13, rdi
  3a2b2c:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2b32:      	shl	rdi, 0x4
  3a2b36:      	mov	rcx, qword ptr [r9 + rdi]
  3a2b3a:      	mov	r10, qword ptr [r9 + rdi + 0x8]
  3a2b3f:      	mov	edi, dword ptr [r12 + 0x10]
  3a2b44:      	cmp	rdi, rsi
  3a2b47:      	je	0x3a2fdf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7df>
  3a2b4d:      	mov	esi, 0xffffffff
  3a2b52:      	cmp	r13, rdi
  3a2b55:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2b5b:      	shl	rdi, 0x4
  3a2b5f:      	mov	rax, qword ptr [r9 + rdi]
  3a2b63:      	mov	r8, qword ptr [r9 + rdi + 0x8]
  3a2b68:      	mov	edi, dword ptr [r12 + 0x14]
  3a2b6d:      	cmp	rdi, rsi
  3a2b70:      	je	0x3a2ff7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7f7>
  3a2b76:      	cmp	r13, rdi
  3a2b79:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2b7f:      	shl	rdi, 0x4
  3a2b83:      	mov	rsi, r9
  3a2b86:      	mov	r9, qword ptr [r9 + rdi]
  3a2b8a:      	mov	r11, qword ptr [rsi + rdi + 0x8]
  3a2b8f:      	jmp	0x3a2ffd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7fd>
  3a2b94:      	add	edx, 0xff03ffed
  3a2b9a:      	cmp	edx, 0x4
  3a2b9d:      	jae	0x3a2d50 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x550>
  3a2ba3:      	mov	rdx, qword ptr [r10 + 0xd8]
  3a2baa:      	mov	edi, dword ptr [r12 + 0x18]
  3a2baf:      	mov	eax, dword ptr [r12 + 0x1c]
  3a2bb4:      	lea	ecx, [rdi + rax]
  3a2bb7:      	add	ecx, 0x2
  3a2bba:      	cmp	rdx, rcx
  3a2bbd:      	jb	0x3a43ef <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bef>
  3a2bc3:      	cmp	edi, ecx
  3a2bc5:      	ja	0x3a43e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1be0>
  3a2bcb:      	sub	rdx, rdi
  3a2bce:      	lea	r8, [4*rdi]
  3a2bd6:      	test	rdx, rdx
  3a2bd9:      	mov	eax, 0x0
  3a2bde:      	cmovle	r8, rax
  3a2be2:      	cmp	ecx, edi
  3a2be4:      	je	0x3a43d7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd7>
  3a2bea:      	add	r8, qword ptr [r10 + 0xc8]
  3a2bf1:      	mov	eax, dword ptr [r8]
  3a2bf4:      	cmp	r13, rax
  3a2bf7:      	jbe	0x3a4378 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b78>
  3a2bfd:      	sub	rcx, rdi
  3a2c00:      	cmp	ecx, 0x1
  3a2c03:      	je	0x3a43c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bc8>
  3a2c09:      	mov	edi, dword ptr [r8 + 0x4]
  3a2c0d:      	cmp	r13, rdi
  3a2c10:      	jbe	0x3a4368 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b68>
  3a2c16:      	shl	rax, 0x4
  3a2c1a:      	mov	rdx, r9
  3a2c1d:      	mov	r9, qword ptr [r9 + rax]
  3a2c21:      	shl	rdi, 0x4
  3a2c25:      	mov	rsi, qword ptr [rdx + rdi]
  3a2c29:      	mov	r10d, dword ptr [r12]
  3a2c2d:      	lea	eax, [r10 - 0xfc0013]
  3a2c34:      	cmp	eax, 0x2
  3a2c37:      	jae	0x3a313e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x93e>
  3a2c3d:      	cmp	ecx, 0x3
  3a2c40:      	jb	0x3a440d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c0d>
  3a2c46:      	mov	edi, dword ptr [r8 + 0x8]
  3a2c4a:      	cmp	r13, rdi
  3a2c4d:      	jbe	0x3a4368 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b68>
  3a2c53:      	cmp	ecx, 0x3
  3a2c56:      	je	0x3a43fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bfe>
  3a2c5c:      	mov	eax, dword ptr [r8 + 0xc]
  3a2c60:      	cmp	r13, rax
  3a2c63:      	jbe	0x3a4378 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b78>
  3a2c69:      	shl	rdi, 0x4
  3a2c6d:      	mov	rdx, qword ptr [rbp - 0x48]
  3a2c71:      	mov	rdi, qword ptr [rdx + rdi]
  3a2c75:      	shl	rax, 0x4
  3a2c79:      	mov	rdx, qword ptr [rdx + rax]
  3a2c7d:      	cmp	r10d, 0xfc0013
  3a2c84:      	jne	0x3a327b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa7b>
  3a2c8a:      	lea	rax, [rdi + r9]
  3a2c8e:      	mov	r10, rdi
  3a2c91:      	and	r10, r9
  3a2c94:      	or	rdi, r9
  3a2c97:      	mov	r9, rax
  3a2c9a:      	not	r9
  3a2c9d:      	and	r9, rdi
  3a2ca0:      	or	r9, r10
  3a2ca3:      	shr	r9, 0x3f
  3a2ca7:      	add	rdx, rsi
  3a2caa:      	add	rdx, r9
  3a2cad:      	jmp	0x3a32a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa0>
  3a2cb2:      	add	edx, -0x2
  3a2cb5:      	cmp	dx, 0x21
  3a2cb9:      	jae	0x3a2eaa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6aa>
  3a2cbf:      	lea	rax, [rsi + 0x1c003e10]
  3a2cc6:      	movzx	ecx, dx
  3a2cc9:      	bt	rax, rcx
  3a2ccd:      	jae	0x3a2eaa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6aa>
  3a2cd3:      	mov	edi, dword ptr [r12 + 0xc]
  3a2cd8:      	mov	eax, 0xffffffff
  3a2cdd:      	cmp	rdi, rsi
  3a2ce0:      	je	0x3a331b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb1b>
  3a2ce6:      	cmp	r13, rdi
  3a2ce9:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2cef:      	shl	rdi, 0x4
  3a2cf3:      	mov	r8, qword ptr [r9 + rdi]
  3a2cf7:      	mov	edi, dword ptr [r12 + 0x10]
  3a2cfc:      	cmp	rdi, rax
  3a2cff:      	je	0x3a332c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb2c>
  3a2d05:      	cmp	r13, rdi
  3a2d08:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2d0e:      	shl	rdi, 0x4
  3a2d12:      	mov	rax, qword ptr [r9 + rdi]
  3a2d16:      	mov	r9, qword ptr [r9 + rdi + 0x8]
  3a2d1b:      	mov	r10d, 0x0
  3a2d21:      	mov	r11d, 0x0
  3a2d27:      	cmp	dx, 0xc
  3a2d2b:      	je	0x3a3343 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb43>
  3a2d31:      	jmp	0x3a3383 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb83>
  3a2d36:      	cmp	edx, 0xe7
  3a2d3c:      	jb	0x3a4145 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1945>
  3a2d42:      	nop	word ptr cs:[rax + rax]
  3a2d50:      	mov	edi, dword ptr [r12 + 0xc]
  3a2d55:      	cmp	rdi, rsi
  3a2d58:      	je	0x3a2dd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d0>
  3a2d5a:      	cmp	r13, rdi
  3a2d5d:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a2d63:      	mov	r8, r10
  3a2d66:      	mov	rax, rdi
  3a2d69:      	shl	rax, 0x4
  3a2d6d:      	mov	r10, qword ptr [r9 + rax]
  3a2d71:      	mov	ecx, dword ptr [r12 + 0x10]
  3a2d76:      	cmp	rcx, rsi
  3a2d79:      	je	0x3a2de0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5e0>
  3a2d7b:      	cmp	r13, rcx
  3a2d7e:      	jbe	0x3a435d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b5d>
  3a2d84:      	mov	rax, rcx
  3a2d87:      	shl	rax, 0x4
  3a2d8b:      	mov	rdx, qword ptr [r9 + rax]
  3a2d8f:      	mov	r9d, dword ptr [r12]
  3a2d93:      	lea	eax, [r9 - 0x45]
  3a2d97:      	cmp	eax, 0x7f
  3a2d9a:      	ja	0x3a2def <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5ef>
  3a2d9c:      	ja	0x3a3c14 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1414>
  3a2da2:      	jmp	qword ptr [8*rax + 0x203ab8]
  3a2da9:      	mov	edi, r9d
  3a2dac:      	mov	rsi, r10
  3a2daf:      	call	0x3d0d30 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a2db4:      	test	dl, dl
  3a2db6:      	jne	0x3a4227 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a27>
  3a2dbc:      	xor	ecx, ecx
  3a2dbe:      	mov	r11, qword ptr [rbp - 0x38]
  3a2dc2:      	mov	r8, qword ptr [rbp - 0x30]
  3a2dc6:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a2dcb:      	nop	dword ptr [rax + rax]
  3a2dd0:      	mov	r8, r10
  3a2dd3:      	xor	r10d, r10d
  3a2dd6:      	mov	ecx, dword ptr [r12 + 0x10]
  3a2ddb:      	cmp	rcx, rsi
  3a2dde:      	jne	0x3a2d7b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x57b>
  3a2de0:      	xor	edx, edx
  3a2de2:      	mov	r9d, dword ptr [r12]
  3a2de6:      	lea	eax, [r9 - 0x45]
  3a2dea:      	cmp	eax, 0x7f
  3a2ded:      	jbe	0x3a2d9c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x59c>
  3a2def:      	lea	eax, [r9 - 0x1]
  3a2df3:      	cmp	eax, 0xd3
  3a2df8:      	mov	esi, 0xffffffff
  3a2dfd:      	ja	0x3a2ee6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6e6>
  3a2e03:      	jmp	qword ptr [8*rax + 0x2033d8]
  3a2e0a:      	mov	rsi, qword ptr [r8 + 0x80]
  3a2e11:      	mov	edi, dword ptr [r12 + 0x4]
  3a2e16:      	cmp	rsi, rdi
  3a2e19:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a2e1f:      	imul	rax, rdi, 0xe0
  3a2e26:      	add	rax, qword ptr [r8 + 0x78]
  3a2e2a:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a2e30:      	mov	rcx, qword ptr [rbp - 0x50]
  3a2e34:      	mov	rsi, qword ptr [rcx + 0x78]
  3a2e38:      	mov	edi, dword ptr [rax + 0x18]
  3a2e3b:      	cmp	rsi, rdi
  3a2e3e:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a2e44:      	shl	rdi, 0x3
  3a2e48:      	add	rdi, qword ptr [rcx + 0x70]
  3a2e4c:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a2e52:      	mov	rcx, qword ptr [rdi]
  3a2e55:      	test	rcx, rcx
  3a2e58:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a2e5e:      	mov	rsi, qword ptr [rcx + 0x10]
  3a2e62:      	mov	ecx, dword ptr [rax + 0x14]
  3a2e65:      	mov	r8d, dword ptr [rax + 0x1c]
  3a2e69:      	shl	r8, 0x20
  3a2e6d:      	or	r8, rcx
  3a2e70:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a2e75:      	je	0x3a2e82 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x682>
  3a2e77:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a2e7c:      	je	0x3a3256 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa56>
  3a2e82:      	xorps	xmm0, xmm0
  3a2e85:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a2e8c:      	movups	xmmword ptr [rsp], xmm0
  3a2e90:      	mov	r15, qword ptr [rbp - 0x68]
  3a2e94:      	mov	rdi, r15
  3a2e97:      	mov	edx, r9d
  3a2e9a:      	mov	rcx, r10
  3a2e9d:      	xor	r9d, r9d
  3a2ea0:      	call	0x3c6e90 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a2ea5:      	jmp	0x3a3237 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa37>
  3a2eaa:      	mov	r15, qword ptr [rbp - 0x60]
  3a2eae:      	mov	dword ptr [r15 - 0x38], r14d
  3a2eb2:      	mov	rdi, r11
  3a2eb5:      	call	0x3a46e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a2eba:      	mov	r9, qword ptr [rbp - 0x48]
  3a2ebe:      	mov	r10, qword ptr [rbp - 0x30]
  3a2ec2:      	mov	esi, 0xffffffff
  3a2ec7:      	mov	r11, qword ptr [rbp - 0x38]
  3a2ecb:      	test	al, al
  3a2ecd:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2ed1:      	jne	0x3a4305 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b05>
  3a2ed7:      	inc	r14d
  3a2eda:      	mov	dword ptr [r15 - 0x38], r14d
  3a2ede:      	mov	r15d, r14d
  3a2ee1:      	jmp	0x3a28fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a2ee6:      	lea	eax, [r9 - 0xfc0000]
  3a2eed:      	cmp	eax, 0x8
  3a2ef0:      	jae	0x3a30e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8e0>
  3a2ef6:      	mov	edi, r9d
  3a2ef9:      	mov	rsi, r10
  3a2efc:      	call	0x3d0d30 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a2f01:      	mov	r8, qword ptr [rbp - 0x30]
  3a2f05:      	mov	r11, qword ptr [rbp - 0x38]
  3a2f09:      	mov	r15d, edx
  3a2f0c:      	xor	ecx, ecx
  3a2f0e:      	jmp	0x3a3c6b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x146b>
  3a2f13:      	xor	eax, eax
  3a2f15:      	xor	edx, edx
  3a2f17:      	test	rax, rax
  3a2f1a:      	jne	0x3a29f6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f6>
  3a2f20:      	jmp	0x3a2ace <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2ce>
  3a2f25:      	mov	rsi, qword ptr [r8 + 0x80]
  3a2f2c:      	mov	edi, dword ptr [r12 + 0x4]
  3a2f31:      	cmp	rsi, rdi
  3a2f34:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a2f3a:      	imul	rax, rdi, 0xe0
  3a2f41:      	add	rax, qword ptr [r8 + 0x78]
  3a2f45:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a2f4b:      	mov	rcx, qword ptr [rbp - 0x50]
  3a2f4f:      	mov	rsi, qword ptr [rcx + 0x78]
  3a2f53:      	mov	edi, dword ptr [rax + 0x18]
  3a2f56:      	cmp	rsi, rdi
  3a2f59:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a2f5f:      	shl	rdi, 0x3
  3a2f63:      	add	rdi, qword ptr [rcx + 0x70]
  3a2f67:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a2f6d:      	mov	rcx, qword ptr [rdi]
  3a2f70:      	test	rcx, rcx
  3a2f73:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a2f79:      	mov	rdi, qword ptr [rcx + 0x10]
  3a2f7d:      	mov	esi, dword ptr [rax + 0x14]
  3a2f80:      	mov	ecx, dword ptr [rax + 0x1c]
  3a2f83:      	shl	rcx, 0x20
  3a2f87:      	or	rcx, rsi
  3a2f8a:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a2f8f:      	je	0x3a2f9c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x79c>
  3a2f91:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a2f96:      	je	0x3a34c1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcc1>
  3a2f9c:      	mov	qword ptr [rbp - 0x90], rdx
  3a2fa3:      	mov	qword ptr [rbp - 0x88], 0x0
  3a2fae:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a2fb5:      	movups	xmmword ptr [rsp], xmm0
  3a2fb9:      	mov	esi, r9d
  3a2fbc:      	mov	rdx, r10
  3a2fbf:      	xor	r8d, r8d
  3a2fc2:      	call	0x3c8380 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a2fc7:      	jmp	0x3a3c5c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x145c>
  3a2fcc:      	xor	r10d, r10d
  3a2fcf:      	xor	ecx, ecx
  3a2fd1:      	mov	edi, dword ptr [r12 + 0x10]
  3a2fd6:      	cmp	rdi, rsi
  3a2fd9:      	jne	0x3a2b4d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x34d>
  3a2fdf:      	mov	esi, 0xffffffff
  3a2fe4:      	xor	eax, eax
  3a2fe6:      	xor	r8d, r8d
  3a2fe9:      	mov	edi, dword ptr [r12 + 0x14]
  3a2fee:      	cmp	rdi, rsi
  3a2ff1:      	jne	0x3a2b76 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x376>
  3a2ff7:      	xor	r9d, r9d
  3a2ffa:      	xor	r11d, r11d
  3a2ffd:      	mov	r15, qword ptr [rbp - 0x30]
  3a3001:      	mov	rsi, qword ptr [r15 + 0x80]
  3a3008:      	mov	edi, dword ptr [r12 + 0x4]
  3a300d:      	cmp	rsi, rdi
  3a3010:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3016:      	imul	rdi, rdi, 0xe0
  3a301d:      	add	rdi, qword ptr [r15 + 0x78]
  3a3021:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3027:      	movzx	esi, dx
  3a302a:      	cmp	esi, 0xb
  3a302d:      	setb	r15b
  3a3031:      	mov	esi, edx
  3a3033:      	and	esi, 0xfffc
  3a3039:      	cmp	esi, 0x54
  3a303c:      	sete	r13b
  3a3040:      	or	r13b, r15b
  3a3043:      	jne	0x3a3057 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x857>
  3a3045:      	movzx	r15d, dx
  3a3049:      	lea	r13d, [r15 - 0x5c]
  3a304d:      	cmp	r13d, 0x2
  3a3051:      	jae	0x3a316e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x96e>
  3a3057:      	mov	r10, qword ptr [rbp - 0x50]
  3a305b:      	mov	rsi, qword ptr [r10 + 0x78]
  3a305f:      	mov	r9d, dword ptr [rdi + 0x18]
  3a3063:      	cmp	rsi, r9
  3a3066:      	jbe	0x3a4370 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b70>
  3a306c:      	shl	r9, 0x3
  3a3070:      	add	r9, qword ptr [r10 + 0x70]
  3a3074:      	mov	r15, qword ptr [rbp - 0x68]
  3a3078:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a307e:      	mov	rsi, qword ptr [r9]
  3a3081:      	test	rsi, rsi
  3a3084:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a308a:      	mov	rsi, qword ptr [rsi + 0x10]
  3a308e:      	mov	r9d, dword ptr [rdi + 0x14]
  3a3092:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a3096:      	shl	r10, 0x20
  3a309a:      	or	r10, r9
  3a309d:      	movzx	edi, byte ptr [rdi + 0x40]
  3a30a1:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a30a6:      	je	0x3a30b3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8b3>
  3a30a8:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a30ad:      	je	0x3a320b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa0b>
  3a30b3:      	mov	qword ptr [rbp - 0xa0], rax
  3a30ba:      	mov	qword ptr [rbp - 0x98], r8
  3a30c1:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a30c8:      	movups	xmmword ptr [rsp], xmm0
  3a30cc:      	movzx	r9d, dil
  3a30d0:      	mov	rdi, r15
  3a30d3:      	mov	r8, r10
  3a30d6:      	call	0x3c6e90 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a30db:      	jmp	0x3a3233 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa33>
  3a30e0:      	lea	eax, [r9 - 0x1000001]
  3a30e7:      	cmp	eax, 0x8
  3a30ea:      	ja	0x3a3de3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15e3>
  3a30f0:      	mov	ecx, 0x117
  3a30f5:      	bt	ecx, eax
  3a30f8:      	jae	0x3a3dd1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15d1>
  3a30fe:      	movzx	eax, r9b
  3a3102:      	xor	edi, edi
  3a3104:      	add	eax, -0x2
  3a3107:      	cmp	eax, 0x7
  3a310a:      	ja	0x3a3cff <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14ff>
  3a3110:      	mov	r9, qword ptr [rbp - 0x48]
  3a3114:      	jmp	qword ptr [8*rax + 0x203a78]
  3a311b:      	xor	edi, edi
  3a311d:      	test	r10d, r10d
  3a3120:      	sete	dil
  3a3124:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a312b:      	add	edi, dword ptr [r12 + 0x18]
  3a3130:      	cmp	rsi, rdi
  3a3133:      	ja	0x3a3592 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd92>
  3a3139:      	jmp	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a313e:      	mov	rax, rsi
  3a3141:      	mul	r9
  3a3144:      	cmp	dword ptr [r12], 0xfc0015
  3a314c:      	jne	0x3a32a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa0>
  3a3152:      	mov	rdi, r9
  3a3155:      	sar	rdi, 0x3f
  3a3159:      	and	rdi, rsi
  3a315c:      	sar	rsi, 0x3f
  3a3160:      	and	rsi, r9
  3a3163:      	add	rsi, rdi
  3a3166:      	sub	rdx, rsi
  3a3169:      	jmp	0x3a32a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa0>
  3a316e:      	cmp	r15d, 0xb
  3a3172:      	je	0x3a317d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x97d>
  3a3174:      	cmp	esi, 0x58
  3a3177:      	jne	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd2a>
  3a317d:      	mov	r10, qword ptr [rbp - 0x50]
  3a3181:      	mov	rsi, qword ptr [r10 + 0x78]
  3a3185:      	mov	r9d, dword ptr [rdi + 0x18]
  3a3189:      	cmp	rsi, r9
  3a318c:      	mov	r13, qword ptr [rbp - 0x78]
  3a3190:      	jbe	0x3a4370 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b70>
  3a3196:      	shl	r9, 0x3
  3a319a:      	add	r9, qword ptr [r10 + 0x70]
  3a319e:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a31a4:      	mov	rsi, qword ptr [r9]
  3a31a7:      	test	rsi, rsi
  3a31aa:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a31b0:      	mov	rsi, qword ptr [rsi + 0x10]
  3a31b4:      	mov	r10d, dword ptr [rdi + 0x14]
  3a31b8:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a31bc:      	shl	r9, 0x20
  3a31c0:      	or	r9, r10
  3a31c3:      	movzx	edi, byte ptr [rdi + 0x40]
  3a31c7:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a31cc:      	je	0x3a31d9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9d9>
  3a31ce:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a31d3:      	je	0x3a3c2f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x142f>
  3a31d9:      	mov	qword ptr [rbp - 0x90], rax
  3a31e0:      	mov	qword ptr [rbp - 0x88], r8
  3a31e7:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a31ee:      	movups	xmmword ptr [rsp], xmm0
  3a31f2:      	movzx	r8d, dil
  3a31f6:      	mov	rdi, rsi
  3a31f9:      	mov	esi, edx
  3a31fb:      	mov	rdx, rcx
  3a31fe:      	mov	rcx, r9
  3a3201:      	call	0x3c8380 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a3206:      	jmp	0x3a3c5c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x145c>
  3a320b:      	mov	qword ptr [rbp - 0xa0], rax
  3a3212:      	mov	qword ptr [rbp - 0x98], r8
  3a3219:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a3220:      	movups	xmmword ptr [rsp], xmm0
  3a3224:      	movzx	r9d, dil
  3a3228:      	mov	rdi, r15
  3a322b:      	mov	r8, r10
  3a322e:      	call	0x3c6fd0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a3233:      	mov	r13, qword ptr [rbp - 0x78]
  3a3237:      	mov	rax, qword ptr [r15]
  3a323a:      	mov	rcx, qword ptr [r15 + 0x8]
  3a323e:      	mov	qword ptr [rbp - 0x68], r15
  3a3242:      	movzx	r15d, byte ptr [r15 + 0x10]
  3a3247:      	jmp	0x3a3c63 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1463>
  3a324c:      	mov	rax, qword ptr [r12 + 0x18]
  3a3251:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3256:      	xorps	xmm0, xmm0
  3a3259:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a3260:      	movups	xmmword ptr [rsp], xmm0
  3a3264:      	mov	r15, qword ptr [rbp - 0x68]
  3a3268:      	mov	rdi, r15
  3a326b:      	mov	edx, r9d
  3a326e:      	mov	rcx, r10
  3a3271:      	xor	r9d, r9d
  3a3274:      	call	0x3c6fd0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a3279:      	jmp	0x3a3237 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa37>
  3a327b:      	mov	rax, r9
  3a327e:      	sub	rax, rdi
  3a3281:      	not	r9
  3a3284:      	mov	r10, rdi
  3a3287:      	and	r10, r9
  3a328a:      	xor	r9, rdi
  3a328d:      	and	r9, rax
  3a3290:      	or	r9, r10
  3a3293:      	shr	r9, 0x3f
  3a3297:      	add	rdx, r9
  3a329a:      	sub	rsi, rdx
  3a329d:      	mov	rdx, rsi
  3a32a0:      	mov	r10, qword ptr [rbp - 0x30]
  3a32a4:      	mov	edi, dword ptr [r12 + 0x1c]
  3a32a9:      	cmp	ecx, edi
  3a32ab:      	jbe	0x3a43aa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1baa>
  3a32b1:      	lea	esi, [rdi + 0x1]
  3a32b4:      	cmp	ecx, esi
  3a32b6:      	mov	r9, qword ptr [rbp - 0x60]
  3a32ba:      	jbe	0x3a43b2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bb2>
  3a32c0:      	mov	ecx, dword ptr [r8 + 4*rdi]
  3a32c4:      	cmp	r13, rcx
  3a32c7:      	jbe	0x3a43bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bbd>
  3a32cd:      	mov	edi, dword ptr [r8 + 4*rsi]
  3a32d1:      	shl	rcx, 0x4
  3a32d5:      	mov	rsi, qword ptr [rbp - 0x48]
  3a32d9:      	mov	qword ptr [rsi + rcx], rax
  3a32dd:      	mov	qword ptr [rsi + rcx + 0x8], 0x0
  3a32e6:      	cmp	r13, rdi
  3a32e9:      	jbe	0x3a4368 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b68>
  3a32ef:      	shl	rdi, 0x4
  3a32f3:      	mov	qword ptr [rsi + rdi], rdx
  3a32f7:      	mov	qword ptr [rsi + rdi + 0x8], 0x0
  3a3300:      	inc	r14d
  3a3303:      	mov	dword ptr [r9 - 0x38], r14d
  3a3307:      	mov	r15d, r14d
  3a330a:      	mov	r9, rsi
  3a330d:      	mov	esi, 0xffffffff
  3a3312:      	mov	rcx, qword ptr [rbp - 0x58]
  3a3316:      	jmp	0x3a28fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a331b:      	xor	r8d, r8d
  3a331e:      	mov	edi, dword ptr [r12 + 0x10]
  3a3323:      	cmp	rdi, rax
  3a3326:      	jne	0x3a2d05 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x505>
  3a332c:      	xor	eax, eax
  3a332e:      	xor	r9d, r9d
  3a3331:      	mov	r10d, 0x0
  3a3337:      	mov	r11d, 0x0
  3a333d:      	cmp	dx, 0xc
  3a3341:      	jne	0x3a3383 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb83>
  3a3343:      	mov	rcx, qword ptr [rbp - 0x30]
  3a3347:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a334e:      	mov	edi, dword ptr [r12 + 0x18]
  3a3353:      	add	edi, 0x2
  3a3356:      	cmp	rsi, rdi
  3a3359:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a335f:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a3366:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a3369:      	cmp	r13, rdi
  3a336c:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3372:      	shl	rdi, 0x4
  3a3376:      	mov	rcx, qword ptr [rbp - 0x48]
  3a337a:      	mov	r10, qword ptr [rcx + rdi]
  3a337e:      	mov	r11, qword ptr [rcx + rdi + 0x8]
  3a3383:      	cmp	qword ptr [rbp - 0x38], 0x0
  3a3388:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a338e:      	mov	rdx, qword ptr [rbp - 0x30]
  3a3392:      	mov	rsi, qword ptr [rdx + 0x80]
  3a3399:      	mov	edi, dword ptr [r12 + 0x4]
  3a339e:      	cmp	rsi, rdi
  3a33a1:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a33a7:      	imul	rcx, rdi, 0xe0
  3a33ae:      	add	rcx, qword ptr [rdx + 0x78]
  3a33b2:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a33b8:      	mov	rdx, qword ptr [rbp - 0x38]
  3a33bc:      	mov	rsi, qword ptr [rdx]
  3a33bf:      	mov	qword ptr [rbp - 0xd0], rax
  3a33c6:      	mov	qword ptr [rbp - 0xc8], r9
  3a33cd:      	mov	qword ptr [rbp - 0xc0], r10
  3a33d4:      	mov	qword ptr [rbp - 0xb8], r11
  3a33db:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a33e2:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a33e7:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a33ee:      	movups	xmmword ptr [rsp], xmm0
  3a33f2:      	lea	rdi, [rbp - 0x160]
  3a33f9:      	mov	rdx, qword ptr [rbp - 0x50]
  3a33fd:      	call	0x3b68a0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a3402:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a3409:      	mov	r15, qword ptr [rbp - 0x150]
  3a3410:      	test	r15, r15
  3a3413:      	je	0x3a34ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcec>
  3a3419:      	cmp	qword ptr [r15 + 0x8], 0x0
  3a341e:      	mov	esi, 0xffffffff
  3a3423:      	mov	r10, qword ptr [rbp - 0x30]
  3a3427:      	mov	r9, qword ptr [rbp - 0x48]
  3a342b:      	mov	rcx, qword ptr [rbp - 0x58]
  3a342f:      	je	0x3a34fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xcfd>
  3a3435:      	jmp	0x3a422c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a2c>
  3a343a:      	cmp	r13, rdi
  3a343d:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3443:      	mov	edx, dword ptr [r12 + 0x18]
  3a3448:      	cmp	r13, rdx
  3a344b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a344f:      	jbe	0x3a443d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c3d>
  3a3455:      	shl	rdi, 0x4
  3a3459:      	mov	rax, qword ptr [rsi + rdi]
  3a345d:      	mov	rcx, qword ptr [rsi + rdi + 0x8]
  3a3462:      	shl	rdx, 0x4
  3a3466:      	mov	qword ptr [rsi + rdx], rax
  3a346a:      	mov	qword ptr [rsi + rdx + 0x8], rcx
  3a346f:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3474:      	mov	eax, dword ptr [r12 + 0x14]
  3a3479:      	cmp	rax, rsi
  3a347c:      	je	0x3a3ca6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14a6>
  3a3482:      	cmp	r13, rax
  3a3485:      	mov	rdx, qword ptr [rbp - 0x48]
  3a3489:      	jbe	0x3a4352 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b52>
  3a348f:      	cmp	r13, rcx
  3a3492:      	jbe	0x3a4448 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c48>
  3a3498:      	shl	rax, 0x4
  3a349c:      	cmp	dword ptr [rdx + rax], 0x0
  3a34a0:      	je	0x3a3cb3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14b3>
  3a34a6:      	cmp	r13, rdi
  3a34a9:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a34af:      	shl	rdi, 0x4
  3a34b3:      	mov	rax, qword ptr [rdx + rdi]
  3a34b7:      	mov	rcx, qword ptr [rdx + rdi + 0x8]
  3a34bc:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a34c1:      	mov	qword ptr [rbp - 0x90], rdx
  3a34c8:      	mov	qword ptr [rbp - 0x88], 0x0
  3a34d3:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a34da:      	movups	xmmword ptr [rsp], xmm0
  3a34de:      	mov	esi, r9d
  3a34e1:      	mov	rdx, r10
  3a34e4:      	xor	r8d, r8d
  3a34e7:      	jmp	0x3a3c57 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1457>
  3a34ec:      	mov	esi, 0xffffffff
  3a34f1:      	mov	r10, qword ptr [rbp - 0x30]
  3a34f5:      	mov	r9, qword ptr [rbp - 0x48]
  3a34f9:      	mov	rcx, qword ptr [rbp - 0x58]
  3a34fd:      	mov	edi, dword ptr [r12 + 0x8]
  3a3502:      	cmp	rdi, rsi
  3a3505:      	mov	r11, qword ptr [rbp - 0x38]
  3a3509:      	mov	r15, qword ptr [rbp - 0x60]
  3a350d:      	je	0x3a2ed7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6d7>
  3a3513:      	cmp	r13, rdi
  3a3516:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a351c:      	shl	rdi, 0x4
  3a3520:      	movups	xmmword ptr [r9 + rdi], xmm0
  3a3525:      	jmp	0x3a2ed7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6d7>
  3a352a:      	mov	qword ptr [rsp], r11
  3a352e:      	mov	rsi, rcx
  3a3531:      	mov	rdx, r10
  3a3534:      	mov	rcx, rax
  3a3537:      	call	0x3cef30 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a353c:      	mov	rcx, rdx
  3a353f:      	mov	r11, qword ptr [rbp - 0x38]
  3a3543:      	mov	r13, qword ptr [rbp - 0x78]
  3a3547:      	mov	r8, qword ptr [rbp - 0x30]
  3a354b:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3550:      	mov	edi, dword ptr [r12 + 0x1c]
  3a3555:      	dec	edi
  3a3557:      	cmp	edi, r10d
  3a355a:      	cmovae	edi, r10d
  3a355e:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a3565:      	add	edi, dword ptr [r12 + 0x18]
  3a356a:      	cmp	rsi, rdi
  3a356d:      	ja	0x3a3592 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd92>
  3a356f:      	jmp	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3574:      	xor	edi, edi
  3a3576:      	test	r10, r10
  3a3579:      	setne	dil
  3a357d:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a3584:      	add	edi, dword ptr [r12 + 0x18]
  3a3589:      	cmp	rsi, rdi
  3a358c:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3592:      	mov	r10, r8
  3a3595:      	mov	rax, qword ptr [r8 + 0xe0]
  3a359c:      	lea	rcx, [rdi + 2*rdi]
  3a35a0:      	mov	r15d, dword ptr [rax + 4*rcx]
  3a35a4:      	mov	r12d, dword ptr [rax + 4*rcx + 0x4]
  3a35a9:      	mov	esi, dword ptr [rax + 4*rcx + 0x8]
  3a35ad:      	cmp	r15d, r14d
  3a35b0:      	ja	0x3a3d3e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x153e>
  3a35b6:      	mov	rax, qword ptr [r11 + 0x160]
  3a35bd:      	test	rax, rax
  3a35c0:      	je	0x3a35d5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdd5>
  3a35c2:      	test	r11, r11
  3a35c5:      	je	0x3a3d1d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x151d>
  3a35cb:      	mov	eax, dword ptr [rax]
  3a35cd:      	test	eax, eax
  3a35cf:      	je	0x3a4017 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1817>
  3a35d5:      	mov	rax, qword ptr [r11 + 0x8]
  3a35d9:      	test	rax, rax
  3a35dc:      	je	0x3a445c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c5c>
  3a35e2:      	mov	rdi, qword ptr [r11 + 0x10]
  3a35e6:      	mov	rcx, qword ptr [rax + 0x8]
  3a35ea:      	test	byte ptr [rcx + 0x17], 0x20
  3a35ee:      	jne	0x3a3cc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14c2>
  3a35f4:      	mov	r14, rsi
  3a35f7:      	jmp	0x3a3cdc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dc>
  3a35fc:      	xor	eax, eax
  3a35fe:      	cmp	r10, rdx
  3a3601:      	sete	al
  3a3604:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3609:      	mov	rax, qword ptr [rbp - 0x50]
  3a360d:      	mov	rsi, qword ptr [rax + 0x78]
  3a3611:      	mov	edi, dword ptr [r12 + 0x18]
  3a3616:      	cmp	rsi, rdi
  3a3619:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a361f:      	shl	rdi, 0x3
  3a3623:      	add	rdi, qword ptr [rax + 0x70]
  3a3627:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a362d:      	test	r11, r11
  3a3630:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3636:      	mov	r15, qword ptr [rdi]
  3a3639:      	test	r15, r15
  3a363c:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3642:      	mov	rdi, qword ptr [r11]
  3a3645:      	mov	rax, qword ptr [r15 + 0x10]
  3a3649:      	movzx	eax, byte ptr [rax + 0x31]
  3a364d:      	mov	byte ptr [rbp - 0x69], al
  3a3650:      	mov	edx, r10d
  3a3653:      	test	al, 0x1
  3a3655:      	cmovne	rdx, r10
  3a3659:      	mov	rsi, r15
  3a365c:      	call	0x3c1000 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a3661:      	test	dl, 0x1
  3a3664:      	jne	0x3a2dbc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5bc>
  3a366a:      	mov	rax, qword ptr [r15 + 0x10]
  3a366e:      	movzx	eax, byte ptr [rax + 0x31]
  3a3672:      	mov	byte ptr [rbp - 0x6a], al
  3a3675:      	and	al, 0x1
  3a3677:      	neg	al
  3a3679:      	mov	eax, 0x0
  3a367e:      	sbb	rax, rax
  3a3681:      	mov	ecx, 0xffffffff
  3a3686:      	or	rax, rcx
  3a3689:      	jmp	0x3a2dbc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5bc>
  3a368e:      	mov	rax, qword ptr [rbp - 0x50]
  3a3692:      	mov	rsi, qword ptr [rax + 0x60]
  3a3696:      	mov	edi, dword ptr [r12 + 0x18]
  3a369b:      	cmp	rsi, rdi
  3a369e:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a36a4:      	mov	rax, qword ptr [rax + 0x58]
  3a36a8:      	lea	rcx, [rax + 8*rdi]
  3a36ac:      	test	rcx, rcx
  3a36af:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a36b5:      	mov	rax, qword ptr [rax + 8*rdi]
  3a36b9:      	test	rax, rax
  3a36bc:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a36c2:      	cmp	byte ptr [rax + 0x29], 0x0
  3a36c6:      	mov	ecx, r10d
  3a36c9:      	cmovne	rcx, r10
  3a36cd:      	cmp	rcx, qword ptr [rax + 0x40]
  3a36d1:      	jae	0x3a4284 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a84>
  3a36d7:      	mov	rax, qword ptr [rax + 0x38]
  3a36db:      	mov	rax, qword ptr [rax + 8*rcx]
  3a36df:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a36e4:      	mov	rax, qword ptr [rbp - 0x50]
  3a36e8:      	mov	rsi, qword ptr [rax + 0x48]
  3a36ec:      	mov	edi, dword ptr [r12 + 0x18]
  3a36f1:      	cmp	rsi, rdi
  3a36f4:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a36fa:      	mov	rax, qword ptr [rax + 0x40]
  3a36fe:      	mov	rax, qword ptr [rax + 8*rdi]
  3a3702:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3707:      	test	r10, r10
  3a370a:      	jne	0x3a37ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfec>
  3a3710:      	jmp	0x3a428e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a8e>
  3a3715:      	mov	rax, qword ptr [rbp - 0x50]
  3a3719:      	mov	rsi, qword ptr [rax + 0x90]
  3a3720:      	mov	eax, dword ptr [r12 + 0x18]
  3a3725:      	cmp	rsi, rax
  3a3728:      	jbe	0x3a4454 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c54>
  3a372e:      	cmp	r13, rdi
  3a3731:      	mov	rdx, qword ptr [rbp - 0x48]
  3a3735:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a373b:      	mov	rcx, qword ptr [rbp - 0x50]
  3a373f:      	mov	rcx, qword ptr [rcx + 0x88]
  3a3746:      	mov	rax, qword ptr [rcx + 8*rax]
  3a374a:      	shl	rdi, 0x4
  3a374e:      	movups	xmm0, xmmword ptr [rdx + rdi]
  3a3752:      	movups	xmmword ptr [rax + 0x20], xmm0
  3a3756:      	jmp	0x3a37dd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfdd>
  3a375b:      	mov	rax, qword ptr [rbp - 0x50]
  3a375f:      	mov	rsi, qword ptr [rax + 0x90]
  3a3766:      	mov	edi, dword ptr [r12 + 0x18]
  3a376b:      	cmp	rsi, rdi
  3a376e:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3774:      	mov	rax, qword ptr [rax + 0x88]
  3a377b:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a377f:      	mov	rax, qword ptr [rcx + 0x20]
  3a3783:      	mov	rcx, qword ptr [rcx + 0x28]
  3a3787:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a378c:      	mov	rax, qword ptr [rbp - 0x50]
  3a3790:      	mov	rsi, qword ptr [rax + 0x60]
  3a3794:      	mov	edi, dword ptr [r12 + 0x18]
  3a3799:      	cmp	rsi, rdi
  3a379c:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a37a2:      	mov	rax, qword ptr [rax + 0x58]
  3a37a6:      	lea	rcx, [rax + 8*rdi]
  3a37aa:      	test	rcx, rcx
  3a37ad:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a37b3:      	mov	rax, qword ptr [rax + 8*rdi]
  3a37b7:      	test	rax, rax
  3a37ba:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a37c0:      	cmp	byte ptr [rax + 0x29], 0x0
  3a37c4:      	mov	ecx, r10d
  3a37c7:      	cmovne	rcx, r10
  3a37cb:      	cmp	rcx, qword ptr [rax + 0x40]
  3a37cf:      	jae	0x3a4284 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a84>
  3a37d5:      	mov	rax, qword ptr [rax + 0x38]
  3a37d9:      	mov	qword ptr [rax + 8*rcx], rdx
  3a37dd:      	xor	eax, eax
  3a37df:      	nop
  3a37e0:      	xor	ecx, ecx
  3a37e2:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a37e7:      	mov	edi, dword ptr [r12 + 0x18]
  3a37ec:      	cmp	r13, rdi
  3a37ef:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a37f5:      	shl	rdi, 0x4
  3a37f9:      	mov	rcx, qword ptr [rbp - 0x48]
  3a37fd:      	mov	rax, qword ptr [rcx + rdi]
  3a3801:      	mov	rcx, qword ptr [rcx + rdi + 0x8]
  3a3806:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a380b:      	xor	eax, eax
  3a380d:      	test	r10, r10
  3a3810:      	sete	al
  3a3813:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3815:      	mov	rax, qword ptr [rbp - 0x50]
  3a3819:      	mov	rsi, qword ptr [rax + 0x78]
  3a381d:      	mov	edi, dword ptr [r12 + 0x18]
  3a3822:      	cmp	rsi, rdi
  3a3825:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a382b:      	shl	rdi, 0x3
  3a382f:      	add	rdi, qword ptr [rax + 0x70]
  3a3833:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3839:      	mov	rax, qword ptr [rdi]
  3a383c:      	test	rax, rax
  3a383f:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3845:      	mov	rdi, qword ptr [rax + 0x10]
  3a3849:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a384e:      	mov	rax, qword ptr [rdi + 0x60]
  3a3852:      	je	0x3a3fd8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17d8>
  3a3858:      	test	rax, rax
  3a385b:      	jne	0x3a3fe1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17e1>
  3a3861:      	call	0x3c4a90 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a3866:      	mov	r8, qword ptr [rbp - 0x30]
  3a386a:      	mov	r11, qword ptr [rbp - 0x38]
  3a386e:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3873:      	mov	eax, dword ptr [r12 + 0x18]
  3a3878:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a387d:      	xor	rdx, r10
  3a3880:      	jmp	0x3a3bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3885:      	mov	eax, r10d
  3a3888:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a388d:      	mov	ecx, edx
  3a388f:      	shr	r10, cl
  3a3892:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3897:      	mov	ecx, edx
  3a3899:      	shr	r10d, cl
  3a389c:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a38a1:      	or	rdx, r10
  3a38a4:      	jmp	0x3a3bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a38a9:      	mov	ecx, edx
  3a38ab:      	sar	r10d, cl
  3a38ae:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a38b3:      	xor	edx, r10d
  3a38b6:      	jmp	0x3a3bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a38bb:      	mov	rax, r10
  3a38be:      	shr	rax
  3a38c1:      	movabs	rcx, 0x5555555555555555
  3a38cb:      	and	rax, rcx
  3a38ce:      	and	r10, rcx
  3a38d1:      	add	r10, rax
  3a38d4:      	mov	rax, r10
  3a38d7:      	shr	rax, 0x2
  3a38db:      	movabs	rcx, 0x3333333333333333
  3a38e5:      	and	rax, rcx
  3a38e8:      	and	r10, rcx
  3a38eb:      	add	r10, rax
  3a38ee:      	mov	rax, r10
  3a38f1:      	shr	rax, 0x4
  3a38f5:      	add	rax, r10
  3a38f8:      	movabs	rcx, 0xf0f0f0f0f0f0f0f
  3a3902:      	and	rcx, rax
  3a3905:      	mov	rax, rcx
  3a3908:      	shr	rax, 0x8
  3a390c:      	add	rax, rcx
  3a390f:      	mov	rcx, rax
  3a3912:      	shr	rcx, 0x10
  3a3916:      	add	rcx, rax
  3a3919:      	mov	rax, rcx
  3a391c:      	shr	rax, 0x20
  3a3920:      	add	eax, ecx
  3a3922:      	and	eax, 0x7f
  3a3925:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a392a:      	xor	eax, eax
  3a392c:      	cmp	r10, rdx
  3a392f:      	setb	al
  3a3932:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3937:      	xor	eax, eax
  3a3939:      	cmp	r10, rdx
  3a393c:      	setbe	al
  3a393f:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3944:      	xor	eax, eax
  3a3946:      	cmp	r10, rdx
  3a3949:      	seta	al
  3a394c:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3951:      	imul	edx, r10d
  3a3955:      	jmp	0x3a3bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a395a:      	mov	ecx, r10d
  3a395d:      	shr	ecx, 0x10
  3a3960:      	xor	eax, eax
  3a3962:      	cmp	r10d, 0x10000
  3a3969:      	setb	al
  3a396c:      	cmovb	ecx, r10d
  3a3970:      	shl	eax, 0x4
  3a3973:      	mov	edx, ecx
  3a3975:      	shr	edx, 0x8
  3a3978:      	xor	esi, esi
  3a397a:      	cmp	ecx, 0x100
  3a3980:      	setb	sil
  3a3984:      	lea	eax, [rax + 8*rsi]
  3a3987:      	cmovb	edx, ecx
  3a398a:      	movzx	ecx, dl
  3a398d:      	movzx	ecx, byte ptr [rcx + 0x3282d1]
  3a3994:      	sub	rax, rcx
  3a3997:      	add	rax, 0x8
  3a399b:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a39a0:      	mov	ecx, edx
  3a39a2:      	rol	r10d, cl
  3a39a5:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a39aa:      	xor	eax, eax
  3a39ac:      	cmp	r10, rdx
  3a39af:      	setg	al
  3a39b2:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a39b7:      	xor	eax, eax
  3a39b9:      	cmp	r10d, edx
  3a39bc:      	setl	al
  3a39bf:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a39c4:      	xor	eax, eax
  3a39c6:      	cmp	r10, rdx
  3a39c9:      	setne	al
  3a39cc:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a39d1:      	test	r10d, r10d
  3a39d4:      	je	0x3a4029 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1829>
  3a39da:      	tzcnt	eax, r10d
  3a39df:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a39e4:      	add	edx, r10d
  3a39e7:      	jmp	0x3a3bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a39ec:      	xor	eax, eax
  3a39ee:      	cmp	r10d, edx
  3a39f1:      	setne	al
  3a39f4:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a39f9:      	xor	eax, eax
  3a39fb:      	cmp	r10, rdx
  3a39fe:      	setle	al
  3a3a01:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a06:      	mov	ecx, edx
  3a3a08:      	ror	r10d, cl
  3a3a0b:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3a10:      	xor	eax, eax
  3a3a12:      	cmp	r10d, edx
  3a3a15:      	setbe	al
  3a3a18:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a1d:      	xor	eax, eax
  3a3a1f:      	cmp	r10d, edx
  3a3a22:      	setge	al
  3a3a25:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a2a:      	xor	eax, eax
  3a3a2c:      	test	r10d, r10d
  3a3a2f:      	sete	al
  3a3a32:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a37:      	xor	eax, eax
  3a3a39:      	cmp	r10d, edx
  3a3a3c:      	setb	al
  3a3a3f:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a44:      	mov	rdx, r10
  3a3a47:      	mov	eax, edx
  3a3a49:      	shr	eax, 0x18
  3a3a4c:      	mov	ecx, edx
  3a3a4e:      	shr	ecx, 0x10
  3a3a51:      	movzx	ecx, cl
  3a3a54:      	movzx	ecx, byte ptr [rcx + 0x3283d1]
  3a3a5b:      	add	cl, byte ptr [rax + 0x3283d1]
  3a3a61:      	movzx	eax, dh
  3a3a64:      	add	cl, byte ptr [rax + 0x3283d1]
  3a3a6a:      	movzx	eax, dl
  3a3a6d:      	add	cl, byte ptr [rax + 0x3283d1]
  3a3a73:      	movzx	eax, cl
  3a3a76:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a7b:      	add	rdx, r10
  3a3a7e:      	jmp	0x3a3bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3a83:      	mov	ecx, edx
  3a3a85:      	shl	r10d, cl
  3a3a88:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3a8d:      	xor	eax, eax
  3a3a8f:      	test	r10, r10
  3a3a92:      	sete	al
  3a3a95:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3a9a:      	and	rdx, r10
  3a3a9d:      	jmp	0x3a3bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3aa2:      	sub	r10, rdx
  3a3aa5:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3aaa:      	xor	eax, eax
  3a3aac:      	cmp	r10d, edx
  3a3aaf:      	sete	al
  3a3ab2:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3ab7:      	mov	ecx, edx
  3a3ab9:      	rol	r10, cl
  3a3abc:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3ac1:      	xor	eax, eax
  3a3ac3:      	cmp	r10d, edx
  3a3ac6:      	setae	al
  3a3ac9:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3ace:      	and	r10d, edx
  3a3ad1:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3ad6:      	xor	eax, eax
  3a3ad8:      	cmp	r10d, edx
  3a3adb:      	setle	al
  3a3ade:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3ae3:      	xor	eax, eax
  3a3ae5:      	cmp	r10d, edx
  3a3ae8:      	setg	al
  3a3aeb:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3af0:      	xor	eax, eax
  3a3af2:      	mov	rcx, r10
  3a3af5:      	shr	rcx, 0x20
  3a3af9:      	cmove	rcx, r10
  3a3afd:      	setne	al
  3a3b00:      	shl	eax, 0x5
  3a3b03:      	mov	rdx, rcx
  3a3b06:      	shr	rdx, 0x10
  3a3b0a:      	lea	rsi, [rax + 0x10]
  3a3b0e:      	cmp	rcx, 0x10000
  3a3b15:      	cmovb	rdx, rcx
  3a3b19:      	cmovb	rsi, rax
  3a3b1d:      	mov	eax, edx
  3a3b1f:      	shr	eax, 0x8
  3a3b22:      	lea	rcx, [rsi + 0x8]
  3a3b26:      	cmp	rdx, 0x100
  3a3b2d:      	cmovb	rax, rdx
  3a3b31:      	cmovb	rcx, rsi
  3a3b35:      	movzx	eax, al
  3a3b38:      	movzx	edx, byte ptr [rax + 0x3282d1]
  3a3b3f:      	add	rdx, rcx
  3a3b42:      	mov	eax, 0x40
  3a3b47:      	sub	rax, rdx
  3a3b4a:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b4f:      	xor	eax, eax
  3a3b51:      	cmp	r10, rdx
  3a3b54:      	setl	al
  3a3b57:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b5c:      	mov	ecx, edx
  3a3b5e:      	ror	r10, cl
  3a3b61:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3b66:      	xor	eax, eax
  3a3b68:      	cmp	r10d, edx
  3a3b6b:      	seta	al
  3a3b6e:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b73:      	mov	eax, r10d
  3a3b76:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b7b:      	movsx	rax, r10w
  3a3b7f:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b84:      	movsxd	rax, r10d
  3a3b87:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b8c:      	xor	eax, eax
  3a3b8e:      	cmp	r10, rdx
  3a3b91:      	sete	al
  3a3b94:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3b99:      	movsxd	rax, r10d
  3a3b9c:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3ba1:      	xor	eax, eax
  3a3ba3:      	cmp	r10, rdx
  3a3ba6:      	setae	al
  3a3ba9:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3bae:      	xor	eax, eax
  3a3bb0:      	cmp	r10, rdx
  3a3bb3:      	setge	al
  3a3bb6:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3bbb:      	mov	ecx, edx
  3a3bbd:      	sar	r10, cl
  3a3bc0:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3bc2:      	or	edx, r10d
  3a3bc5:      	jmp	0x3a3bd0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d0>
  3a3bc7:      	sub	r10d, edx
  3a3bca:      	jmp	0x3a3bf9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13f9>
  3a3bcc:      	imul	rdx, r10
  3a3bd0:      	xor	ecx, ecx
  3a3bd2:      	mov	rax, rdx
  3a3bd5:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3bda:      	movsx	rax, r10w
  3a3bde:      	mov	eax, eax
  3a3be0:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3be5:      	mov	eax, 0x40
  3a3bea:      	tzcnt	rax, r10
  3a3bef:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3bf4:      	mov	ecx, edx
  3a3bf6:      	shl	r10, cl
  3a3bf9:      	xor	ecx, ecx
  3a3bfb:      	mov	rax, r10
  3a3bfe:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3c00:      	movsx	rax, r10b
  3a3c04:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3c09:      	movsx	rax, r10b
  3a3c0d:      	mov	eax, eax
  3a3c0f:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a3c14:      	movzx	esi, r9b
  3a3c18:      	add	esi, -0x61
  3a3c1b:      	cmp	esi, 0x5e
  3a3c1e:      	ja	0x3a37dd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfdd>
  3a3c24:      	xor	eax, eax
  3a3c26:      	xor	ecx, ecx
  3a3c28:      	jmp	qword ptr [8*rsi + 0x203eb8]
  3a3c2f:      	mov	qword ptr [rbp - 0x90], rax
  3a3c36:      	mov	qword ptr [rbp - 0x88], r8
  3a3c3d:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a3c44:      	movups	xmmword ptr [rsp], xmm0
  3a3c48:      	movzx	r8d, dil
  3a3c4c:      	mov	rdi, rsi
  3a3c4f:      	mov	esi, edx
  3a3c51:      	mov	rdx, rcx
  3a3c54:      	mov	rcx, r9
  3a3c57:      	call	0x3c8660 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a3c5c:      	mov	r15d, eax
  3a3c5f:      	xor	eax, eax
  3a3c61:      	xor	ecx, ecx
  3a3c63:      	mov	r11, qword ptr [rbp - 0x38]
  3a3c67:      	mov	r8, qword ptr [rbp - 0x30]
  3a3c6b:      	test	r15b, r15b
  3a3c6e:      	jne	0x3a4291 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a3c74:      	mov	edi, dword ptr [r12 + 0x8]
  3a3c79:      	mov	esi, 0xffffffff
  3a3c7e:      	cmp	rdi, rsi
  3a3c81:      	mov	r9, qword ptr [rbp - 0x48]
  3a3c85:      	je	0x3a28f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0>
  3a3c8b:      	cmp	r13, rdi
  3a3c8e:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3c94:      	shl	rdi, 0x4
  3a3c98:      	mov	qword ptr [r9 + rdi], rax
  3a3c9c:      	mov	qword ptr [r9 + rdi + 0x8], rcx
  3a3ca1:      	jmp	0x3a28f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0>
  3a3ca6:      	cmp	r13, rcx
  3a3ca9:      	mov	rdx, qword ptr [rbp - 0x48]
  3a3cad:      	jbe	0x3a4448 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c48>
  3a3cb3:      	shl	rcx, 0x4
  3a3cb7:      	mov	rax, qword ptr [rdx + rcx]
  3a3cbb:      	mov	rcx, qword ptr [rdx + rcx + 0x8]
  3a3cc0:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3cc2:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a3cc6:      	and	cl, 0x1f
  3a3cc9:      	mov	r14, rsi
  3a3ccc:      	cmp	cl, 0x16
  3a3ccf:      	je	0x3a3cdc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14dc>
  3a3cd1:      	mov	qword ptr [rbp - 0x90], rdi
  3a3cd8:      	mov	rdi, qword ptr [rbp - 0x68]
  3a3cdc:      	call	qword ptr [rax + 0x28]
  3a3cdf:      	mov	r10, qword ptr [rbp - 0x30]
  3a3ce3:      	mov	r9, qword ptr [rbp - 0x48]
  3a3ce7:      	mov	rsi, r14
  3a3cea:      	test	rax, rax
  3a3ced:      	je	0x3a3d3a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x153a>
  3a3cef:      	cmp	qword ptr [rax + 0x8], 0x0
  3a3cf4:      	mov	r11, qword ptr [rbp - 0x38]
  3a3cf8:      	je	0x3a3d3e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x153e>
  3a3cfa:      	jmp	0x3a421c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a1c>
  3a3cff:      	mov	r9, qword ptr [rbp - 0x48]
  3a3d03:      	mov	rsi, qword ptr [r8 + 0xe8]
  3a3d0a:      	add	edi, dword ptr [r12 + 0x18]
  3a3d0f:      	cmp	rsi, rdi
  3a3d12:      	ja	0x3a3592 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd92>
  3a3d18:      	jmp	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3d1d:      	mov	edi, 0x1
  3a3d22:      	mov	r14, rsi
  3a3d25:      	call	0x5e7260 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a3d2a:      	mov	rsi, r14
  3a3d2d:      	mov	r9, qword ptr [rbp - 0x48]
  3a3d31:      	mov	r10, qword ptr [rbp - 0x30]
  3a3d35:      	test	rax, rax
  3a3d38:      	jne	0x3a3cef <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14ef>
  3a3d3a:      	mov	r11, qword ptr [rbp - 0x38]
  3a3d3e:      	mov	rax, qword ptr [r10 + 0x108]
  3a3d45:      	lea	edx, [rsi + r12]
  3a3d49:      	cmp	rax, rdx
  3a3d4c:      	jb	0x3a442b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c2b>
  3a3d52:      	cmp	r12d, edx
  3a3d55:      	ja	0x3a441c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c1c>
  3a3d5b:      	sub	rax, r12
  3a3d5e:      	lea	rcx, [8*r12]
  3a3d66:      	test	rax, rax
  3a3d69:      	mov	eax, 0x0
  3a3d6e:      	cmovle	rcx, rax
  3a3d72:      	test	esi, esi
  3a3d74:      	je	0x3a3dc3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15c3>
  3a3d76:      	mov	rax, qword ptr [r10 + 0xf8]
  3a3d7d:      	sub	rdx, r12
  3a3d80:      	add	rcx, rax
  3a3d83:      	add	rcx, 0x4
  3a3d87:      	xor	esi, esi
  3a3d89:      	nop	dword ptr [rax]
  3a3d90:      	mov	edi, dword ptr [rcx + 8*rsi]
  3a3d93:      	cmp	r13, rdi
  3a3d96:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3d9c:      	mov	eax, dword ptr [rcx + 8*rsi - 0x4]
  3a3da0:      	cmp	r13, rax
  3a3da3:      	jbe	0x3a4352 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b52>
  3a3da9:      	shl	rdi, 0x4
  3a3dad:      	shl	rax, 0x4
  3a3db1:      	movups	xmm0, xmmword ptr [r9 + rdi]
  3a3db6:      	movups	xmmword ptr [r9 + rax], xmm0
  3a3dbb:      	inc	rsi
  3a3dbe:      	cmp	rdx, rsi
  3a3dc1:      	jne	0x3a3d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1590>
  3a3dc3:      	mov	rcx, qword ptr [rbp - 0x58]
  3a3dc7:      	mov	esi, 0xffffffff
  3a3dcc:      	jmp	0x3a28fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a3dd1:      	cmp	eax, 0x3
  3a3dd4:      	je	0x3a419f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x199f>
  3a3dda:      	cmp	eax, 0x6
  3a3ddd:      	je	0x3a4145 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1945>
  3a3de3:      	lea	eax, [r9 - 0x12]
  3a3de7:      	cmp	eax, 0x4
  3a3dea:      	jb	0x3a4175 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1975>
  3a3df0:      	mov	edi, dword ptr [r12 + 0x14]
  3a3df5:      	cmp	rdi, rsi
  3a3df8:      	je	0x3a3e5b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x165b>
  3a3dfa:      	cmp	r13, rdi
  3a3dfd:      	jbe	0x3a433a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b3a>
  3a3e03:      	shl	rdi, 0x4
  3a3e07:      	mov	rax, qword ptr [rbp - 0x48]
  3a3e0b:      	mov	r15, qword ptr [rax + rdi]
  3a3e0f:      	mov	eax, r9d
  3a3e12:      	and	eax, 0xffff0000
  3a3e17:      	cmp	eax, 0xfe0000
  3a3e1c:      	jne	0x3a3e6d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x166d>
  3a3e1e:      	movzx	r9d, r9w
  3a3e22:      	cmp	r9d, 0x3
  3a3e26:      	jne	0x3a4033 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1833>
  3a3e2c:      	test	r11, r11
  3a3e2f:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3e35:      	mov	rax, qword ptr [r11]
  3a3e38:      	test	rax, rax
  3a3e3b:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3e41:      	mov	rdi, qword ptr [rax + 0x8]
  3a3e45:      	call	0x3c4030 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a3e4a:      	mov	r11, qword ptr [rbp - 0x38]
  3a3e4e:      	xor	eax, eax
  3a3e50:      	xor	ecx, ecx
  3a3e52:      	mov	r8, qword ptr [rbp - 0x30]
  3a3e56:      	jmp	0x3a3c74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1474>
  3a3e5b:      	xor	r15d, r15d
  3a3e5e:      	mov	eax, r9d
  3a3e61:      	and	eax, 0xffff0000
  3a3e66:      	cmp	eax, 0xfe0000
  3a3e6b:      	je	0x3a3e1e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x161e>
  3a3e6d:      	add	r9d, 0xff03fff8
  3a3e74:      	cmp	r9d, 0xa
  3a3e78:      	jae	0x3a447a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7a>
  3a3e7e:      	mov	rcx, qword ptr [rbp - 0x30]
  3a3e82:      	mov	rsi, qword ptr [rcx + 0x80]
  3a3e89:      	mov	edi, dword ptr [r12 + 0x4]
  3a3e8e:      	cmp	rsi, rdi
  3a3e91:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a3e97:      	imul	rax, rdi, 0xe0
  3a3e9e:      	add	rax, qword ptr [rcx + 0x78]
  3a3ea2:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a3ea8:      	movups	xmm0, xmmword ptr [rax + 0x98]
  3a3eaf:      	movups	xmm1, xmmword ptr [rax + 0x88]
  3a3eb6:      	movups	xmm2, xmmword ptr [rax + 0x78]
  3a3eba:      	movups	xmm3, xmmword ptr [rax + 0x68]
  3a3ebe:      	movups	xmm4, xmmword ptr [rax + 0x58]
  3a3ec2:      	movups	xmm5, xmmword ptr [rax + 0x48]
  3a3ec6:      	movzx	ecx, byte ptr [rax + 0x40]
  3a3eca:      	movups	xmm6, xmmword ptr [rax + 0x30]
  3a3ece:      	movups	xmm7, xmmword ptr [rax + 0x20]
  3a3ed2:      	movups	xmm8, xmmword ptr [rax + 0x10]
  3a3ed7:      	mov	rsi, qword ptr [rax + 0x8]
  3a3edb:      	mov	eax, dword ptr [rax]
  3a3edd:      	mov	dword ptr [rbp - 0x208], eax
  3a3ee3:      	mov	qword ptr [rbp - 0x200], rsi
  3a3eea:      	movups	xmmword ptr [rbp - 0x1f8], xmm8
  3a3ef2:      	movups	xmmword ptr [rbp - 0x1e8], xmm7
  3a3ef9:      	movups	xmmword ptr [rbp - 0x1d8], xmm6
  3a3f00:      	mov	byte ptr [rbp - 0x1c8], cl
  3a3f06:      	movups	xmmword ptr [rbp - 0x1c0], xmm5
  3a3f0d:      	movups	xmmword ptr [rbp - 0x1b0], xmm4
  3a3f14:      	movups	xmmword ptr [rbp - 0x1a0], xmm3
  3a3f1b:      	movups	xmmword ptr [rbp - 0x190], xmm2
  3a3f22:      	movups	xmmword ptr [rbp - 0x180], xmm1
  3a3f29:      	movups	xmmword ptr [rbp - 0x170], xmm0
  3a3f30:      	mov	qword ptr [rbp - 0xf0], rdx
  3a3f37:      	mov	qword ptr [rbp - 0xe8], 0x0
  3a3f42:      	mov	qword ptr [rbp - 0xe0], r15
  3a3f49:      	mov	qword ptr [rbp - 0xd8], 0x0
  3a3f54:      	mov	ecx, 0x15
  3a3f59:      	mov	rdi, rsp
  3a3f5c:      	lea	rsi, [rbp - 0x208]
  3a3f63:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a3f66:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a3f6d:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a3f75:      	movups	xmm0, xmmword ptr [rbp - 0xf0]
  3a3f7c:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a3f84:      	lea	rdi, [rbp - 0x118]
  3a3f8b:      	mov	rsi, r11
  3a3f8e:      	mov	rcx, r10
  3a3f91:      	mov	rdx, qword ptr [rbp - 0x50]
  3a3f95:      	call	0x3a9420 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a3f9a:      	mov	rax, qword ptr [rbp - 0x118]
  3a3fa1:      	mov	rcx, qword ptr [rbp - 0x110]
  3a3fa8:      	movzx	r15d, byte ptr [rbp - 0x108]
  3a3fb0:      	mov	rsi, qword ptr [rbp - 0x100]
  3a3fb7:      	test	rsi, rsi
  3a3fba:      	je	0x3a410e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x190e>
  3a3fc0:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a3fc5:      	mov	r11, qword ptr [rbp - 0x38]
  3a3fc9:      	mov	r8, qword ptr [rbp - 0x30]
  3a3fcd:      	je	0x3a3c6b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x146b>
  3a3fd3:      	jmp	0x3a431a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b1a>
  3a3fd8:      	test	rax, rax
  3a3fdb:      	je	0x3a411c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x191c>
  3a3fe1:      	mov	rax, qword ptr [rax]
  3a3fe4:      	test	rdi, rdi
  3a3fe7:      	je	0x3a4129 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1929>
  3a3fed:      	mov	rcx, qword ptr [rdi + 0x38]
  3a3ff1:      	test	rcx, rcx
  3a3ff4:      	mov	edx, 0x10000
  3a3ff9:      	cmove	rcx, rdx
  3a3ffd:      	mov	rdx, rax
  3a4000:      	or	rdx, rcx
  3a4003:      	shr	rdx, 0x20
  3a4007:      	je	0x3a4113 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1913>
  3a400d:      	xor	edx, edx
  3a400f:      	div	rcx
  3a4012:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a4017:      	xor	eax, eax
  3a4019:      	xor	edx, edx
  3a401b:      	test	rax, rax
  3a401e:      	jne	0x3a3cef <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14ef>
  3a4024:      	jmp	0x3a3d3a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x153a>
  3a4029:      	mov	eax, 0x20
  3a402e:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a4033:      	mov	rax, qword ptr [rbp - 0x30]
  3a4037:      	mov	rsi, qword ptr [rax + 0x80]
  3a403e:      	mov	edi, dword ptr [r12 + 0x4]
  3a4043:      	cmp	rsi, rdi
  3a4046:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a404c:      	imul	r11, rdi, 0xe0
  3a4053:      	add	r11, qword ptr [rax + 0x78]
  3a4057:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a405d:      	mov	rax, qword ptr [rbp - 0x50]
  3a4061:      	mov	rsi, qword ptr [rax + 0x78]
  3a4065:      	mov	edi, dword ptr [r11 + 0x18]
  3a4069:      	cmp	rsi, rdi
  3a406c:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a4072:      	shl	rdi, 0x3
  3a4076:      	add	rdi, qword ptr [rax + 0x70]
  3a407a:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a4080:      	mov	rsi, qword ptr [rdi]
  3a4083:      	test	rsi, rsi
  3a4086:      	je	0x3a41bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bd>
  3a408c:      	mov	rdi, qword ptr [rbp - 0x38]
  3a4090:      	mov	rcx, qword ptr [rdi]
  3a4093:      	mov	rax, qword ptr [rdi + 0x8]
  3a4097:      	mov	r8, qword ptr [rcx + 0x8]
  3a409b:      	mov	rcx, qword ptr [rdi + 0x10]
  3a409f:      	mov	rsi, qword ptr [rsi + 0x10]
  3a40a3:      	mov	edi, dword ptr [r11 + 0x14]
  3a40a7:      	mov	r11d, dword ptr [r11 + 0x1c]
  3a40ab:      	shl	r11, 0x20
  3a40af:      	or	r11, rdi
  3a40b2:      	mov	qword ptr [rsp + 0x18], r15
  3a40b7:      	mov	qword ptr [rsp + 0x10], rdx
  3a40bc:      	mov	qword ptr [rsp + 0x8], r11
  3a40c1:      	mov	qword ptr [rsp], r10
  3a40c5:      	lea	rdi, [rbp - 0x140]
  3a40cc:      	mov	rdx, rax
  3a40cf:      	call	0x3c5340 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a40d4:      	mov	rax, qword ptr [rbp - 0x140]
  3a40db:      	mov	rcx, qword ptr [rbp - 0x138]
  3a40e2:      	movzx	r15d, byte ptr [rbp - 0x130]
  3a40ea:      	mov	rsi, qword ptr [rbp - 0x128]
  3a40f1:      	test	rsi, rsi
  3a40f4:      	je	0x3a4140 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1940>
  3a40f6:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a40fb:      	mov	r11, qword ptr [rbp - 0x38]
  3a40ff:      	mov	r8, qword ptr [rbp - 0x30]
  3a4103:      	je	0x3a3c6b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x146b>
  3a4109:      	jmp	0x3a4323 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b23>
  3a410e:      	jmp	0x3a3c63 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1463>
  3a4113:      	xor	edx, edx
  3a4115:      	div	ecx
  3a4117:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a411c:      	mov	rax, qword ptr [rdi + 0x8]
  3a4120:      	test	rdi, rdi
  3a4123:      	jne	0x3a3fed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17ed>
  3a4129:      	mov	edi, 0x1
  3a412e:      	call	0x5e7260 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a4133:      	mov	r8, qword ptr [rbp - 0x30]
  3a4137:      	mov	r11, qword ptr [rbp - 0x38]
  3a413b:      	jmp	0x3a37e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe0>
  3a4140:      	jmp	0x3a3c63 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1463>
  3a4145:      	mov	rax, qword ptr [rbp - 0x60]
  3a4149:      	mov	dword ptr [rax - 0x38], r14d
  3a414d:      	mov	rdi, r11
  3a4150:      	mov	r14, r11
  3a4153:      	call	0x3a6860 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a4158:      	test	al, al
  3a415a:      	jne	0x3a4305 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b05>
  3a4160:      	mov	rcx, qword ptr [r14 + 0x48]
  3a4164:      	test	rcx, rcx
  3a4167:      	mov	r11, r14
  3a416a:      	jg	0x3a283d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x3d>
  3a4170:      	jmp	0x3a446b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6b>
  3a4175:      	mov	rax, qword ptr [rbp - 0x60]
  3a4179:      	mov	dword ptr [rax - 0x38], r14d
  3a417d:      	mov	rdi, r11
  3a4180:      	mov	r14, r11
  3a4183:      	call	0x3a5600 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a4188:      	jmp	0x3a4158 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1958>
  3a418a:      	mov	rax, qword ptr [rbp - 0x60]
  3a418e:      	mov	dword ptr [rax - 0x38], r14d
  3a4192:      	mov	rdi, r11
  3a4195:      	mov	r14, r11
  3a4198:      	call	0x3a6360 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a419d:      	jmp	0x3a4158 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1958>
  3a419f:      	mov	rax, qword ptr [rbp - 0x60]
  3a41a3:      	mov	dword ptr [rax - 0x38], r14d
  3a41a7:      	mov	rdi, r11
  3a41aa:      	mov	r14, r11
  3a41ad:      	call	0x3a5170 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a41b2:      	test	al, al
  3a41b4:      	je	0x3a4160 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1960>
  3a41b6:      	mov	al, 0x1
  3a41b8:      	jmp	0x3a4305 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b05>
  3a41bd:      	mov	edi, 0x1
  3a41c2:      	call	0x5e7260 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a41c7:      	jmp	0x3a4305 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b05>
  3a41cc:      	mov	edi, 0x28
  3a41d1:      	call	0x5fd830 <GC_malloc@plt>
  3a41d6:      	test	rax, rax
  3a41d9:      	je	0x3a4489 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c89>
  3a41df:      	mov	r14, rax
  3a41e2:      	add	r15, r13
  3a41e5:      	mov	edi, 0x28
  3a41ea:      	call	0x5e2ef0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a41ef:      	mov	rax, qword ptr [r15 + 0x8]
  3a41f3:      	movaps	xmm0, xmmword ptr [rip + 0x317046] # 0x6bb240 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a41fa:      	movups	xmmword ptr [r14], xmm0
  3a41fe:      	mov	qword ptr [r14 + 0x10], rax
  3a4202:      	mov	qword ptr [r14 + 0x18], 0x299e78
  3a420a:      	mov	qword ptr [r14 + 0x20], 0xe
  3a4212:      	mov	esi, 0x2ba878
  3a4217:      	jmp	0x3a42f7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1af7>
  3a421c:      	mov	rdi, r11
  3a421f:      	mov	rsi, rax
  3a4222:      	jmp	0x3a42fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1afe>
  3a4227:      	mov	r15d, edx
  3a422a:      	jmp	0x3a4291 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a422c:      	mov	r14, qword ptr [rbp - 0x148]
  3a4233:      	mov	edi, 0x8
  3a4238:      	call	0x5fd830 <GC_malloc@plt>
  3a423d:      	test	rax, rax
  3a4240:      	je	0x3a4489 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c89>
  3a4246:      	mov	r12, rax
  3a4249:      	mov	edi, 0x8
  3a424e:      	call	0x5e2ef0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a4253:      	mov	edx, 0x299de8
  3a4258:      	mov	rdi, r15
  3a425b:      	mov	rsi, r14
  3a425e:      	mov	rcx, r12
  3a4261:      	call	0x5b9e00 <errors.As>
  3a4266:      	test	al, 0x1
  3a4268:      	je	0x3a427b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a7b>
  3a426a:      	mov	rax, qword ptr [r12]
  3a426e:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a4275:      	mov	ecx, dword ptr [rcx + 0x50]
  3a4278:      	mov	dword ptr [rax + 0x4], ecx
  3a427b:      	mov	rdi, qword ptr [rbp - 0x38]
  3a427f:      	mov	rsi, r15
  3a4282:      	jmp	0x3a42fb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1afb>
  3a4284:      	mov	r15b, 0x3
  3a4287:      	jmp	0x3a4291 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a4289:      	mov	r15b, 0x1
  3a428c:      	jmp	0x3a4291 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a91>
  3a428e:      	mov	r15b, 0x9
  3a4291:      	mov	edi, 0x10
  3a4296:      	call	0x5fd8e0 <GC_malloc_atomic@plt>
  3a429b:      	test	rax, rax
  3a429e:      	je	0x3a4489 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c89>
  3a42a4:      	mov	r14, rax
  3a42a7:      	mov	edi, 0x10
  3a42ac:      	call	0x5e2ef0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a42b1:      	xorps	xmm0, xmm0
  3a42b4:      	movups	xmmword ptr [r14], xmm0
  3a42b8:      	mov	rax, qword ptr [rbp - 0x30]
  3a42bc:      	mov	rsi, qword ptr [rax + 0x80]
  3a42c3:      	mov	edi, dword ptr [r12 + 0x4]
  3a42c8:      	cmp	rsi, rdi
  3a42cb:      	jbe	0x3a4342 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b42>
  3a42cd:      	mov	rax, qword ptr [rax + 0x78]
  3a42d1:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a42d8:      	mov	ecx, dword ptr [rcx + 0x50]
  3a42db:      	imul	rdx, rdi, 0xe0
  3a42e2:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a42e7:      	mov	byte ptr [r14], r15b
  3a42ea:      	mov	dword ptr [r14 + 0x4], ecx
  3a42ee:      	mov	qword ptr [r14 + 0x8], rax
  3a42f2:      	mov	esi, 0x29a200
  3a42f7:      	mov	rdi, qword ptr [rbp - 0x38]
  3a42fb:      	mov	rdx, r14
  3a42fe:      	call	0x3a21d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a4303:      	mov	al, 0x3
  3a4305:      	mov	rdx, rbx
  3a4308:      	add	rsp, 0x2a8
  3a430f:      	pop	rbx
  3a4310:      	pop	r12
  3a4312:      	pop	r13
  3a4314:      	pop	r14
  3a4316:      	pop	r15
  3a4318:      	pop	rbp
  3a4319:      	ret
  3a431a:      	mov	rdx, qword ptr [rbp - 0xf8]
  3a4321:      	jmp	0x3a432a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b2a>
  3a4323:      	mov	rdx, qword ptr [rbp - 0x120]
  3a432a:      	mov	rdi, r11
  3a432d:      	jmp	0x3a42fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1afe>
  3a432f:      	mov	rdi, r14
  3a4332:      	mov	rsi, r15
  3a4335:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a433a:      	mov	rsi, r13
  3a433d:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4342:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4347:      	mov	rdi, r12
  3a434a:      	mov	rsi, r15
  3a434d:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4352:      	mov	rdi, rax
  3a4355:      	mov	rsi, r13
  3a4358:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a435d:      	mov	rdi, rcx
  3a4360:      	mov	rsi, r13
  3a4363:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4368:      	mov	rsi, r13
  3a436b:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4370:      	mov	rdi, r9
  3a4373:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4378:      	mov	rdi, rax
  3a437b:      	mov	rsi, r13
  3a437e:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a4383:      	mov	rdi, rax
  3a4386:      	mov	esi, 0x1
  3a438b:      	mov	ecx, 0x2
  3a4390:      	call	0x5e3d10 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a4395:      	mov	rdi, r9
  3a4398:      	mov	esi, 0x1
  3a439d:      	mov	rdx, rax
  3a43a0:      	mov	ecx, 0x3
  3a43a5:      	call	0x5e3d10 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a43aa:      	mov	rsi, rcx
  3a43ad:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a43b2:      	mov	rdi, rsi
  3a43b5:      	mov	rsi, rcx
  3a43b8:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a43bd:      	mov	rdi, rcx
  3a43c0:      	mov	rsi, r13
  3a43c3:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a43c8:      	mov	edi, 0x1
  3a43cd:      	mov	esi, 0x1
  3a43d2:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a43d7:      	xor	edi, edi
  3a43d9:      	xor	esi, esi
  3a43db:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a43e0:      	xor	esi, esi
  3a43e2:      	mov	rdx, rcx
  3a43e5:      	mov	ecx, 0x3
  3a43ea:      	call	0x5e3d10 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a43ef:      	mov	rdi, rcx
  3a43f2:      	xor	esi, esi
  3a43f4:      	mov	ecx, 0x2
  3a43f9:      	call	0x5e3d10 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a43fe:      	mov	edi, 0x3
  3a4403:      	mov	esi, 0x3
  3a4408:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a440d:      	mov	edi, 0x2
  3a4412:      	mov	esi, 0x2
  3a4417:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a441c:      	mov	rdi, r12
  3a441f:      	xor	esi, esi
  3a4421:      	mov	ecx, 0x3
  3a4426:      	call	0x5e3d10 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a442b:      	mov	rdi, rdx
  3a442e:      	xor	esi, esi
  3a4430:      	mov	rdx, rax
  3a4433:      	mov	ecx, 0x2
  3a4438:      	call	0x5e3d10 <github.com/xgo-dev/llgo/runtime/internal/runtime.panicBounds>
  3a443d:      	mov	rdi, rdx
  3a4440:      	mov	rsi, r13
  3a4443:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4448:      	mov	rdi, rcx
  3a444b:      	mov	rsi, qword ptr [rbp - 0x78]
  3a444f:      	call	0x5e2fd0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a4454:      	mov	rdi, rax
  3a4457:      	call	0x5e4500 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a445c:      	mov	edi, 0x3293b0
  3a4461:      	mov	esi, 0x3296e0
  3a4466:      	call	0x5e2c10 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a446b:      	mov	edi, 0x333868
  3a4470:      	mov	esi, 0x299e68
  3a4475:      	call	0x5e2c10 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a447a:      	mov	edi, 0x333868
  3a447f:      	mov	esi, 0x299e88
  3a4484:      	call	0x5e2c10 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a4489:      	mov	edi, 0x333868
  3a448e:      	mov	esi, 0x329388
  3a4493:      	call	0x5e2c10 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a4498:      	int3
  3a4499:      	int3
  3a449a:      	int3
  3a449b:      	int3
  3a449c:      	int3
  3a449d:      	int3
  3a449e:      	int3
  3a449f:      	int3
