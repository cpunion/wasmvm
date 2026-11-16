
/home/runner/work/wasmvm/wasmvm/main-regression/dense-fix:	file format elf64-x86-64

Disassembly of section .text:

00000000003a16b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a16b0:      	push	rbp
  3a16b1:      	mov	rbp, rsp
  3a16b4:      	push	r15
  3a16b6:      	push	r14
  3a16b8:      	push	r13
  3a16ba:      	push	r12
  3a16bc:      	push	rbx
  3a16bd:      	sub	rsp, 0x2f8
  3a16c4:      	mov	rcx, qword ptr [rdi + 0x48]
  3a16c8:      	test	rcx, rcx
  3a16cb:      	jle	0x3a3685 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd5>
  3a16d1:      	mov	rbx, rsi
  3a16d4:      	mov	r11, rdi
  3a16d7:      	movzx	r15d, byte ptr [rdi + 0x20]
  3a16dc:      	mov	r10d, 0xffffffff
  3a16e2:      	lea	rax, [rbp - 0xb0]
  3a16e9:      	mov	qword ptr [rbp - 0x80], rax
  3a16ed:      	mov	qword ptr [rbp - 0x48], rdi
  3a16f1:      	mov	byte ptr [rbp - 0x29], r15b
  3a16f5:      	mov	r8, qword ptr [r11 + 0x40]
  3a16f9:      	shl	rcx, 0x6
  3a16fd:      	mov	byte ptr [r8 + rcx - 0x30], 0x0
  3a1703:      	mov	rsi, qword ptr [r8 + rcx - 0x40]
  3a1708:      	mov	rax, qword ptr [rsi + 0x48]
  3a170c:      	mov	qword ptr [rbp - 0x60], rax
  3a1710:      	mov	r9, qword ptr [rax + 0x98]
  3a1717:      	mov	r13, qword ptr [r8 + rcx - 0x28]
  3a171c:      	mov	rdx, qword ptr [r11 + 0x88]
  3a1723:      	lea	rax, [r9 + r13]
  3a1727:      	cmp	rax, rdx
  3a172a:      	ja	0x3a35b7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f07>
  3a1730:      	cmp	r13, rax
  3a1733:      	ja	0x3a35c9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f19>
  3a1739:      	mov	rdi, qword ptr [rbp - 0x60]
  3a173d:      	mov	r14, qword ptr [rdi + 0xb8]
  3a1744:      	mov	r12d, dword ptr [r8 + rcx - 0x38]
  3a1749:      	sub	rdx, r13
  3a174c:      	test	rdx, rdx
  3a174f:      	mov	eax, 0x0
  3a1754:      	cmovle	r13, rax
  3a1758:      	mov	qword ptr [rbp - 0x68], r14
  3a175c:      	cmp	r14, r12
  3a175f:      	jbe	0x3a35a8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ef8>
  3a1765:      	add	r8, rcx
  3a1768:      	mov	qword ptr [rbp - 0x70], r8
  3a176c:      	mov	qword ptr [rbp - 0x88], rsi
  3a1773:      	mov	rax, qword ptr [rsi + 0x40]
  3a1777:      	mov	qword ptr [rbp - 0x58], rax
  3a177b:      	mov	rcx, qword ptr [rdi + 0xb0]
  3a1782:      	shl	r13, 0x4
  3a1786:      	add	r13, qword ptr [r11 + 0x78]
  3a178a:      	mov	r14d, r12d
  3a178d:      	mov	qword ptr [rbp - 0x38], r9
  3a1791:      	mov	qword ptr [rbp - 0x50], r13
  3a1795:      	mov	r8, qword ptr [rbp - 0x68]
  3a1799:      	mov	qword ptr [rbp - 0x40], rcx
  3a179d:      	jmp	0x3a17cb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11b>
  3a179f:      	nop
  3a17a0:      	shl	rdi, 0x4
  3a17a4:      	mov	qword ptr [r13 + rdi], rax
  3a17a9:      	mov	qword ptr [r13 + rdi + 0x8], r8
  3a17ae:      	inc	r14d
  3a17b1:      	mov	r13d, r14d
  3a17b4:      	mov	r8, qword ptr [rbp - 0x68]
  3a17b8:      	mov	r12d, r13d
  3a17bb:      	mov	r14d, r13d
  3a17be:      	cmp	r8, r12
  3a17c1:      	mov	r13, qword ptr [rbp - 0x50]
  3a17c5:      	jbe	0x3a35ac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1efc>
  3a17cb:      	shl	r12, 0x5
  3a17cf:      	add	r12, rcx
  3a17d2:      	test	r15b, 0x1
  3a17d6:      	je	0x3a1836 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x186>
  3a17d8:      	mov	rax, qword ptr [rbp - 0x60]
  3a17dc:      	mov	rsi, qword ptr [rax + 0x80]
  3a17e3:      	mov	edi, dword ptr [r12 + 0x4]
  3a17e8:      	cmp	rsi, rdi
  3a17eb:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a17f1:      	mov	r13, qword ptr [rax + 0x78]
  3a17f5:      	imul	r15, rdi, 0xe0
  3a17fc:      	mov	eax, dword ptr [r13 + r15]
  3a1801:      	cmp	eax, 0x5
  3a1804:      	setne	cl
  3a1807:      	cmp	eax, 0xb
  3a180a:      	setne	al
  3a180d:      	and	al, cl
  3a180f:      	movzx	ecx, al
  3a1812:      	mov	rax, qword ptr [r11 + 0x150]
  3a1819:      	sub	rax, rcx
  3a181c:      	jb	0x3a2e80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x17d0>
  3a1822:      	mov	qword ptr [r11 + 0x150], rax
  3a1829:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a182e:      	mov	r13, qword ptr [rbp - 0x50]
  3a1832:      	mov	rcx, qword ptr [rbp - 0x40]
  3a1836:      	inc	rbx
  3a1839:      	test	bl, 0x3f
  3a183c:      	jne	0x3a18bc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20c>
  3a183e:      	mov	rax, qword ptr [r11 + 0x160]
  3a1845:      	test	rax, rax
  3a1848:      	je	0x3a185d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ad>
  3a184a:      	test	r11, r11
  3a184d:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a1853:      	mov	eax, dword ptr [rax]
  3a1855:      	test	eax, eax
  3a1857:      	je	0x3a1953 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2a3>
  3a185d:      	mov	rax, qword ptr [r11 + 0x8]
  3a1861:      	test	rax, rax
  3a1864:      	je	0x3a3641 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f91>
  3a186a:      	mov	rdi, qword ptr [r11 + 0x10]
  3a186e:      	mov	rcx, qword ptr [rax + 0x8]
  3a1872:      	test	byte ptr [rcx + 0x17], 0x20
  3a1876:      	je	0x3a188f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1df>
  3a1878:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a187c:      	and	cl, 0x1f
  3a187f:      	cmp	cl, 0x16
  3a1882:      	je	0x3a188f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1df>
  3a1884:      	mov	qword ptr [rbp - 0xb0], rdi
  3a188b:      	mov	rdi, qword ptr [rbp - 0x80]
  3a188f:      	call	qword ptr [rax + 0x28]
  3a1892:      	mov	r11, qword ptr [rbp - 0x48]
  3a1896:      	mov	r10d, 0xffffffff
  3a189c:      	mov	r9, qword ptr [rbp - 0x38]
  3a18a0:      	mov	r8, qword ptr [rbp - 0x68]
  3a18a4:      	mov	rcx, qword ptr [rbp - 0x40]
  3a18a8:      	test	rax, rax
  3a18ab:      	je	0x3a1960 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2b0>
  3a18b1:      	cmp	qword ptr [rax + 0x8], 0x0
  3a18b6:      	jne	0x3a300a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x195a>
  3a18bc:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a18c2:      	je	0x3a1970 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a18c8:      	test	r15b, 0x1
  3a18cc:      	je	0x3a18f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a18ce:      	mov	edx, dword ptr [r12]
  3a18d2:      	cmp	edx, 0x20
  3a18d5:      	je	0x3a1970 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a18db:      	cmp	edx, 0x41
  3a18de:      	je	0x3a1970 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a18e4:      	jmp	0x3a1974 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c4>
  3a18e9:      	nop	dword ptr [rax]
  3a18f0:      	mov	edx, dword ptr [r12]
  3a18f4:      	cmp	edx, 0x20
  3a18f7:      	je	0x3a18fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x24e>
  3a18f9:      	cmp	edx, 0x41
  3a18fc:      	jne	0x3a1974 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c4>
  3a18fe:      	mov	eax, ebx
  3a1900:      	not	eax
  3a1902:      	test	al, 0x3f
  3a1904:      	je	0x3a1970 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a1906:      	cmp	edx, 0x41
  3a1909:      	jne	0x3a1930 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x280>
  3a190b:      	mov	edi, dword ptr [r12 + 0x8]
  3a1910:      	cmp	r9, rdi
  3a1913:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a1919:      	mov	eax, dword ptr [r12 + 0x18]
  3a191e:      	shl	rdi, 0x4
  3a1922:      	mov	qword ptr [r13 + rdi], rax
  3a1927:      	mov	qword ptr [r13 + rdi + 0x8], 0x0
  3a1930:      	inc	r14d
  3a1933:      	cmp	r8, r14
  3a1936:      	jbe	0x3a3311 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c61>
  3a193c:      	mov	rax, r14
  3a193f:      	shl	rax, 0x5
  3a1943:      	lea	r12, [rcx + rax]
  3a1947:      	inc	rbx
  3a194a:      	cmp	dword ptr [rcx + rax + 0x1c], 0x0
  3a194f:      	jne	0x3a18f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a1951:      	jmp	0x3a1970 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a1953:      	xor	eax, eax
  3a1955:      	xor	edx, edx
  3a1957:      	test	rax, rax
  3a195a:      	jne	0x3a18b1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x201>
  3a1960:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a1966:      	jne	0x3a18c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x218>
  3a196c:      	nop	dword ptr [rax]
  3a1970:      	mov	edx, dword ptr [r12]
  3a1974:      	lea	eax, [rdx - 0xe0]
  3a197a:      	cmp	eax, 0xffff20
  3a197f:      	jae	0x3a1c60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5b0>
  3a1985:      	mov	eax, edx
  3a1987:      	shr	eax, 0x10
  3a198a:      	cmp	eax, 0xfb
  3a198f:      	je	0x3a1b65 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4b5>
  3a1995:      	cmp	eax, 0xfc
  3a199a:      	je	0x3a1a3d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x38d>
  3a19a0:      	cmp	eax, 0xfd
  3a19a5:      	jne	0x3a1bde <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x52e>
  3a19ab:      	mov	edi, dword ptr [r12 + 0xc]
  3a19b0:      	mov	esi, 0xffffffff
  3a19b5:      	cmp	rdi, r10
  3a19b8:      	je	0x3a1cf8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x648>
  3a19be:      	cmp	r9, rdi
  3a19c1:      	mov	r15, qword ptr [rbp - 0x60]
  3a19c5:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a19cb:      	shl	rdi, 0x4
  3a19cf:      	mov	rcx, qword ptr [r13 + rdi]
  3a19d4:      	mov	r10, qword ptr [r13 + rdi + 0x8]
  3a19d9:      	mov	edi, dword ptr [r12 + 0x10]
  3a19de:      	cmp	rdi, rsi
  3a19e1:      	je	0x3a1d0f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x65f>
  3a19e7:      	cmp	r9, rdi
  3a19ea:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a19f0:      	shl	rdi, 0x4
  3a19f4:      	mov	rax, qword ptr [r13 + rdi]
  3a19f9:      	mov	r8, qword ptr [r13 + rdi + 0x8]
  3a19fe:      	mov	edi, dword ptr [r12 + 0x14]
  3a1a03:      	cmp	rdi, rsi
  3a1a06:      	je	0x3a1d22 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x672>
  3a1a0c:      	cmp	r9, rdi
  3a1a0f:      	jbe	0x3a332e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a1a15:      	shl	rdi, 0x4
  3a1a19:      	mov	r9, qword ptr [r13 + rdi]
  3a1a1e:      	mov	r11, qword ptr [r13 + rdi + 0x8]
  3a1a23:      	mov	rsi, qword ptr [r15 + 0x80]
  3a1a2a:      	mov	edi, dword ptr [r12 + 0x4]
  3a1a2f:      	cmp	rsi, rdi
  3a1a32:      	ja	0x3a1d3d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x68d>
  3a1a38:      	jmp	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a1a3d:      	add	edx, 0xff03ffed
  3a1a43:      	cmp	edx, 0x4
  3a1a46:      	jae	0x3a1c2a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x57a>
  3a1a4c:      	mov	rax, qword ptr [rbp - 0x60]
  3a1a50:      	mov	rdx, qword ptr [rax + 0xd8]
  3a1a57:      	mov	edi, dword ptr [r12 + 0x18]
  3a1a5c:      	mov	eax, dword ptr [r12 + 0x1c]
  3a1a61:      	lea	esi, [rdi + rax]
  3a1a64:      	add	esi, 0x2
  3a1a67:      	cmp	rdx, rsi
  3a1a6a:      	jb	0x3a358e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ede>
  3a1a70:      	cmp	edi, esi
  3a1a72:      	ja	0x3a357f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ecf>
  3a1a78:      	mov	eax, edi
  3a1a7a:      	mov	ecx, esi
  3a1a7c:      	sub	rsi, rdi
  3a1a7f:      	sub	rdx, rdi
  3a1a82:      	test	rdx, rdx
  3a1a85:      	mov	edx, 0x0
  3a1a8a:      	cmovle	rdi, rdx
  3a1a8e:      	cmp	ecx, eax
  3a1a90:      	je	0x3a3578 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ec8>
  3a1a96:      	mov	rax, qword ptr [rbp - 0x60]
  3a1a9a:      	mov	rax, qword ptr [rax + 0xc8]
  3a1aa1:      	lea	r8, [rax + 4*rdi]
  3a1aa5:      	mov	edi, dword ptr [r8]
  3a1aa8:      	cmp	r9, rdi
  3a1aab:      	jbe	0x3a354a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e9a>
  3a1ab1:      	cmp	rsi, 0x1
  3a1ab5:      	jbe	0x3a356e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ebe>
  3a1abb:      	mov	eax, dword ptr [r8 + 0x4]
  3a1abf:      	cmp	r9, rax
  3a1ac2:      	jbe	0x3a355a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eaa>
  3a1ac8:      	shl	rdi, 0x4
  3a1acc:      	mov	r9, qword ptr [r13 + rdi]
  3a1ad1:      	shl	rax, 0x4
  3a1ad5:      	mov	rcx, qword ptr [r13 + rax]
  3a1ada:      	mov	r10d, dword ptr [r12]
  3a1ade:      	lea	eax, [r10 - 0xfc0013]
  3a1ae5:      	cmp	eax, 0x2
  3a1ae8:      	jae	0x3a1eb1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x801>
  3a1aee:      	cmp	rsi, 0x2
  3a1af2:      	je	0x3a35f1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f41>
  3a1af8:      	mov	edi, dword ptr [r8 + 0x8]
  3a1afc:      	cmp	qword ptr [rbp - 0x38], rdi
  3a1b00:      	jbe	0x3a35de <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f2e>
  3a1b06:      	cmp	rsi, 0x3
  3a1b0a:      	jbe	0x3a35e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f37>
  3a1b10:      	mov	eax, dword ptr [r8 + 0xc]
  3a1b14:      	cmp	qword ptr [rbp - 0x38], rax
  3a1b18:      	jbe	0x3a355a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eaa>
  3a1b1e:      	shl	rdi, 0x4
  3a1b22:      	mov	rdi, qword ptr [r13 + rdi]
  3a1b27:      	shl	rax, 0x4
  3a1b2b:      	mov	rdx, qword ptr [r13 + rax]
  3a1b30:      	cmp	r10d, 0xfc0013
  3a1b37:      	jne	0x3a1f31 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x881>
  3a1b3d:      	lea	rax, [rdi + r9]
  3a1b41:      	mov	r10, rdi
  3a1b44:      	and	r10, r9
  3a1b47:      	or	rdi, r9
  3a1b4a:      	mov	r9, rax
  3a1b4d:      	not	r9
  3a1b50:      	and	r9, rdi
  3a1b53:      	or	r9, r10
  3a1b56:      	shr	r9, 0x3f
  3a1b5a:      	add	rdx, rcx
  3a1b5d:      	add	rdx, r9
  3a1b60:      	jmp	0x3a1f56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8a6>
  3a1b65:      	add	edx, -0x2
  3a1b68:      	cmp	dx, 0x21
  3a1b6c:      	jae	0x3a1beb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x53b>
  3a1b6e:      	lea	rax, [r10 + 0x1c003e10]
  3a1b75:      	movzx	ecx, dx
  3a1b78:      	bt	rax, rcx
  3a1b7c:      	jae	0x3a1beb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x53b>
  3a1b7e:      	mov	edi, dword ptr [r12 + 0xc]
  3a1b83:      	cmp	rdi, r10
  3a1b86:      	je	0x3a22be <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc0e>
  3a1b8c:      	cmp	r9, rdi
  3a1b8f:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a1b95:      	shl	rdi, 0x4
  3a1b99:      	mov	r8, qword ptr [r13 + rdi]
  3a1b9e:      	mov	edi, dword ptr [r12 + 0x10]
  3a1ba3:      	cmp	rdi, r10
  3a1ba6:      	je	0x3a22cf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc1f>
  3a1bac:      	cmp	r9, rdi
  3a1baf:      	jbe	0x3a332e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a1bb5:      	shl	rdi, 0x4
  3a1bb9:      	mov	rax, qword ptr [r13 + rdi]
  3a1bbe:      	mov	r9, qword ptr [r13 + rdi + 0x8]
  3a1bc3:      	mov	r10d, 0x0
  3a1bc9:      	mov	r11d, 0x0
  3a1bcf:      	cmp	dx, 0xc
  3a1bd3:      	je	0x3a22e6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc36>
  3a1bd9:      	jmp	0x3a2324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc74>
  3a1bde:      	cmp	edx, 0xe7
  3a1be4:      	jae	0x3a1c60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5b0>
  3a1be6:      	jmp	0x3a2e04 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1754>
  3a1beb:      	mov	r12, qword ptr [rbp - 0x70]
  3a1bef:      	mov	dword ptr [r12 - 0x38], r14d
  3a1bf4:      	mov	rdi, r11
  3a1bf7:      	call	0x3a38e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a1bfc:      	mov	r8, qword ptr [rbp - 0x68]
  3a1c00:      	mov	r9, qword ptr [rbp - 0x38]
  3a1c04:      	mov	r10d, 0xffffffff
  3a1c0a:      	mov	r11, qword ptr [rbp - 0x48]
  3a1c0e:      	test	al, al
  3a1c10:      	mov	rcx, qword ptr [rbp - 0x40]
  3a1c14:      	jne	0x3a3530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e80>
  3a1c1a:      	inc	r14d
  3a1c1d:      	mov	dword ptr [r12 - 0x38], r14d
  3a1c22:      	mov	r13d, r14d
  3a1c25:      	jmp	0x3a17b8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x108>
  3a1c2a:      	mov	edx, dword ptr [r12]
  3a1c2e:      	cmp	edx, 0xfc0008
  3a1c34:      	jae	0x3a1c60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5b0>
  3a1c36:      	mov	edi, dword ptr [r12 + 0xc]
  3a1c3b:      	cmp	r9, rdi
  3a1c3e:      	jbe	0x3a332e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a1c44:      	shl	rdi, 0x4
  3a1c48:      	mov	rsi, qword ptr [r13 + rdi]
  3a1c4d:      	mov	edi, edx
  3a1c4f:      	xor	edx, edx
  3a1c51:      	call	0x3d0200 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a1c56:      	xor	r8d, r8d
  3a1c59:      	jmp	0x3a24ab <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdfb>
  3a1c5e:      	nop
  3a1c60:      	mov	edi, dword ptr [r12 + 0xc]
  3a1c65:      	cmp	rdi, r10
  3a1c68:      	je	0x3a1cb0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x600>
  3a1c6a:      	cmp	r9, rdi
  3a1c6d:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a1c73:      	shl	rdi, 0x4
  3a1c77:      	mov	rax, qword ptr [r13 + rdi]
  3a1c7c:      	mov	r8, qword ptr [r13 + rdi + 0x8]
  3a1c81:      	mov	edi, dword ptr [r12 + 0x10]
  3a1c86:      	cmp	rdi, r10
  3a1c89:      	je	0x3a1cbf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x60f>
  3a1c8b:      	cmp	r9, rdi
  3a1c8e:      	jbe	0x3a332e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a1c94:      	shl	rdi, 0x4
  3a1c98:      	mov	rcx, qword ptr [r13 + rdi]
  3a1c9d:      	mov	r9, qword ptr [r13 + rdi + 0x8]
  3a1ca2:      	jmp	0x3a1cc4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x614>
  3a1ca4:      	nop	word ptr cs:[rax + rax]
  3a1cb0:      	xor	eax, eax
  3a1cb2:      	xor	r8d, r8d
  3a1cb5:      	mov	edi, dword ptr [r12 + 0x10]
  3a1cba:      	cmp	rdi, r10
  3a1cbd:      	jne	0x3a1c8b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5db>
  3a1cbf:      	xor	ecx, ecx
  3a1cc1:      	xor	r9d, r9d
  3a1cc4:      	mov	esi, edx
  3a1cc6:      	and	esi, 0xffffff
  3a1ccc:      	mov	edi, edx
  3a1cce:      	shr	edi, 0x10
  3a1cd1:      	and	edi, 0xff00
  3a1cd7:      	or	edi, esi
  3a1cd9:      	lea	esi, [rdi - 0x45]
  3a1cdc:      	cmp	esi, 0x39
  3a1cdf:      	ja	0x3a1fd4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x924>
  3a1ce5:      	jmp	qword ptr [8*rsi + 0x203258]
  3a1cec:      	xor	ecx, ecx
  3a1cee:      	test	eax, eax
  3a1cf0:      	sete	cl
  3a1cf3:      	jmp	0x3a229e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbee>
  3a1cf8:      	xor	r10d, r10d
  3a1cfb:      	xor	ecx, ecx
  3a1cfd:      	mov	r15, qword ptr [rbp - 0x60]
  3a1d01:      	mov	edi, dword ptr [r12 + 0x10]
  3a1d06:      	cmp	rdi, rsi
  3a1d09:      	jne	0x3a19e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x337>
  3a1d0f:      	xor	eax, eax
  3a1d11:      	xor	r8d, r8d
  3a1d14:      	mov	edi, dword ptr [r12 + 0x14]
  3a1d19:      	cmp	rdi, rsi
  3a1d1c:      	jne	0x3a1a0c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x35c>
  3a1d22:      	xor	r9d, r9d
  3a1d25:      	xor	r11d, r11d
  3a1d28:      	mov	rsi, qword ptr [r15 + 0x80]
  3a1d2f:      	mov	edi, dword ptr [r12 + 0x4]
  3a1d34:      	cmp	rsi, rdi
  3a1d37:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a1d3d:      	imul	rdi, rdi, 0xe0
  3a1d44:      	add	rdi, qword ptr [r15 + 0x78]
  3a1d48:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a1d4e:      	movzx	esi, dx
  3a1d51:      	cmp	esi, 0xb
  3a1d54:      	setb	r15b
  3a1d58:      	mov	esi, edx
  3a1d5a:      	and	esi, 0xfffc
  3a1d60:      	cmp	esi, 0x54
  3a1d63:      	sete	r13b
  3a1d67:      	or	r13b, r15b
  3a1d6a:      	jne	0x3a1d7e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6ce>
  3a1d6c:      	movzx	r15d, dx
  3a1d70:      	lea	r13d, [r15 - 0x5c]
  3a1d74:      	cmp	r13d, 0x2
  3a1d78:      	jae	0x3a1e0f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x75f>
  3a1d7e:      	mov	r10, qword ptr [rbp - 0x58]
  3a1d82:      	mov	rsi, qword ptr [r10 + 0x78]
  3a1d86:      	mov	r9d, dword ptr [rdi + 0x18]
  3a1d8a:      	cmp	rsi, r9
  3a1d8d:      	jbe	0x3a3552 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ea2>
  3a1d93:      	shl	r9, 0x3
  3a1d97:      	add	r9, qword ptr [r10 + 0x70]
  3a1d9b:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a1da0:      	mov	r11, qword ptr [rbp - 0x80]
  3a1da4:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a1daa:      	mov	rsi, qword ptr [r9]
  3a1dad:      	test	rsi, rsi
  3a1db0:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a1db6:      	mov	rsi, qword ptr [rsi + 0x10]
  3a1dba:      	mov	r9d, dword ptr [rdi + 0x14]
  3a1dbe:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a1dc2:      	shl	r10, 0x20
  3a1dc6:      	or	r10, r9
  3a1dc9:      	movzx	edi, byte ptr [rdi + 0x40]
  3a1dcd:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a1dd2:      	je	0x3a1ddf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x72f>
  3a1dd4:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a1dd9:      	je	0x3a1ede <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x82e>
  3a1ddf:      	mov	qword ptr [rbp - 0x98], rax
  3a1de6:      	mov	qword ptr [rbp - 0x90], r8
  3a1ded:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a1df4:      	movups	xmmword ptr [rsp], xmm0
  3a1df8:      	movzx	r9d, dil
  3a1dfc:      	mov	rdi, r11
  3a1dff:      	mov	r8, r10
  3a1e02:      	mov	r13, r11
  3a1e05:      	call	0x3c6800 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a1e0a:      	jmp	0x3a1f09 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x859>
  3a1e0f:      	cmp	r15d, 0xb
  3a1e13:      	je	0x3a1e1e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x76e>
  3a1e15:      	cmp	esi, 0x58
  3a1e18:      	jne	0x3a2414 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd64>
  3a1e1e:      	mov	r10, qword ptr [rbp - 0x58]
  3a1e22:      	mov	rsi, qword ptr [r10 + 0x78]
  3a1e26:      	mov	r9d, dword ptr [rdi + 0x18]
  3a1e2a:      	cmp	rsi, r9
  3a1e2d:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a1e32:      	mov	r13, qword ptr [rbp - 0x50]
  3a1e36:      	jbe	0x3a3552 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ea2>
  3a1e3c:      	shl	r9, 0x3
  3a1e40:      	add	r9, qword ptr [r10 + 0x70]
  3a1e44:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a1e4a:      	mov	rsi, qword ptr [r9]
  3a1e4d:      	test	rsi, rsi
  3a1e50:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a1e56:      	mov	rsi, qword ptr [rsi + 0x10]
  3a1e5a:      	mov	r10d, dword ptr [rdi + 0x14]
  3a1e5e:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a1e62:      	shl	r9, 0x20
  3a1e66:      	or	r9, r10
  3a1e69:      	movzx	edi, byte ptr [rdi + 0x40]
  3a1e6d:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a1e72:      	je	0x3a1e7f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7cf>
  3a1e74:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a1e79:      	je	0x3a2477 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdc7>
  3a1e7f:      	mov	qword ptr [rbp - 0xb0], rax
  3a1e86:      	mov	qword ptr [rbp - 0xa8], r8
  3a1e8d:      	movups	xmm0, xmmword ptr [rbp - 0xb0]
  3a1e94:      	movups	xmmword ptr [rsp], xmm0
  3a1e98:      	movzx	r8d, dil
  3a1e9c:      	mov	rdi, rsi
  3a1e9f:      	mov	esi, edx
  3a1ea1:      	mov	rdx, rcx
  3a1ea4:      	mov	rcx, r9
  3a1ea7:      	call	0x3c7b70 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a1eac:      	jmp	0x3a24a4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdf4>
  3a1eb1:      	mov	rax, rcx
  3a1eb4:      	mul	r9
  3a1eb7:      	cmp	dword ptr [r12], 0xfc0015
  3a1ebf:      	jne	0x3a1f56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8a6>
  3a1ec5:      	mov	rdi, r9
  3a1ec8:      	sar	rdi, 0x3f
  3a1ecc:      	and	rdi, rcx
  3a1ecf:      	sar	rcx, 0x3f
  3a1ed3:      	and	rcx, r9
  3a1ed6:      	add	rcx, rdi
  3a1ed9:      	sub	rdx, rcx
  3a1edc:      	jmp	0x3a1f56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8a6>
  3a1ede:      	mov	qword ptr [rbp - 0x98], rax
  3a1ee5:      	mov	qword ptr [rbp - 0x90], r8
  3a1eec:      	movups	xmm0, xmmword ptr [rbp - 0x98]
  3a1ef3:      	movups	xmmword ptr [rsp], xmm0
  3a1ef7:      	movzx	r9d, dil
  3a1efb:      	mov	rdi, r11
  3a1efe:      	mov	r8, r10
  3a1f01:      	mov	r13, r11
  3a1f04:      	call	0x3c6940 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a1f09:      	mov	rax, qword ptr [r13]
  3a1f0d:      	mov	r8, qword ptr [r13 + 0x8]
  3a1f11:      	mov	qword ptr [rbp - 0x80], r13
  3a1f15:      	movzx	edx, byte ptr [r13 + 0x10]
  3a1f1a:      	mov	r11, qword ptr [rbp - 0x48]
  3a1f1e:      	mov	r10d, 0xffffffff
  3a1f24:      	mov	r9, qword ptr [rbp - 0x38]
  3a1f28:      	mov	r13, qword ptr [rbp - 0x50]
  3a1f2c:      	jmp	0x3a24b9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe09>
  3a1f31:      	mov	rax, r9
  3a1f34:      	sub	rax, rdi
  3a1f37:      	not	r9
  3a1f3a:      	mov	r10, rdi
  3a1f3d:      	and	r10, r9
  3a1f40:      	xor	r9, rdi
  3a1f43:      	and	r9, rax
  3a1f46:      	or	r9, r10
  3a1f49:      	shr	r9, 0x3f
  3a1f4d:      	add	rdx, r9
  3a1f50:      	sub	rcx, rdx
  3a1f53:      	mov	rdx, rcx
  3a1f56:      	mov	r10d, 0xffffffff
  3a1f5c:      	mov	edi, dword ptr [r12 + 0x1c]
  3a1f61:      	cmp	rsi, rdi
  3a1f64:      	mov	r9, qword ptr [rbp - 0x38]
  3a1f68:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a1f6e:      	lea	ecx, [rdi + 0x1]
  3a1f71:      	cmp	rsi, rcx
  3a1f74:      	mov	r12, qword ptr [rbp - 0x70]
  3a1f78:      	jbe	0x3a3566 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eb6>
  3a1f7e:      	mov	esi, dword ptr [r8 + 4*rdi]
  3a1f82:      	cmp	r9, rsi
  3a1f85:      	jbe	0x3a359d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eed>
  3a1f8b:      	mov	edi, dword ptr [r8 + 4*rcx]
  3a1f8f:      	shl	rsi, 0x4
  3a1f93:      	mov	qword ptr [r13 + rsi], rax
  3a1f98:      	mov	qword ptr [r13 + rsi + 0x8], 0x0
  3a1fa1:      	cmp	r9, rdi
  3a1fa4:      	jbe	0x3a354a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e9a>
  3a1faa:      	shl	rdi, 0x4
  3a1fae:      	mov	qword ptr [r13 + rdi], rdx
  3a1fb3:      	mov	qword ptr [r13 + rdi + 0x8], 0x0
  3a1fbc:      	inc	r14d
  3a1fbf:      	mov	dword ptr [r12 - 0x38], r14d
  3a1fc4:      	mov	r13d, r14d
  3a1fc7:      	mov	r8, qword ptr [rbp - 0x68]
  3a1fcb:      	mov	rcx, qword ptr [rbp - 0x40]
  3a1fcf:      	jmp	0x3a17b8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x108>
  3a1fd4:      	lea	esi, [rdi - 0x5b]
  3a1fd7:      	cmp	esi, 0x69
  3a1fda:      	ja	0x3a2544 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe94>
  3a1fe0:      	jmp	qword ptr [8*rsi + 0x203428]
  3a1fe7:      	mov	rsi, rax
  3a1fea:      	mov	rdx, rcx
  3a1fed:      	call	0x3d0200 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a1ff2:      	mov	r10d, 0xffffffff
  3a1ff8:      	mov	r11, qword ptr [rbp - 0x48]
  3a1ffc:      	xor	r8d, r8d
  3a1fff:      	jmp	0x3a24b5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe05>
  3a2004:      	add	rax, rcx
  3a2007:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a200c:      	mov	rcx, rax
  3a200f:      	shr	rcx
  3a2012:      	movabs	rdx, 0x5555555555555555
  3a201c:      	and	rcx, rdx
  3a201f:      	and	rax, rdx
  3a2022:      	add	rax, rcx
  3a2025:      	mov	rcx, rax
  3a2028:      	shr	rcx, 0x2
  3a202c:      	movabs	rdx, 0x3333333333333333
  3a2036:      	and	rcx, rdx
  3a2039:      	and	rax, rdx
  3a203c:      	add	rax, rcx
  3a203f:      	mov	rcx, rax
  3a2042:      	shr	rcx, 0x4
  3a2046:      	add	rcx, rax
  3a2049:      	movabs	rax, 0xf0f0f0f0f0f0f0f
  3a2053:      	and	rax, rcx
  3a2056:      	mov	rcx, rax
  3a2059:      	shr	rcx, 0x8
  3a205d:      	add	rcx, rax
  3a2060:      	mov	rdx, rcx
  3a2063:      	shr	rdx, 0x10
  3a2067:      	add	rdx, rcx
  3a206a:      	mov	rax, rdx
  3a206d:      	shr	rax, 0x20
  3a2071:      	add	eax, edx
  3a2073:      	and	eax, 0x7f
  3a2076:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a207b:      	xor	ecx, ecx
  3a207d:      	mov	rdx, rax
  3a2080:      	shr	rdx, 0x20
  3a2084:      	cmove	rdx, rax
  3a2088:      	setne	cl
  3a208b:      	shl	ecx, 0x5
  3a208e:      	mov	rax, rdx
  3a2091:      	shr	rax, 0x10
  3a2095:      	lea	rsi, [rcx + 0x10]
  3a2099:      	cmp	rdx, 0x10000
  3a20a0:      	cmovb	rax, rdx
  3a20a4:      	cmovb	rsi, rcx
  3a20a8:      	mov	ecx, eax
  3a20aa:      	shr	ecx, 0x8
  3a20ad:      	lea	rdx, [rsi + 0x8]
  3a20b1:      	cmp	rax, 0x100
  3a20b7:      	cmovb	rcx, rax
  3a20bb:      	cmovb	rdx, rsi
  3a20bf:      	movzx	eax, cl
  3a20c2:      	movzx	ecx, byte ptr [rax + 0x3285c9]
  3a20c9:      	add	rcx, rdx
  3a20cc:      	mov	eax, 0x40
  3a20d1:      	sub	rax, rcx
  3a20d4:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a20d9:      	xor	edx, edx
  3a20db:      	cmp	rax, rcx
  3a20de:      	setle	dl
  3a20e1:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a20e6:      	xor	edx, edx
  3a20e8:      	cmp	rax, rcx
  3a20eb:      	setae	dl
  3a20ee:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a20f3:      	add	ecx, eax
  3a20f5:      	jmp	0x3a229e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbee>
  3a20fa:      	test	eax, eax
  3a20fc:      	je	0x3a2442 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd92>
  3a2102:      	tzcnt	eax, eax
  3a2106:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a210b:      	shr	eax, cl
  3a210d:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2112:      	and	eax, ecx
  3a2114:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2119:      	xor	edx, edx
  3a211b:      	cmp	rax, rcx
  3a211e:      	sete	dl
  3a2121:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2126:      	sub	rax, rcx
  3a2129:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a212e:      	xor	ecx, ecx
  3a2130:      	test	rax, rax
  3a2133:      	sete	cl
  3a2136:      	jmp	0x3a229e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbee>
  3a213b:      	mov	ecx, eax
  3a213d:      	shr	ecx, 0x10
  3a2140:      	xor	edx, edx
  3a2142:      	cmp	eax, 0x10000
  3a2147:      	setb	dl
  3a214a:      	cmovb	ecx, eax
  3a214d:      	shl	edx, 0x4
  3a2150:      	mov	esi, ecx
  3a2152:      	shr	esi, 0x8
  3a2155:      	xor	eax, eax
  3a2157:      	cmp	ecx, 0x100
  3a215d:      	setb	al
  3a2160:      	lea	eax, [rdx + 8*rax]
  3a2163:      	cmovb	esi, ecx
  3a2166:      	movzx	ecx, sil
  3a216a:      	movzx	ecx, byte ptr [rcx + 0x3285c9]
  3a2171:      	sub	rax, rcx
  3a2174:      	add	rax, 0x8
  3a2178:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a217d:      	xor	edx, edx
  3a217f:      	cmp	rax, rcx
  3a2182:      	setbe	dl
  3a2185:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a218a:      	or	ecx, eax
  3a218c:      	jmp	0x3a229e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbee>
  3a2191:      	shl	eax, cl
  3a2193:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2198:      	xor	edx, edx
  3a219a:      	cmp	eax, ecx
  3a219c:      	seta	dl
  3a219f:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a21a4:      	xor	edx, edx
  3a21a6:      	cmp	eax, ecx
  3a21a8:      	setle	dl
  3a21ab:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a21b0:      	xor	edx, edx
  3a21b2:      	cmp	eax, ecx
  3a21b4:      	setb	dl
  3a21b7:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a21bc:      	mov	ecx, eax
  3a21be:      	shr	ecx, 0x10
  3a21c1:      	movzx	ecx, cl
  3a21c4:      	movzx	ecx, byte ptr [rcx + 0x3286c9]
  3a21cb:      	movzx	edx, ah
  3a21ce:      	movzx	esi, al
  3a21d1:      	shr	eax, 0x18
  3a21d4:      	add	cl, byte ptr [rax + 0x3286c9]
  3a21da:      	add	cl, byte ptr [rdx + 0x3286c9]
  3a21e0:      	add	cl, byte ptr [rsi + 0x3286c9]
  3a21e6:      	movzx	eax, cl
  3a21e9:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a21ee:      	imul	rax, rcx
  3a21f2:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a21f7:      	xor	edx, edx
  3a21f9:      	cmp	eax, ecx
  3a21fb:      	setne	dl
  3a21fe:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2203:      	xor	edx, edx
  3a2205:      	cmp	eax, ecx
  3a2207:      	sete	dl
  3a220a:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a220f:      	xor	edx, edx
  3a2211:      	cmp	eax, ecx
  3a2213:      	setge	dl
  3a2216:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2218:      	xor	edx, edx
  3a221a:      	cmp	eax, ecx
  3a221c:      	setbe	dl
  3a221f:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2221:      	xor	edx, edx
  3a2223:      	cmp	rax, rcx
  3a2226:      	setl	dl
  3a2229:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a222b:      	xor	edx, edx
  3a222d:      	cmp	rax, rcx
  3a2230:      	setb	dl
  3a2233:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2235:      	xor	ecx, eax
  3a2237:      	jmp	0x3a229e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbee>
  3a2239:      	xor	edx, edx
  3a223b:      	cmp	eax, ecx
  3a223d:      	setg	dl
  3a2240:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2242:      	mov	ecx, 0x40
  3a2247:      	tzcnt	rcx, rax
  3a224c:      	jmp	0x3a229e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbee>
  3a224e:      	xor	edx, edx
  3a2250:      	cmp	eax, ecx
  3a2252:      	setl	dl
  3a2255:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2257:      	xor	edx, edx
  3a2259:      	cmp	rax, rcx
  3a225c:      	seta	dl
  3a225f:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2261:      	rol	eax, cl
  3a2263:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2268:      	xor	edx, edx
  3a226a:      	cmp	eax, ecx
  3a226c:      	setae	dl
  3a226f:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2271:      	xor	edx, edx
  3a2273:      	cmp	rax, rcx
  3a2276:      	setg	dl
  3a2279:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a227b:      	xor	edx, edx
  3a227d:      	cmp	rax, rcx
  3a2280:      	setne	dl
  3a2283:      	jmp	0x3a2290 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbe0>
  3a2285:      	xor	edx, edx
  3a2287:      	cmp	rax, rcx
  3a228a:      	setge	dl
  3a228d:      	nop	dword ptr [rax]
  3a2290:      	xor	r8d, r8d
  3a2293:      	mov	rax, rdx
  3a2296:      	jmp	0x3a2453 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda3>
  3a229b:      	imul	ecx, eax
  3a229e:      	xor	r8d, r8d
  3a22a1:      	mov	rax, rcx
  3a22a4:      	jmp	0x3a2453 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda3>
  3a22a9:      	sub	eax, ecx
  3a22ab:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a22b0:      	ror	eax, cl
  3a22b2:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a22b7:      	sar	eax, cl
  3a22b9:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a22be:      	xor	r8d, r8d
  3a22c1:      	mov	edi, dword ptr [r12 + 0x10]
  3a22c6:      	cmp	rdi, r10
  3a22c9:      	jne	0x3a1bac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4fc>
  3a22cf:      	xor	eax, eax
  3a22d1:      	xor	r9d, r9d
  3a22d4:      	mov	r10d, 0x0
  3a22da:      	mov	r11d, 0x0
  3a22e0:      	cmp	dx, 0xc
  3a22e4:      	jne	0x3a2324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc74>
  3a22e6:      	mov	rcx, qword ptr [rbp - 0x60]
  3a22ea:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a22f1:      	mov	edi, dword ptr [r12 + 0x18]
  3a22f6:      	add	edi, 0x2
  3a22f9:      	cmp	rsi, rdi
  3a22fc:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2302:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a2309:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a230c:      	cmp	qword ptr [rbp - 0x38], rdi
  3a2310:      	jbe	0x3a332e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a2316:      	shl	rdi, 0x4
  3a231a:      	mov	r10, qword ptr [r13 + rdi]
  3a231f:      	mov	r11, qword ptr [r13 + rdi + 0x8]
  3a2324:      	cmp	qword ptr [rbp - 0x48], 0x0
  3a2329:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a232f:      	mov	rdx, qword ptr [rbp - 0x60]
  3a2333:      	mov	rsi, qword ptr [rdx + 0x80]
  3a233a:      	mov	edi, dword ptr [r12 + 0x4]
  3a233f:      	cmp	rsi, rdi
  3a2342:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2348:      	imul	rcx, rdi, 0xe0
  3a234f:      	add	rcx, qword ptr [rdx + 0x78]
  3a2353:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2359:      	mov	rdx, qword ptr [rbp - 0x48]
  3a235d:      	mov	rsi, qword ptr [rdx]
  3a2360:      	mov	qword ptr [rbp - 0xd0], rax
  3a2367:      	mov	qword ptr [rbp - 0xc8], r9
  3a236e:      	mov	qword ptr [rbp - 0xc0], r10
  3a2375:      	mov	qword ptr [rbp - 0xb8], r11
  3a237c:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a2383:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a2388:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a238f:      	movups	xmmword ptr [rsp], xmm0
  3a2393:      	lea	rdi, [rbp - 0x180]
  3a239a:      	mov	rdx, qword ptr [rbp - 0x58]
  3a239e:      	call	0x3b6a00 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a23a3:      	movaps	xmm0, xmmword ptr [rbp - 0x180]
  3a23aa:      	mov	rax, qword ptr [rbp - 0x170]
  3a23b1:      	test	rax, rax
  3a23b4:      	je	0x3a23d4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd24>
  3a23b6:      	cmp	qword ptr [rax + 0x8], 0x0
  3a23bb:      	mov	r10d, 0xffffffff
  3a23c1:      	mov	r9, qword ptr [rbp - 0x38]
  3a23c5:      	mov	r8, qword ptr [rbp - 0x68]
  3a23c9:      	mov	rcx, qword ptr [rbp - 0x40]
  3a23cd:      	je	0x3a23e6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd36>
  3a23cf:      	jmp	0x3a317f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1acf>
  3a23d4:      	mov	r10d, 0xffffffff
  3a23da:      	mov	r9, qword ptr [rbp - 0x38]
  3a23de:      	mov	r8, qword ptr [rbp - 0x68]
  3a23e2:      	mov	rcx, qword ptr [rbp - 0x40]
  3a23e6:      	mov	edi, dword ptr [r12 + 0x8]
  3a23eb:      	cmp	rdi, r10
  3a23ee:      	mov	r11, qword ptr [rbp - 0x48]
  3a23f2:      	mov	r12, qword ptr [rbp - 0x70]
  3a23f6:      	je	0x3a1c1a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x56a>
  3a23fc:      	cmp	r9, rdi
  3a23ff:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a2405:      	shl	rdi, 0x4
  3a2409:      	movups	xmmword ptr [r13 + rdi], xmm0
  3a240f:      	jmp	0x3a1c1a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x56a>
  3a2414:      	mov	qword ptr [rsp], r11
  3a2418:      	mov	rsi, rcx
  3a241b:      	mov	rdx, r10
  3a241e:      	mov	rcx, rax
  3a2421:      	call	0x3ce3c0 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a2426:      	mov	r8, rdx
  3a2429:      	mov	r11, qword ptr [rbp - 0x48]
  3a242d:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a2432:      	mov	r10d, 0xffffffff
  3a2438:      	mov	r9, qword ptr [rbp - 0x38]
  3a243c:      	mov	r13, qword ptr [rbp - 0x50]
  3a2440:      	jmp	0x3a2457 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda7>
  3a2442:      	mov	eax, 0x20
  3a2447:      	nop	word ptr [rax + rax]
  3a2450:      	xor	r8d, r8d
  3a2453:      	mov	r9, qword ptr [rbp - 0x38]
  3a2457:      	mov	rcx, qword ptr [rbp - 0x40]
  3a245b:      	mov	edi, dword ptr [r12 + 0x8]
  3a2460:      	cmp	rdi, r10
  3a2463:      	je	0x3a17ae <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe>
  3a2469:      	cmp	r9, rdi
  3a246c:      	ja	0x3a17a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0>
  3a2472:      	jmp	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a2477:      	mov	qword ptr [rbp - 0xb0], rax
  3a247e:      	mov	qword ptr [rbp - 0xa8], r8
  3a2485:      	movups	xmm0, xmmword ptr [rbp - 0xb0]
  3a248c:      	movups	xmmword ptr [rsp], xmm0
  3a2490:      	movzx	r8d, dil
  3a2494:      	mov	rdi, rsi
  3a2497:      	mov	esi, edx
  3a2499:      	mov	rdx, rcx
  3a249c:      	mov	rcx, r9
  3a249f:      	call	0x3c7e50 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a24a4:      	mov	edx, eax
  3a24a6:      	xor	r8d, r8d
  3a24a9:      	xor	eax, eax
  3a24ab:      	mov	r11, qword ptr [rbp - 0x48]
  3a24af:      	mov	r10d, 0xffffffff
  3a24b5:      	mov	r9, qword ptr [rbp - 0x38]
  3a24b9:      	mov	rcx, qword ptr [rbp - 0x40]
  3a24bd:      	test	dl, dl
  3a24bf:      	je	0x3a245b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdab>
  3a24c1:      	jmp	0x3a334d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c9d>
  3a24c6:      	or	rax, rcx
  3a24c9:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a24cb:      	mov	eax, eax
  3a24cd:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a24cf:      	shr	rax, cl
  3a24d2:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a24d7:      	and	rax, rcx
  3a24da:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a24df:      	rol	rax, cl
  3a24e2:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a24e7:      	ror	rax, cl
  3a24ea:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a24ef:      	mov	eax, eax
  3a24f1:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a24f6:      	movsx	rax, ax
  3a24fa:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a24ff:      	cdqe
  3a2501:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2506:      	cdqe
  3a2508:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a250d:      	sar	rax, cl
  3a2510:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2515:      	movsx	rax, ax
  3a2519:      	mov	eax, eax
  3a251b:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2520:      	shl	rax, cl
  3a2523:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2528:      	movsx	rax, al
  3a252c:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2531:      	movsx	rax, al
  3a2535:      	mov	eax, eax
  3a2537:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a253c:      	xor	rax, rcx
  3a253f:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2544:      	lea	esi, [rdi - 0x1]
  3a2547:      	cmp	esi, 0x108
  3a254d:      	ja	0x3a26a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xff0>
  3a2553:      	jmp	qword ptr [8*rsi + 0x203778]
  3a255a:      	mov	rdx, qword ptr [rbp - 0x60]
  3a255e:      	mov	rsi, qword ptr [rdx + 0x80]
  3a2565:      	mov	ecx, dword ptr [r12 + 0x4]
  3a256a:      	cmp	rsi, rcx
  3a256d:      	jbe	0x3a3566 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eb6>
  3a2573:      	imul	r8, rcx, 0xe0
  3a257a:      	add	r8, qword ptr [rdx + 0x78]
  3a257e:      	mov	rdx, qword ptr [rbp - 0x58]
  3a2582:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2588:      	mov	rsi, qword ptr [rdx + 0x78]
  3a258c:      	mov	ecx, dword ptr [r8 + 0x18]
  3a2590:      	cmp	rsi, rcx
  3a2593:      	jbe	0x3a3566 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eb6>
  3a2599:      	shl	rcx, 0x3
  3a259d:      	add	rcx, qword ptr [rdx + 0x70]
  3a25a1:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a25a7:      	mov	rcx, qword ptr [rcx]
  3a25aa:      	test	rcx, rcx
  3a25ad:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a25b3:      	mov	rdx, rdi
  3a25b6:      	mov	rsi, qword ptr [rcx + 0x10]
  3a25ba:      	mov	ecx, dword ptr [r8 + 0x14]
  3a25be:      	mov	r8d, dword ptr [r8 + 0x1c]
  3a25c2:      	shl	r8, 0x20
  3a25c6:      	or	r8, rcx
  3a25c9:      	xorps	xmm0, xmm0
  3a25cc:      	movaps	xmmword ptr [rbp - 0x250], xmm0
  3a25d3:      	movups	xmmword ptr [rsp], xmm0
  3a25d7:      	lea	rdi, [rbp - 0x108]
  3a25de:      	mov	rcx, rax
  3a25e1:      	xor	r9d, r9d
  3a25e4:      	call	0x3c6770 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Load>
  3a25e9:      	mov	rax, qword ptr [rbp - 0x108]
  3a25f0:      	mov	r8, qword ptr [rbp - 0x100]
  3a25f7:      	movzx	edx, byte ptr [rbp - 0xf8]
  3a25fe:      	jmp	0x3a24ab <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdfb>
  3a2603:      	movzx	ecx, dl
  3a2606:      	xor	edi, edi
  3a2608:      	add	ecx, -0x2
  3a260b:      	cmp	ecx, 0x7
  3a260e:      	ja	0x3a2898 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11e8>
  3a2614:      	mov	r8, qword ptr [rbp - 0x68]
  3a2618:      	jmp	qword ptr [8*rcx + 0x203fc0]
  3a261f:      	xor	edi, edi
  3a2621:      	test	eax, eax
  3a2623:      	sete	dil
  3a2627:      	jmp	0x3a289c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11ec>
  3a262c:      	mov	edi, dword ptr [r12 + 0x18]
  3a2631:      	mov	r9, qword ptr [rbp - 0x38]
  3a2635:      	cmp	r9, rdi
  3a2638:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a263e:      	shl	rdi, 0x4
  3a2642:      	mov	qword ptr [r13 + rdi], rax
  3a2647:      	mov	qword ptr [r13 + rdi + 0x8], r8
  3a264c:      	jmp	0x3a2457 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda7>
  3a2651:      	mov	edi, dword ptr [r12 + 0x14]
  3a2656:      	cmp	rdi, r10
  3a2659:      	je	0x3a288d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11dd>
  3a265f:      	cmp	qword ptr [rbp - 0x38], rdi
  3a2663:      	jbe	0x3a332e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a2669:      	shl	rdi, 0x4
  3a266d:      	mov	rdx, qword ptr [r13 + rdi]
  3a2672:      	test	edx, edx
  3a2674:      	cmove	rax, rcx
  3a2678:      	cmove	r8, r9
  3a267c:      	jmp	0x3a2453 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda3>
  3a2681:      	mov	edi, dword ptr [r12 + 0x1c]
  3a2686:      	dec	edi
  3a2688:      	cmp	edi, eax
  3a268a:      	cmovae	edi, eax
  3a268d:      	jmp	0x3a289c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11ec>
  3a2692:      	xor	edi, edi
  3a2694:      	test	rax, rax
  3a2697:      	setne	dil
  3a269b:      	jmp	0x3a289c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11ec>
  3a26a0:      	lea	esi, [rdi - 0x10]
  3a26a3:      	cmp	esi, 0x34
  3a26a6:      	ja	0x3a29ac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12fc>
  3a26ac:      	jmp	qword ptr [8*rsi + 0x204000]
  3a26b3:      	mov	rax, qword ptr [r12 + 0x18]
  3a26b8:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a26bd:      	mov	r10, qword ptr [rbp - 0x60]
  3a26c1:      	mov	rsi, qword ptr [r10 + 0x80]
  3a26c8:      	mov	edx, dword ptr [r12 + 0x4]
  3a26cd:      	cmp	rsi, rdx
  3a26d0:      	jbe	0x3a3650 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa0>
  3a26d6:      	imul	r8, rdx, 0xe0
  3a26dd:      	add	r8, qword ptr [r10 + 0x78]
  3a26e1:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a26e7:      	mov	r10, qword ptr [rbp - 0x58]
  3a26eb:      	mov	rsi, qword ptr [r10 + 0x78]
  3a26ef:      	mov	edx, dword ptr [r8 + 0x18]
  3a26f3:      	cmp	rsi, rdx
  3a26f6:      	jbe	0x3a3650 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa0>
  3a26fc:      	shl	rdx, 0x3
  3a2700:      	add	rdx, qword ptr [r10 + 0x70]
  3a2704:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a270a:      	mov	rdx, qword ptr [rdx]
  3a270d:      	test	rdx, rdx
  3a2710:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2716:      	mov	rsi, rdi
  3a2719:      	mov	rdi, qword ptr [rdx + 0x10]
  3a271d:      	mov	edx, dword ptr [r8 + 0x14]
  3a2721:      	mov	r8d, dword ptr [r8 + 0x1c]
  3a2725:      	shl	r8, 0x20
  3a2729:      	or	r8, rdx
  3a272c:      	mov	qword ptr [rbp - 0xe0], rcx
  3a2733:      	mov	qword ptr [rbp - 0xd8], r9
  3a273a:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a2741:      	movups	xmmword ptr [rsp], xmm0
  3a2745:      	mov	rdx, rax
  3a2748:      	mov	rcx, r8
  3a274b:      	xor	r8d, r8d
  3a274e:      	call	0x3c7b10 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Store>
  3a2753:      	jmp	0x3a24a4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdf4>
  3a2758:      	mov	rax, qword ptr [rbp - 0x58]
  3a275c:      	mov	rsi, qword ptr [rax + 0x90]
  3a2763:      	mov	edi, dword ptr [r12 + 0x18]
  3a2768:      	cmp	rsi, rdi
  3a276b:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2771:      	mov	rax, qword ptr [rax + 0x88]
  3a2778:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a277c:      	mov	rax, qword ptr [rcx + 0x20]
  3a2780:      	mov	r8, qword ptr [rcx + 0x28]
  3a2784:      	jmp	0x3a2453 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda3>
  3a2789:      	mov	rcx, qword ptr [rbp - 0x58]
  3a278d:      	mov	rsi, qword ptr [rcx + 0x90]
  3a2794:      	mov	edi, dword ptr [r12 + 0x18]
  3a2799:      	cmp	rsi, rdi
  3a279c:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a27a2:      	mov	rcx, qword ptr [rcx + 0x88]
  3a27a9:      	mov	rcx, qword ptr [rcx + 8*rdi]
  3a27ad:      	mov	qword ptr [rcx + 0x20], rax
  3a27b1:      	mov	qword ptr [rcx + 0x28], r8
  3a27b5:      	xor	eax, eax
  3a27b7:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a27bc:      	mov	edi, dword ptr [r12 + 0x18]
  3a27c1:      	mov	r9, qword ptr [rbp - 0x38]
  3a27c5:      	cmp	r9, rdi
  3a27c8:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a27ce:      	shl	rdi, 0x4
  3a27d2:      	mov	rax, qword ptr [r13 + rdi]
  3a27d7:      	mov	r8, qword ptr [r13 + rdi + 0x8]
  3a27dc:      	jmp	0x3a2457 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda7>
  3a27e1:      	mov	rcx, qword ptr [rbp - 0x58]
  3a27e5:      	mov	rsi, qword ptr [rcx + 0x60]
  3a27e9:      	mov	edi, dword ptr [r12 + 0x18]
  3a27ee:      	cmp	rsi, rdi
  3a27f1:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a27f7:      	mov	rcx, qword ptr [rcx + 0x58]
  3a27fb:      	lea	rdx, [rcx + 8*rdi]
  3a27ff:      	test	rdx, rdx
  3a2802:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2808:      	mov	rcx, qword ptr [rcx + 8*rdi]
  3a280c:      	test	rcx, rcx
  3a280f:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2815:      	cmp	byte ptr [rcx + 0x29], 0x0
  3a2819:      	mov	edx, eax
  3a281b:      	cmovne	rdx, rax
  3a281f:      	cmp	rdx, qword ptr [rcx + 0x40]
  3a2823:      	jae	0x3a3337 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c87>
  3a2829:      	mov	rax, qword ptr [rcx + 0x38]
  3a282d:      	mov	rax, qword ptr [rax + 8*rdx]
  3a2831:      	jmp	0x3a2885 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11d5>
  3a2833:      	mov	rdx, qword ptr [rbp - 0x58]
  3a2837:      	mov	rsi, qword ptr [rdx + 0x60]
  3a283b:      	mov	edi, dword ptr [r12 + 0x18]
  3a2840:      	cmp	rsi, rdi
  3a2843:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2849:      	mov	rdx, qword ptr [rdx + 0x58]
  3a284d:      	lea	rsi, [rdx + 8*rdi]
  3a2851:      	test	rsi, rsi
  3a2854:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a285a:      	mov	rdx, qword ptr [rdx + 8*rdi]
  3a285e:      	test	rdx, rdx
  3a2861:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2867:      	cmp	byte ptr [rdx + 0x29], 0x0
  3a286b:      	mov	esi, eax
  3a286d:      	cmovne	rsi, rax
  3a2871:      	cmp	rsi, qword ptr [rdx + 0x40]
  3a2875:      	jae	0x3a3337 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c87>
  3a287b:      	mov	rax, qword ptr [rdx + 0x38]
  3a287f:      	mov	qword ptr [rax + 8*rsi], rcx
  3a2883:      	xor	eax, eax
  3a2885:      	xor	r8d, r8d
  3a2888:      	jmp	0x3a2429 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd79>
  3a288d:      	mov	rax, rcx
  3a2890:      	mov	r8, r9
  3a2893:      	jmp	0x3a2453 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda3>
  3a2898:      	mov	r8, qword ptr [rbp - 0x68]
  3a289c:      	mov	rax, qword ptr [rbp - 0x60]
  3a28a0:      	mov	rsi, qword ptr [rax + 0xe8]
  3a28a7:      	add	edi, dword ptr [r12 + 0x18]
  3a28ac:      	cmp	rsi, rdi
  3a28af:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a28b5:      	mov	rax, qword ptr [rax + 0xe0]
  3a28bc:      	lea	rcx, [rdi + 2*rdi]
  3a28c0:      	mov	r13d, dword ptr [rax + 4*rcx]
  3a28c4:      	mov	edi, dword ptr [rax + 4*rcx + 0x4]
  3a28c8:      	mov	r12d, dword ptr [rax + 4*rcx + 0x8]
  3a28cd:      	cmp	r13d, r14d
  3a28d0:      	mov	r9, qword ptr [rbp - 0x38]
  3a28d4:      	ja	0x3a291c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x126c>
  3a28d6:      	mov	r14, rdi
  3a28d9:      	mov	rdi, r11
  3a28dc:      	call	0x3a7130 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).poll>
  3a28e1:      	test	rax, rax
  3a28e4:      	je	0x3a2907 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1257>
  3a28e6:      	cmp	qword ptr [rax + 0x8], 0x0
  3a28eb:      	mov	r11, qword ptr [rbp - 0x48]
  3a28ef:      	mov	r10d, 0xffffffff
  3a28f5:      	mov	r9, qword ptr [rbp - 0x38]
  3a28f9:      	mov	r8, qword ptr [rbp - 0x68]
  3a28fd:      	mov	rdi, r14
  3a2900:      	je	0x3a291c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x126c>
  3a2902:      	jmp	0x3a333b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c8b>
  3a2907:      	mov	r11, qword ptr [rbp - 0x48]
  3a290b:      	mov	r10d, 0xffffffff
  3a2911:      	mov	r9, qword ptr [rbp - 0x38]
  3a2915:      	mov	r8, qword ptr [rbp - 0x68]
  3a2919:      	mov	rdi, r14
  3a291c:      	mov	rax, qword ptr [rbp - 0x60]
  3a2920:      	mov	rdx, qword ptr [rax + 0x108]
  3a2927:      	add	r12d, edi
  3a292a:      	cmp	rdx, r12
  3a292d:      	jb	0x3a3658 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa8>
  3a2933:      	cmp	edi, r12d
  3a2936:      	ja	0x3a3667 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fb7>
  3a293c:      	sub	rdx, rdi
  3a293f:      	test	rdx, rdx
  3a2942:      	mov	eax, 0x0
  3a2947:      	cmovg	rax, rdi
  3a294b:      	sub	r12, rdi
  3a294e:      	jle	0x3a29a3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12f3>
  3a2950:      	mov	rcx, qword ptr [rbp - 0x60]
  3a2954:      	mov	rcx, qword ptr [rcx + 0xf8]
  3a295b:      	lea	rcx, [rcx + 8*rax]
  3a295f:      	add	rcx, 0x4
  3a2963:      	xor	edx, edx
  3a2965:      	cmp	r12, rdx
  3a2968:      	je	0x3a362b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f7b>
  3a296e:      	mov	edi, dword ptr [rcx + 8*rdx]
  3a2971:      	cmp	r9, rdi
  3a2974:      	jbe	0x3a331c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a297a:      	mov	eax, dword ptr [rcx + 8*rdx - 0x4]
  3a297e:      	cmp	r9, rax
  3a2981:      	jbe	0x3a3636 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f86>
  3a2987:      	shl	rdi, 0x4
  3a298b:      	shl	rax, 0x4
  3a298f:      	mov	rsi, qword ptr [rbp - 0x50]
  3a2993:      	movups	xmm0, xmmword ptr [rsi + rdi]
  3a2997:      	movups	xmmword ptr [rsi + rax], xmm0
  3a299b:      	inc	rdx
  3a299e:      	cmp	r12, rdx
  3a29a1:      	jne	0x3a2965 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x12b5>
  3a29a3:      	mov	rcx, qword ptr [rbp - 0x40]
  3a29a7:      	jmp	0x3a17b8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x108>
  3a29ac:      	lea	esi, [rdi - 0xd0]
  3a29b2:      	cmp	esi, 0x4
  3a29b5:      	ja	0x3a29ce <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x131e>
  3a29b7:      	jmp	qword ptr [8*rsi + 0x2041a8]
  3a29be:      	xor	ecx, ecx
  3a29c0:      	test	rax, rax
  3a29c3:      	sete	cl
  3a29c6:      	mov	rax, rcx
  3a29c9:      	jmp	0x3a2885 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11d5>
  3a29ce:      	mov	rsi, rdi
  3a29d1:      	mov	edi, dword ptr [r12 + 0x14]
  3a29d6:      	cmp	rdi, r10
  3a29d9:      	je	0x3a2b43 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1493>
  3a29df:      	cmp	qword ptr [rbp - 0x38], rdi
  3a29e3:      	jbe	0x3a332e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c7e>
  3a29e9:      	shl	rdi, 0x4
  3a29ed:      	mov	r8, qword ptr [rbp - 0x50]
  3a29f1:      	movups	xmm0, xmmword ptr [r8 + rdi]
  3a29f6:      	jmp	0x3a2b46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1496>
  3a29fb:      	mov	eax, dword ptr [r12 + 0x18]
  3a2a00:      	jmp	0x3a2450 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xda0>
  3a2a05:      	mov	rax, qword ptr [rbp - 0x58]
  3a2a09:      	mov	rsi, qword ptr [rax + 0x78]
  3a2a0d:      	mov	edi, dword ptr [r12 + 0x18]
  3a2a12:      	cmp	rsi, rdi
  3a2a15:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2a1b:      	shl	rdi, 0x3
  3a2a1f:      	mov	rax, qword ptr [rbp - 0x58]
  3a2a23:      	add	rdi, qword ptr [rax + 0x70]
  3a2a27:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2a2d:      	mov	rax, qword ptr [rdi]
  3a2a30:      	test	rax, rax
  3a2a33:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2a39:      	mov	rdi, qword ptr [rax + 0x10]
  3a2a3d:      	call	0x3c4450 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Pages>
  3a2a42:      	jmp	0x3a2885 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11d5>
  3a2a47:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2a4b:      	mov	rsi, qword ptr [rcx + 0x78]
  3a2a4f:      	mov	edi, dword ptr [r12 + 0x18]
  3a2a54:      	cmp	rsi, rdi
  3a2a57:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2a5d:      	shl	rdi, 0x3
  3a2a61:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2a65:      	add	rdi, qword ptr [rcx + 0x70]
  3a2a69:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2a6f:      	cmp	qword ptr [rbp - 0x48], 0x0
  3a2a74:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2a7a:      	mov	r15, qword ptr [rdi]
  3a2a7d:      	test	r15, r15
  3a2a80:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2a86:      	mov	rcx, qword ptr [rbp - 0x48]
  3a2a8a:      	mov	rdi, qword ptr [rcx]
  3a2a8d:      	mov	rcx, qword ptr [r15 + 0x10]
  3a2a91:      	movzx	ecx, byte ptr [rcx + 0x31]
  3a2a95:      	mov	byte ptr [rbp - 0x71], cl
  3a2a98:      	mov	edx, eax
  3a2a9a:      	test	cl, 0x1
  3a2a9d:      	cmovne	rdx, rax
  3a2aa1:      	mov	rsi, r15
  3a2aa4:      	call	0x3c0c70 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a2aa9:      	test	dl, 0x1
  3a2aac:      	jne	0x3a2885 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11d5>
  3a2ab2:      	mov	rax, qword ptr [r15 + 0x10]
  3a2ab6:      	movzx	eax, byte ptr [rax + 0x31]
  3a2aba:      	mov	byte ptr [rbp - 0x72], al
  3a2abd:      	and	al, 0x1
  3a2abf:      	neg	al
  3a2ac1:      	mov	eax, 0x0
  3a2ac6:      	sbb	rax, rax
  3a2ac9:      	mov	r10d, 0xffffffff
  3a2acf:      	or	rax, r10
  3a2ad2:      	xor	r8d, r8d
  3a2ad5:      	mov	r11, qword ptr [rbp - 0x48]
  3a2ad9:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a2ade:      	jmp	0x3a2438 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd88>
  3a2ae3:      	test	rax, rax
  3a2ae6:      	mov	r11, qword ptr [rbp - 0x48]
  3a2aea:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a2aef:      	mov	r10d, 0xffffffff
  3a2af5:      	mov	r9, qword ptr [rbp - 0x38]
  3a2af9:      	mov	r13, qword ptr [rbp - 0x50]
  3a2afd:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2b01:      	jne	0x3a245b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xdab>
  3a2b07:      	jmp	0x3a3600 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f50>
  3a2b0c:      	mov	rax, qword ptr [rbp - 0x58]
  3a2b10:      	mov	rsi, qword ptr [rax + 0x48]
  3a2b14:      	mov	edi, dword ptr [r12 + 0x18]
  3a2b19:      	cmp	rsi, rdi
  3a2b1c:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2b22:      	mov	rax, qword ptr [rbp - 0x58]
  3a2b26:      	mov	rax, qword ptr [rax + 0x40]
  3a2b2a:      	mov	rax, qword ptr [rax + 8*rdi]
  3a2b2e:      	jmp	0x3a2885 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11d5>
  3a2b33:      	xor	edx, edx
  3a2b35:      	cmp	rax, rcx
  3a2b38:      	sete	dl
  3a2b3b:      	mov	rax, rdx
  3a2b3e:      	jmp	0x3a2885 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11d5>
  3a2b43:      	xorps	xmm0, xmm0
  3a2b46:      	and	edx, 0xff0000
  3a2b4c:      	cmp	edx, 0xfe0000
  3a2b52:      	jne	0x3a2b8b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14db>
  3a2b54:      	movzx	r9d, si
  3a2b58:      	cmp	r9d, 0x3
  3a2b5c:      	jne	0x3a2cff <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x164f>
  3a2b62:      	cmp	qword ptr [rbp - 0x48], 0x0
  3a2b67:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2b6d:      	mov	rax, qword ptr [rbp - 0x48]
  3a2b71:      	mov	rax, qword ptr [rax]
  3a2b74:      	test	rax, rax
  3a2b77:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2b7d:      	mov	rdi, qword ptr [rax + 0x8]
  3a2b81:      	call	0x3c3aa0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a2b86:      	jmp	0x3a2883 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11d3>
  3a2b8b:      	add	esi, 0xff03fff8
  3a2b91:      	cmp	esi, 0xa
  3a2b94:      	jae	0x3a3676 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fc6>
  3a2b9a:      	mov	rdx, qword ptr [rbp - 0x60]
  3a2b9e:      	mov	rsi, qword ptr [rdx + 0x80]
  3a2ba5:      	mov	edi, dword ptr [r12 + 0x4]
  3a2baa:      	cmp	rsi, rdi
  3a2bad:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2bb3:      	imul	rdx, rdi, 0xe0
  3a2bba:      	mov	rsi, qword ptr [rbp - 0x60]
  3a2bbe:      	add	rdx, qword ptr [rsi + 0x78]
  3a2bc2:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2bc8:      	movups	xmm1, xmmword ptr [rdx + 0x98]
  3a2bcf:      	movups	xmm2, xmmword ptr [rdx + 0x88]
  3a2bd6:      	movups	xmm3, xmmword ptr [rdx + 0x78]
  3a2bda:      	movups	xmm4, xmmword ptr [rdx + 0x68]
  3a2bde:      	movups	xmm5, xmmword ptr [rdx + 0x58]
  3a2be2:      	movups	xmm6, xmmword ptr [rdx + 0x48]
  3a2be6:      	movzx	esi, byte ptr [rdx + 0x40]
  3a2bea:      	movups	xmm7, xmmword ptr [rdx + 0x30]
  3a2bee:      	movups	xmm8, xmmword ptr [rdx + 0x20]
  3a2bf3:      	movups	xmm9, xmmword ptr [rdx + 0x10]
  3a2bf8:      	mov	rdi, qword ptr [rdx + 0x8]
  3a2bfc:      	mov	edx, dword ptr [rdx]
  3a2bfe:      	mov	dword ptr [rbp - 0x238], edx
  3a2c04:      	mov	qword ptr [rbp - 0x230], rdi
  3a2c0b:      	movups	xmmword ptr [rbp - 0x228], xmm9
  3a2c13:      	movups	xmmword ptr [rbp - 0x218], xmm8
  3a2c1b:      	movups	xmmword ptr [rbp - 0x208], xmm7
  3a2c22:      	mov	byte ptr [rbp - 0x1f8], sil
  3a2c29:      	movups	xmmword ptr [rbp - 0x1f0], xmm6
  3a2c30:      	movups	xmmword ptr [rbp - 0x1e0], xmm5
  3a2c37:      	movups	xmmword ptr [rbp - 0x1d0], xmm4
  3a2c3e:      	movups	xmmword ptr [rbp - 0x1c0], xmm3
  3a2c45:      	movups	xmmword ptr [rbp - 0x1b0], xmm2
  3a2c4c:      	movups	xmmword ptr [rbp - 0x1a0], xmm1
  3a2c53:      	mov	qword ptr [rbp - 0xf0], rcx
  3a2c5a:      	mov	qword ptr [rbp - 0xe8], r9
  3a2c61:      	movaps	xmmword ptr [rbp - 0x190], xmm0
  3a2c68:      	mov	ecx, 0x15
  3a2c6d:      	mov	rdi, rsp
  3a2c70:      	lea	rsi, [rbp - 0x238]
  3a2c77:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a2c7a:      	movaps	xmm0, xmmword ptr [rbp - 0x190]
  3a2c81:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a2c89:      	movups	xmm0, xmmword ptr [rbp - 0xf0]
  3a2c90:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a2c98:      	lea	rdi, [rbp - 0x130]
  3a2c9f:      	mov	rsi, qword ptr [rbp - 0x48]
  3a2ca3:      	mov	rcx, rax
  3a2ca6:      	mov	rdx, qword ptr [rbp - 0x58]
  3a2caa:      	call	0x3a9020 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a2caf:      	mov	rax, qword ptr [rbp - 0x130]
  3a2cb6:      	mov	r8, qword ptr [rbp - 0x128]
  3a2cbd:      	movzx	edx, byte ptr [rbp - 0x120]
  3a2cc4:      	mov	rsi, qword ptr [rbp - 0x118]
  3a2ccb:      	test	rsi, rsi
  3a2cce:      	je	0x3a2df4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1744>
  3a2cd4:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a2cd9:      	mov	r11, qword ptr [rbp - 0x48]
  3a2cdd:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a2ce2:      	mov	r10d, 0xffffffff
  3a2ce8:      	mov	r9, qword ptr [rbp - 0x38]
  3a2cec:      	mov	r13, qword ptr [rbp - 0x50]
  3a2cf0:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2cf4:      	je	0x3a24bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe0d>
  3a2cfa:      	jmp	0x3a360e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f5e>
  3a2cff:      	mov	rdx, qword ptr [rbp - 0x60]
  3a2d03:      	mov	rsi, qword ptr [rdx + 0x80]
  3a2d0a:      	mov	edi, dword ptr [r12 + 0x4]
  3a2d0f:      	cmp	rsi, rdi
  3a2d12:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2d18:      	imul	r11, rdi, 0xe0
  3a2d1f:      	mov	rdx, qword ptr [rbp - 0x60]
  3a2d23:      	add	r11, qword ptr [rdx + 0x78]
  3a2d27:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2d2d:      	mov	rdx, qword ptr [rbp - 0x58]
  3a2d31:      	mov	rsi, qword ptr [rdx + 0x78]
  3a2d35:      	mov	edi, dword ptr [r11 + 0x18]
  3a2d39:      	cmp	rsi, rdi
  3a2d3c:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a2d42:      	shl	rdi, 0x3
  3a2d46:      	mov	rdx, qword ptr [rbp - 0x58]
  3a2d4a:      	add	rdi, qword ptr [rdx + 0x70]
  3a2d4e:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2d54:      	mov	rsi, qword ptr [rdi]
  3a2d57:      	test	rsi, rsi
  3a2d5a:      	je	0x3a3324 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c74>
  3a2d60:      	mov	r10, qword ptr [rbp - 0x48]
  3a2d64:      	mov	rdi, qword ptr [r10]
  3a2d67:      	mov	rdx, qword ptr [r10 + 0x8]
  3a2d6b:      	mov	r8, qword ptr [rdi + 0x8]
  3a2d6f:      	mov	r10, qword ptr [r10 + 0x10]
  3a2d73:      	mov	rsi, qword ptr [rsi + 0x10]
  3a2d77:      	mov	edi, dword ptr [r11 + 0x14]
  3a2d7b:      	mov	r11d, dword ptr [r11 + 0x1c]
  3a2d7f:      	shl	r11, 0x20
  3a2d83:      	or	r11, rdi
  3a2d86:      	movlps	qword ptr [rsp + 0x18], xmm0
  3a2d8b:      	mov	qword ptr [rsp + 0x10], rcx
  3a2d90:      	mov	qword ptr [rsp + 0x8], r11
  3a2d95:      	mov	qword ptr [rsp], rax
  3a2d99:      	lea	rdi, [rbp - 0x158]
  3a2da0:      	mov	rcx, r10
  3a2da3:      	call	0x3c4d30 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a2da8:      	mov	rax, qword ptr [rbp - 0x158]
  3a2daf:      	mov	r8, qword ptr [rbp - 0x150]
  3a2db6:      	movzx	edx, byte ptr [rbp - 0x148]
  3a2dbd:      	mov	rsi, qword ptr [rbp - 0x140]
  3a2dc4:      	test	rsi, rsi
  3a2dc7:      	je	0x3a2df6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1746>
  3a2dc9:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a2dce:      	mov	r11, qword ptr [rbp - 0x48]
  3a2dd2:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a2dd7:      	mov	r10d, 0xffffffff
  3a2ddd:      	mov	r9, qword ptr [rbp - 0x38]
  3a2de1:      	mov	r13, qword ptr [rbp - 0x50]
  3a2de5:      	mov	rcx, qword ptr [rbp - 0x40]
  3a2de9:      	je	0x3a24bd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xe0d>
  3a2def:      	jmp	0x3a3617 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f67>
  3a2df4:      	jmp	0x3a2df6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1746>
  3a2df6:      	mov	r11, qword ptr [rbp - 0x48]
  3a2dfa:      	movzx	r15d, byte ptr [rbp - 0x29]
  3a2dff:      	jmp	0x3a1f1e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x86e>
  3a2e04:      	mov	rax, qword ptr [rbp - 0x70]
  3a2e08:      	mov	dword ptr [rax - 0x38], r14d
  3a2e0c:      	mov	rdi, r11
  3a2e0f:      	mov	r14, r11
  3a2e12:      	call	0x3a5e70 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a2e17:      	test	al, al
  3a2e19:      	jne	0x3a3530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e80>
  3a2e1f:      	mov	rcx, qword ptr [r14 + 0x48]
  3a2e23:      	test	rcx, rcx
  3a2e26:      	mov	r11, r14
  3a2e29:      	mov	r10d, 0xffffffff
  3a2e2f:      	jg	0x3a16f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x45>
  3a2e35:      	jmp	0x3a3685 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd5>
  3a2e3a:      	mov	rax, qword ptr [rbp - 0x70]
  3a2e3e:      	mov	dword ptr [rax - 0x38], r14d
  3a2e42:      	mov	rdi, r11
  3a2e45:      	mov	r14, r11
  3a2e48:      	call	0x3a5870 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a2e4d:      	jmp	0x3a2e17 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1767>
  3a2e4f:      	mov	rax, qword ptr [rbp - 0x70]
  3a2e53:      	mov	dword ptr [rax - 0x38], r14d
  3a2e57:      	mov	rdi, r11
  3a2e5a:      	mov	r14, r11
  3a2e5d:      	call	0x3a4500 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a2e62:      	test	al, al
  3a2e64:      	je	0x3a2e1f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x176f>
  3a2e66:      	jmp	0x3a3607 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f57>
  3a2e6b:      	mov	rax, qword ptr [rbp - 0x70]
  3a2e6f:      	mov	dword ptr [rax - 0x38], r14d
  3a2e73:      	mov	rdi, r11
  3a2e76:      	mov	r14, r11
  3a2e79:      	call	0x3a4960 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a2e7e:      	jmp	0x3a2e17 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1767>
  3a2e80:      	mov	edi, 0x28
  3a2e85:      	call	0x5fa310 <GC_malloc@plt>
  3a2e8a:      	test	rax, rax
  3a2e8d:      	je	0x3a3694 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fe4>
  3a2e93:      	mov	r14, rax
  3a2e96:      	add	r13, r15
  3a2e99:      	mov	edi, 0x28
  3a2e9e:      	call	0x5e00e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a2ea3:      	mov	rax, qword ptr [r13 + 0x8]
  3a2ea7:      	movaps	xmm0, xmmword ptr [rip + 0x30f962] # 0x6b2810 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a2eae:      	movups	xmmword ptr [r14], xmm0
  3a2eb2:      	mov	qword ptr [r14 + 0x10], rax
  3a2eb6:      	mov	qword ptr [r14 + 0x18], 0x29a200
  3a2ebe:      	mov	qword ptr [r14 + 0x20], 0xe
  3a2ec6:      	mov	r15, qword ptr [rbp - 0x48]
  3a2eca:      	mov	qword ptr [r15 + 0x140], 0x2bac00
  3a2ed5:      	mov	qword ptr [r15 + 0x148], r14
  3a2edc:      	mov	byte ptr [r15 + 0x138], 0x3
  3a2ee4:      	lea	rdi, [r15 + 0x168]
  3a2eeb:      	mov	edx, 0x158
  3a2ef0:      	xor	esi, esi
  3a2ef2:      	call	0x5fa0e0 <memset@plt>
  3a2ef7:      	mov	rdi, qword ptr [r15 + 0x40]
  3a2efb:      	mov	rdx, qword ptr [r15 + 0x48]
  3a2eff:      	shl	rdx, 0x6
  3a2f03:      	xor	esi, esi
  3a2f05:      	call	0x5fa0e0 <memset@plt>
  3a2f0a:      	mov	qword ptr [r15 + 0x48], 0x0
  3a2f12:      	mov	rdi, qword ptr [r15 + 0x58]
  3a2f16:      	mov	rax, qword ptr [r15 + 0x60]
  3a2f1a:      	shl	rax, 0x4
  3a2f1e:      	lea	rdx, [rax + 2*rax]
  3a2f22:      	xor	esi, esi
  3a2f24:      	call	0x5fa0e0 <memset@plt>
  3a2f29:      	mov	qword ptr [r15 + 0x60], 0x0
  3a2f31:      	mov	qword ptr [r15 + 0x70], 0x0
  3a2f39:      	mov	qword ptr [r15 + 0x80], 0x0
  3a2f44:      	mov	qword ptr [r15 + 0x118], 0x0
  3a2f4f:      	xorps	xmm0, xmm0
  3a2f52:      	movups	xmmword ptr [r15 + 0xd8], xmm0
  3a2f5a:      	movups	xmmword ptr [r15 + 0xe8], xmm0
  3a2f62:      	mov	qword ptr [r15 + 0xf8], 0x0
  3a2f6d:      	mov	rdi, qword ptr [r15 + 0x100]
  3a2f74:      	mov	rdx, qword ptr [r15 + 0x108]
  3a2f7b:      	shl	rdx, 0x4
  3a2f7f:      	xor	esi, esi
  3a2f81:      	call	0x5fa0e0 <memset@plt>
  3a2f86:      	cmp	qword ptr [r15 + 0x110], 0x41
  3a2f8e:      	jl	0x3a2fa6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18f6>
  3a2f90:      	mov	qword ptr [r15 + 0x100], 0x0
  3a2f9b:      	mov	qword ptr [r15 + 0x110], 0x0
  3a2fa6:      	mov	qword ptr [r15 + 0x108], 0x0
  3a2fb1:      	xorps	xmm0, xmm0
  3a2fb4:      	movups	xmmword ptr [r15 + 0x120], xmm0
  3a2fbc:      	mov	qword ptr [r15 + 0x130], 0x0
  3a2fc7:      	mov	rdi, qword ptr [r15 + 0xa8]
  3a2fce:      	mov	rdx, qword ptr [r15 + 0xb0]
  3a2fd5:      	shl	rdx, 0x4
  3a2fd9:      	xor	esi, esi
  3a2fdb:      	call	0x5fa0e0 <memset@plt>
  3a2fe0:      	mov	rax, qword ptr [rbp - 0x48]
  3a2fe4:      	mov	qword ptr [rax + 0xb0], 0x0
  3a2fef:      	mov	rax, qword ptr [rbp - 0x48]
  3a2ff3:      	mov	rdi, qword ptr [rax + 0xc0]
  3a2ffa:      	mov	rax, qword ptr [rbp - 0x48]
  3a2ffe:      	mov	rdx, qword ptr [rax + 0xc8]
  3a3005:      	jmp	0x3a34f4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e44>
  3a300a:      	mov	qword ptr [r11 + 0x140], rax
  3a3011:      	mov	qword ptr [r11 + 0x148], rdx
  3a3018:      	mov	byte ptr [r11 + 0x138], 0x3
  3a3020:      	lea	rdi, [r11 + 0x168]
  3a3027:      	mov	edx, 0x158
  3a302c:      	xor	esi, esi
  3a302e:      	call	0x5fa0e0 <memset@plt>
  3a3033:      	mov	rax, qword ptr [rbp - 0x48]
  3a3037:      	mov	rdi, qword ptr [rax + 0x40]
  3a303b:      	mov	rax, qword ptr [rbp - 0x48]
  3a303f:      	mov	rdx, qword ptr [rax + 0x48]
  3a3043:      	shl	rdx, 0x6
  3a3047:      	xor	esi, esi
  3a3049:      	call	0x5fa0e0 <memset@plt>
  3a304e:      	mov	rax, qword ptr [rbp - 0x48]
  3a3052:      	mov	qword ptr [rax + 0x48], 0x0
  3a305a:      	mov	rax, qword ptr [rbp - 0x48]
  3a305e:      	mov	rdi, qword ptr [rax + 0x58]
  3a3062:      	mov	rax, qword ptr [rbp - 0x48]
  3a3066:      	mov	rax, qword ptr [rax + 0x60]
  3a306a:      	shl	rax, 0x4
  3a306e:      	lea	rdx, [rax + 2*rax]
  3a3072:      	xor	esi, esi
  3a3074:      	call	0x5fa0e0 <memset@plt>
  3a3079:      	mov	rax, qword ptr [rbp - 0x48]
  3a307d:      	mov	qword ptr [rax + 0x60], 0x0
  3a3085:      	mov	rax, qword ptr [rbp - 0x48]
  3a3089:      	mov	qword ptr [rax + 0x70], 0x0
  3a3091:      	mov	rax, qword ptr [rbp - 0x48]
  3a3095:      	mov	qword ptr [rax + 0x80], 0x0
  3a30a0:      	mov	rax, qword ptr [rbp - 0x48]
  3a30a4:      	mov	qword ptr [rax + 0x118], 0x0
  3a30af:      	xorps	xmm0, xmm0
  3a30b2:      	mov	rax, qword ptr [rbp - 0x48]
  3a30b6:      	movups	xmmword ptr [rax + 0xd8], xmm0
  3a30bd:      	mov	rax, qword ptr [rbp - 0x48]
  3a30c1:      	movups	xmmword ptr [rax + 0xe8], xmm0
  3a30c8:      	mov	rax, qword ptr [rbp - 0x48]
  3a30cc:      	mov	qword ptr [rax + 0xf8], 0x0
  3a30d7:      	mov	rax, qword ptr [rbp - 0x48]
  3a30db:      	mov	rdi, qword ptr [rax + 0x100]
  3a30e2:      	mov	rax, qword ptr [rbp - 0x48]
  3a30e6:      	mov	rdx, qword ptr [rax + 0x108]
  3a30ed:      	shl	rdx, 0x4
  3a30f1:      	xor	esi, esi
  3a30f3:      	call	0x5fa0e0 <memset@plt>
  3a30f8:      	mov	rax, qword ptr [rbp - 0x48]
  3a30fc:      	cmp	qword ptr [rax + 0x110], 0x41
  3a3104:      	jl	0x3a311c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a6c>
  3a3106:      	mov	qword ptr [rax + 0x100], 0x0
  3a3111:      	mov	qword ptr [rax + 0x110], 0x0
  3a311c:      	mov	qword ptr [rax + 0x108], 0x0
  3a3127:      	xorps	xmm0, xmm0
  3a312a:      	movups	xmmword ptr [rax + 0x120], xmm0
  3a3131:      	mov	qword ptr [rax + 0x130], 0x0
  3a313c:      	mov	rdi, qword ptr [rax + 0xa8]
  3a3143:      	mov	rdx, qword ptr [rax + 0xb0]
  3a314a:      	shl	rdx, 0x4
  3a314e:      	xor	esi, esi
  3a3150:      	call	0x5fa0e0 <memset@plt>
  3a3155:      	mov	rax, qword ptr [rbp - 0x48]
  3a3159:      	mov	qword ptr [rax + 0xb0], 0x0
  3a3164:      	mov	rax, qword ptr [rbp - 0x48]
  3a3168:      	mov	rdi, qword ptr [rax + 0xc0]
  3a316f:      	mov	rax, qword ptr [rbp - 0x48]
  3a3173:      	mov	rdx, qword ptr [rax + 0xc8]
  3a317a:      	jmp	0x3a34f4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e44>
  3a317f:      	mov	r13, rax
  3a3182:      	mov	r14, qword ptr [rbp - 0x168]
  3a3189:      	mov	edi, 0x8
  3a318e:      	call	0x5fa310 <GC_malloc@plt>
  3a3193:      	test	rax, rax
  3a3196:      	je	0x3a3694 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fe4>
  3a319c:      	mov	r12, rax
  3a319f:      	mov	edi, 0x8
  3a31a4:      	call	0x5e00e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a31a9:      	mov	edx, 0x29a170
  3a31ae:      	mov	rdi, r13
  3a31b1:      	mov	rsi, r14
  3a31b4:      	mov	rcx, r12
  3a31b7:      	call	0x5b7120 <errors.As>
  3a31bc:      	test	al, 0x1
  3a31be:      	je	0x3a31d1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b21>
  3a31c0:      	mov	rax, qword ptr [r12]
  3a31c4:      	mov	rcx, qword ptr [rbp - 0x88]
  3a31cb:      	mov	ecx, dword ptr [rcx + 0x50]
  3a31ce:      	mov	dword ptr [rax + 0x4], ecx
  3a31d1:      	mov	r15, qword ptr [rbp - 0x48]
  3a31d5:      	mov	qword ptr [r15 + 0x140], r13
  3a31dc:      	mov	qword ptr [r15 + 0x148], r14
  3a31e3:      	mov	byte ptr [r15 + 0x138], 0x3
  3a31eb:      	lea	rdi, [r15 + 0x168]
  3a31f2:      	mov	edx, 0x158
  3a31f7:      	xor	esi, esi
  3a31f9:      	call	0x5fa0e0 <memset@plt>
  3a31fe:      	mov	rdi, qword ptr [r15 + 0x40]
  3a3202:      	mov	rdx, qword ptr [r15 + 0x48]
  3a3206:      	shl	rdx, 0x6
  3a320a:      	xor	esi, esi
  3a320c:      	call	0x5fa0e0 <memset@plt>
  3a3211:      	mov	qword ptr [r15 + 0x48], 0x0
  3a3219:      	mov	rdi, qword ptr [r15 + 0x58]
  3a321d:      	mov	rax, qword ptr [r15 + 0x60]
  3a3221:      	shl	rax, 0x4
  3a3225:      	lea	rdx, [rax + 2*rax]
  3a3229:      	xor	esi, esi
  3a322b:      	call	0x5fa0e0 <memset@plt>
  3a3230:      	mov	qword ptr [r15 + 0x60], 0x0
  3a3238:      	mov	qword ptr [r15 + 0x70], 0x0
  3a3240:      	mov	qword ptr [r15 + 0x80], 0x0
  3a324b:      	mov	qword ptr [r15 + 0x118], 0x0
  3a3256:      	xorps	xmm0, xmm0
  3a3259:      	movups	xmmword ptr [r15 + 0xd8], xmm0
  3a3261:      	movups	xmmword ptr [r15 + 0xe8], xmm0
  3a3269:      	mov	qword ptr [r15 + 0xf8], 0x0
  3a3274:      	mov	rdi, qword ptr [r15 + 0x100]
  3a327b:      	mov	rdx, qword ptr [r15 + 0x108]
  3a3282:      	shl	rdx, 0x4
  3a3286:      	xor	esi, esi
  3a3288:      	call	0x5fa0e0 <memset@plt>
  3a328d:      	cmp	qword ptr [r15 + 0x110], 0x41
  3a3295:      	jl	0x3a32ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bfd>
  3a3297:      	mov	qword ptr [r15 + 0x100], 0x0
  3a32a2:      	mov	qword ptr [r15 + 0x110], 0x0
  3a32ad:      	mov	qword ptr [r15 + 0x108], 0x0
  3a32b8:      	xorps	xmm0, xmm0
  3a32bb:      	movups	xmmword ptr [r15 + 0x120], xmm0
  3a32c3:      	mov	qword ptr [r15 + 0x130], 0x0
  3a32ce:      	mov	rdi, qword ptr [r15 + 0xa8]
  3a32d5:      	mov	rdx, qword ptr [r15 + 0xb0]
  3a32dc:      	shl	rdx, 0x4
  3a32e0:      	xor	esi, esi
  3a32e2:      	call	0x5fa0e0 <memset@plt>
  3a32e7:      	mov	rax, qword ptr [rbp - 0x48]
  3a32eb:      	mov	qword ptr [rax + 0xb0], 0x0
  3a32f6:      	mov	rax, qword ptr [rbp - 0x48]
  3a32fa:      	mov	rdi, qword ptr [rax + 0xc0]
  3a3301:      	mov	rax, qword ptr [rbp - 0x48]
  3a3305:      	mov	rdx, qword ptr [rax + 0xc8]
  3a330c:      	jmp	0x3a34f4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e44>
  3a3311:      	mov	rdi, r14
  3a3314:      	mov	rsi, r8
  3a3317:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a331c:      	mov	rsi, r9
  3a331f:      	call	0x5e01c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3324:      	mov	edi, 0x1
  3a3329:      	call	0x5e4450 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a332e:      	mov	rsi, qword ptr [rbp - 0x38]
  3a3332:      	call	0x5e01c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3337:      	mov	dl, 0x3
  3a3339:      	jmp	0x3a334d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c9d>
  3a333b:      	mov	rdi, r11
  3a333e:      	mov	rsi, rax
  3a3341:      	call	0x3a0d20 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a3346:      	jmp	0x3a352e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e7e>
  3a334b:      	mov	dl, 0x1
  3a334d:      	mov	r15d, edx
  3a3350:      	mov	edi, 0x10
  3a3355:      	call	0x5fa3c0 <GC_malloc_atomic@plt>
  3a335a:      	test	rax, rax
  3a335d:      	je	0x3a3694 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fe4>
  3a3363:      	mov	r14, rax
  3a3366:      	mov	edi, 0x10
  3a336b:      	call	0x5e00e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a3370:      	xorps	xmm0, xmm0
  3a3373:      	movups	xmmword ptr [r14], xmm0
  3a3377:      	mov	rax, qword ptr [rbp - 0x60]
  3a337b:      	mov	rsi, qword ptr [rax + 0x80]
  3a3382:      	mov	edi, dword ptr [r12 + 0x4]
  3a3387:      	cmp	rsi, rdi
  3a338a:      	jbe	0x3a3545 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e95>
  3a3390:      	mov	rax, qword ptr [rax + 0x78]
  3a3394:      	mov	rcx, qword ptr [rbp - 0x88]
  3a339b:      	mov	ecx, dword ptr [rcx + 0x50]
  3a339e:      	imul	rdx, rdi, 0xe0
  3a33a5:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a33aa:      	mov	byte ptr [r14], r15b
  3a33ad:      	mov	dword ptr [r14 + 0x4], ecx
  3a33b1:      	mov	qword ptr [r14 + 0x8], rax
  3a33b5:      	mov	r15, qword ptr [rbp - 0x48]
  3a33b9:      	mov	qword ptr [r15 + 0x140], 0x29a588
  3a33c4:      	mov	qword ptr [r15 + 0x148], r14
  3a33cb:      	mov	byte ptr [r15 + 0x138], 0x3
  3a33d3:      	lea	rdi, [r15 + 0x168]
  3a33da:      	mov	edx, 0x158
  3a33df:      	xor	esi, esi
  3a33e1:      	call	0x5fa0e0 <memset@plt>
  3a33e6:      	mov	rdi, qword ptr [r15 + 0x40]
  3a33ea:      	mov	rdx, qword ptr [r15 + 0x48]
  3a33ee:      	shl	rdx, 0x6
  3a33f2:      	xor	esi, esi
  3a33f4:      	call	0x5fa0e0 <memset@plt>
  3a33f9:      	mov	qword ptr [r15 + 0x48], 0x0
  3a3401:      	mov	rdi, qword ptr [r15 + 0x58]
  3a3405:      	mov	rax, qword ptr [r15 + 0x60]
  3a3409:      	shl	rax, 0x4
  3a340d:      	lea	rdx, [rax + 2*rax]
  3a3411:      	xor	esi, esi
  3a3413:      	call	0x5fa0e0 <memset@plt>
  3a3418:      	mov	qword ptr [r15 + 0x60], 0x0
  3a3420:      	mov	qword ptr [r15 + 0x70], 0x0
  3a3428:      	mov	qword ptr [r15 + 0x80], 0x0
  3a3433:      	mov	qword ptr [r15 + 0x118], 0x0
  3a343e:      	xorps	xmm0, xmm0
  3a3441:      	movups	xmmword ptr [r15 + 0xd8], xmm0
  3a3449:      	movups	xmmword ptr [r15 + 0xe8], xmm0
  3a3451:      	mov	qword ptr [r15 + 0xf8], 0x0
  3a345c:      	mov	rdi, qword ptr [r15 + 0x100]
  3a3463:      	mov	rdx, qword ptr [r15 + 0x108]
  3a346a:      	shl	rdx, 0x4
  3a346e:      	xor	esi, esi
  3a3470:      	call	0x5fa0e0 <memset@plt>
  3a3475:      	cmp	qword ptr [r15 + 0x110], 0x41
  3a347d:      	jl	0x3a3495 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1de5>
  3a347f:      	mov	qword ptr [r15 + 0x100], 0x0
  3a348a:      	mov	qword ptr [r15 + 0x110], 0x0
  3a3495:      	mov	qword ptr [r15 + 0x108], 0x0
  3a34a0:      	xorps	xmm0, xmm0
  3a34a3:      	movups	xmmword ptr [r15 + 0x120], xmm0
  3a34ab:      	mov	qword ptr [r15 + 0x130], 0x0
  3a34b6:      	mov	rdi, qword ptr [r15 + 0xa8]
  3a34bd:      	mov	rdx, qword ptr [r15 + 0xb0]
  3a34c4:      	shl	rdx, 0x4
  3a34c8:      	xor	esi, esi
  3a34ca:      	call	0x5fa0e0 <memset@plt>
  3a34cf:      	mov	rax, qword ptr [rbp - 0x48]
  3a34d3:      	mov	qword ptr [rax + 0xb0], 0x0
  3a34de:      	mov	rax, qword ptr [rbp - 0x48]
  3a34e2:      	mov	rdi, qword ptr [rax + 0xc0]
  3a34e9:      	mov	rax, qword ptr [rbp - 0x48]
  3a34ed:      	mov	rdx, qword ptr [rax + 0xc8]
  3a34f4:      	shl	rdx, 0x4
  3a34f8:      	xor	esi, esi
  3a34fa:      	call	0x5fa0e0 <memset@plt>
  3a34ff:      	mov	rax, qword ptr [rbp - 0x48]
  3a3503:      	cmp	qword ptr [rax + 0xd0], 0x41
  3a350b:      	jl	0x3a3523 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e73>
  3a350d:      	mov	qword ptr [rax + 0xc0], 0x0
  3a3518:      	mov	qword ptr [rax + 0xd0], 0x0
  3a3523:      	mov	qword ptr [rax + 0xc8], 0x0
  3a352e:      	mov	al, 0x3
  3a3530:      	mov	rdx, rbx
  3a3533:      	add	rsp, 0x2f8
  3a353a:      	pop	rbx
  3a353b:      	pop	r12
  3a353d:      	pop	r13
  3a353f:      	pop	r14
  3a3541:      	pop	r15
  3a3543:      	pop	rbp
  3a3544:      	ret
  3a3545:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a354a:      	mov	rsi, r9
  3a354d:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3552:      	mov	rdi, r9
  3a3555:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a355a:      	mov	rdi, rax
  3a355d:      	mov	rsi, qword ptr [rbp - 0x38]
  3a3561:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3566:      	mov	rdi, rcx
  3a3569:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a356e:      	mov	edi, 0x1
  3a3573:      	call	0x5e01c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3578:      	xor	edi, edi
  3a357a:      	call	0x5e01c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a357f:      	mov	rdx, rsi
  3a3582:      	xor	esi, esi
  3a3584:      	mov	ecx, 0x3
  3a3589:      	call	0x5eb050 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a358e:      	mov	rdi, rsi
  3a3591:      	xor	esi, esi
  3a3593:      	mov	ecx, 0x2
  3a3598:      	call	0x5eb050 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a359d:      	mov	rdi, rsi
  3a35a0:      	mov	rsi, r9
  3a35a3:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a35a8:      	mov	r8, qword ptr [rbp - 0x68]
  3a35ac:      	mov	rdi, r12
  3a35af:      	mov	rsi, r8
  3a35b2:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a35b7:      	mov	rdi, rax
  3a35ba:      	mov	esi, 0x1
  3a35bf:      	mov	ecx, 0x2
  3a35c4:      	call	0x5eb050 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a35c9:      	mov	rdi, r13
  3a35cc:      	mov	esi, 0x1
  3a35d1:      	mov	rdx, rax
  3a35d4:      	mov	ecx, 0x3
  3a35d9:      	call	0x5eb050 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a35de:      	mov	rsi, qword ptr [rbp - 0x38]
  3a35e2:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a35e7:      	mov	edi, 0x3
  3a35ec:      	call	0x5e01c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a35f1:      	mov	edi, 0x2
  3a35f6:      	mov	esi, 0x2
  3a35fb:      	call	0x5e01c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3600:      	mov	dl, 0x9
  3a3602:      	jmp	0x3a334d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c9d>
  3a3607:      	mov	al, 0x1
  3a3609:      	jmp	0x3a3530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e80>
  3a360e:      	mov	rdx, qword ptr [rbp - 0x110]
  3a3615:      	jmp	0x3a361e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f6e>
  3a3617:      	mov	rdx, qword ptr [rbp - 0x138]
  3a361e:      	mov	rdi, r11
  3a3621:      	call	0x3a0d20 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a3626:      	jmp	0x3a352e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e7e>
  3a362b:      	mov	rdi, r12
  3a362e:      	mov	rsi, r12
  3a3631:      	call	0x5e01c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3636:      	mov	rdi, rax
  3a3639:      	mov	rsi, r9
  3a363c:      	call	0x5e01c0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3641:      	mov	edi, 0x3296a8
  3a3646:      	mov	esi, 0x3299d8
  3a364b:      	call	0x5dfe00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3650:      	mov	rdi, rdx
  3a3653:      	call	0x5e16f0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3658:      	mov	rdi, r12
  3a365b:      	xor	esi, esi
  3a365d:      	mov	ecx, 0x2
  3a3662:      	call	0x5eb050 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a3667:      	xor	esi, esi
  3a3669:      	mov	rdx, r12
  3a366c:      	mov	ecx, 0x3
  3a3671:      	call	0x5eb050 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a3676:      	mov	edi, 0x333b58
  3a367b:      	mov	esi, 0x29a210
  3a3680:      	call	0x5dfe00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3685:      	mov	edi, 0x333b58
  3a368a:      	mov	esi, 0x29a1f0
  3a368f:      	call	0x5dfe00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3694:      	mov	edi, 0x333b58
  3a3699:      	mov	esi, 0x329680
  3a369e:      	call	0x5dfe00 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a36a3:      	int3
  3a36a4:      	int3
  3a36a5:      	int3
  3a36a6:      	int3
  3a36a7:      	int3
  3a36a8:      	int3
  3a36a9:      	int3
  3a36aa:      	int3
  3a36ab:      	int3
  3a36ac:      	int3
  3a36ad:      	int3
  3a36ae:      	int3
  3a36af:      	int3
