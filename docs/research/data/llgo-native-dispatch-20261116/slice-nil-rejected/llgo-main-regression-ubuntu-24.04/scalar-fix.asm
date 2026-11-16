
/home/runner/work/wasmvm/wasmvm/main-regression/scalar-fix:	file format elf64-x86-64

Disassembly of section .text:

00000000003a1530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a1530:      	push	rbp
  3a1531:      	mov	rbp, rsp
  3a1534:      	push	r15
  3a1536:      	push	r14
  3a1538:      	push	r13
  3a153a:      	push	r12
  3a153c:      	push	rbx
  3a153d:      	sub	rsp, 0x2a8
  3a1544:      	mov	rcx, qword ptr [rdi + 0x48]
  3a1548:      	test	rcx, rcx
  3a154b:      	jle	0x3a3764 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2234>
  3a1551:      	mov	rbx, rsi
  3a1554:      	mov	r11, rdi
  3a1557:      	movzx	eax, byte ptr [rdi + 0x20]
  3a155b:      	mov	byte ptr [rbp - 0x39], al
  3a155e:      	mov	r8d, 0xffffffff
  3a1564:      	lea	rax, [rbp - 0x90]
  3a156b:      	mov	qword ptr [rbp - 0x70], rax
  3a156f:      	mov	qword ptr [rbp - 0x30], rdi
  3a1573:      	mov	rdi, qword ptr [r11 + 0x40]
  3a1577:      	shl	rcx, 0x6
  3a157b:      	mov	byte ptr [rdi + rcx - 0x30], 0x0
  3a1580:      	mov	rsi, qword ptr [rdi + rcx - 0x40]
  3a1585:      	mov	r10, qword ptr [rsi + 0x48]
  3a1589:      	mov	r13, qword ptr [r10 + 0x98]
  3a1590:      	mov	r9, qword ptr [rdi + rcx - 0x28]
  3a1595:      	mov	rdx, qword ptr [r11 + 0x88]
  3a159c:      	lea	rax, [r9 + r13]
  3a15a0:      	cmp	rax, rdx
  3a15a3:      	ja	0x3a36d5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21a5>
  3a15a9:      	cmp	r9, rax
  3a15ac:      	ja	0x3a36e7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21b7>
  3a15b2:      	mov	r15, qword ptr [r10 + 0xb8]
  3a15b9:      	mov	r12d, dword ptr [rdi + rcx - 0x38]
  3a15be:      	sub	rdx, r9
  3a15c1:      	test	rdx, rdx
  3a15c4:      	mov	eax, 0x0
  3a15c9:      	cmovle	r9, rax
  3a15cd:      	cmp	r15, r12
  3a15d0:      	jbe	0x3a364c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x211c>
  3a15d6:      	add	rdi, rcx
  3a15d9:      	mov	qword ptr [rbp - 0x60], rdi
  3a15dd:      	mov	qword ptr [rbp - 0xa8], rsi
  3a15e4:      	mov	rax, qword ptr [rsi + 0x40]
  3a15e8:      	mov	qword ptr [rbp - 0x50], rax
  3a15ec:      	mov	rcx, qword ptr [r10 + 0xb0]
  3a15f3:      	shl	r9, 0x4
  3a15f7:      	add	r9, qword ptr [r11 + 0x78]
  3a15fb:      	mov	r14d, r12d
  3a15fe:      	mov	qword ptr [rbp - 0x78], r13
  3a1602:      	mov	qword ptr [rbp - 0x38], r10
  3a1606:      	mov	qword ptr [rbp - 0x48], r9
  3a160a:      	mov	qword ptr [rbp - 0xb0], r15
  3a1611:      	mov	qword ptr [rbp - 0x58], rcx
  3a1615:      	jmp	0x3a164d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11d>
  3a1617:      	nop	word ptr [rax + rax]
  3a1620:      	shl	rdi, 0x4
  3a1624:      	mov	qword ptr [r9 + rdi], rax
  3a1628:      	mov	qword ptr [r9 + rdi + 0x8], rcx
  3a162d:      	inc	r14d
  3a1630:      	mov	r15d, r14d
  3a1633:      	mov	rcx, qword ptr [rbp - 0x58]
  3a1637:      	mov	r12d, r15d
  3a163a:      	mov	r14d, r15d
  3a163d:      	mov	r15, qword ptr [rbp - 0xb0]
  3a1644:      	cmp	r15, r12
  3a1647:      	jbe	0x3a364c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x211c>
  3a164d:      	shl	r12, 0x5
  3a1651:      	add	r12, rcx
  3a1654:      	test	byte ptr [rbp - 0x39], 0x1
  3a1658:      	je	0x3a16b5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x185>
  3a165a:      	mov	rsi, qword ptr [r10 + 0x80]
  3a1661:      	mov	edi, dword ptr [r12 + 0x4]
  3a1666:      	cmp	rsi, rdi
  3a1669:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a166f:      	mov	r15, qword ptr [r10 + 0x78]
  3a1673:      	imul	r13, rdi, 0xe0
  3a167a:      	mov	eax, dword ptr [r15 + r13]
  3a167e:      	cmp	eax, 0x5
  3a1681:      	setne	cl
  3a1684:      	cmp	eax, 0xb
  3a1687:      	setne	al
  3a168a:      	and	al, cl
  3a168c:      	movzx	ecx, al
  3a168f:      	mov	rax, qword ptr [r11 + 0x150]
  3a1696:      	sub	rax, rcx
  3a1699:      	jb	0x3a2f4e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a1e>
  3a169f:      	mov	qword ptr [r11 + 0x150], rax
  3a16a6:      	mov	r13, qword ptr [rbp - 0x78]
  3a16aa:      	mov	r15, qword ptr [rbp - 0xb0]
  3a16b1:      	mov	rcx, qword ptr [rbp - 0x58]
  3a16b5:      	inc	rbx
  3a16b8:      	test	bl, 0x3f
  3a16bb:      	jne	0x3a173b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20b>
  3a16bd:      	mov	rax, qword ptr [r11 + 0x160]
  3a16c4:      	test	rax, rax
  3a16c7:      	je	0x3a16dc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ac>
  3a16c9:      	test	r11, r11
  3a16cc:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a16d2:      	mov	eax, dword ptr [rax]
  3a16d4:      	test	eax, eax
  3a16d6:      	je	0x3a17d2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2a2>
  3a16dc:      	mov	rax, qword ptr [r11 + 0x8]
  3a16e0:      	test	rax, rax
  3a16e3:      	je	0x3a3755 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2225>
  3a16e9:      	mov	rdi, qword ptr [r11 + 0x10]
  3a16ed:      	mov	rcx, qword ptr [rax + 0x8]
  3a16f1:      	test	byte ptr [rcx + 0x17], 0x20
  3a16f5:      	je	0x3a170e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1de>
  3a16f7:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a16fb:      	and	cl, 0x1f
  3a16fe:      	cmp	cl, 0x16
  3a1701:      	je	0x3a170e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1de>
  3a1703:      	mov	qword ptr [rbp - 0x90], rdi
  3a170a:      	mov	rdi, qword ptr [rbp - 0x70]
  3a170e:      	call	qword ptr [rax + 0x28]
  3a1711:      	mov	r11, qword ptr [rbp - 0x30]
  3a1715:      	mov	r8d, 0xffffffff
  3a171b:      	mov	r10, qword ptr [rbp - 0x38]
  3a171f:      	mov	r9, qword ptr [rbp - 0x48]
  3a1723:      	mov	rcx, qword ptr [rbp - 0x58]
  3a1727:      	test	rax, rax
  3a172a:      	je	0x3a17df <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2af>
  3a1730:      	cmp	qword ptr [rax + 0x8], 0x0
  3a1735:      	jne	0x3a30d8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ba8>
  3a173b:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a1741:      	je	0x3a17f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a1747:      	test	byte ptr [rbp - 0x39], 0x1
  3a174b:      	je	0x3a1770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a174d:      	mov	edx, dword ptr [r12]
  3a1751:      	cmp	edx, 0x20
  3a1754:      	je	0x3a17f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a175a:      	cmp	edx, 0x41
  3a175d:      	je	0x3a17f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a1763:      	jmp	0x3a17f4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c4>
  3a1768:      	nop	dword ptr [rax + rax]
  3a1770:      	mov	edx, dword ptr [r12]
  3a1774:      	cmp	edx, 0x20
  3a1777:      	je	0x3a177e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x24e>
  3a1779:      	cmp	edx, 0x41
  3a177c:      	jne	0x3a17f4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c4>
  3a177e:      	mov	eax, ebx
  3a1780:      	not	eax
  3a1782:      	test	al, 0x3f
  3a1784:      	je	0x3a17f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a1786:      	cmp	edx, 0x41
  3a1789:      	jne	0x3a17af <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x27f>
  3a178b:      	mov	edi, dword ptr [r12 + 0x8]
  3a1790:      	cmp	r13, rdi
  3a1793:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a1799:      	mov	eax, dword ptr [r12 + 0x18]
  3a179e:      	shl	rdi, 0x4
  3a17a2:      	mov	qword ptr [r9 + rdi], rax
  3a17a6:      	mov	qword ptr [r9 + rdi + 0x8], 0x0
  3a17af:      	inc	r14d
  3a17b2:      	cmp	r15, r14
  3a17b5:      	jbe	0x3a362a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20fa>
  3a17bb:      	mov	rax, r14
  3a17be:      	shl	rax, 0x5
  3a17c2:      	lea	r12, [rcx + rax]
  3a17c6:      	inc	rbx
  3a17c9:      	cmp	dword ptr [rcx + rax + 0x1c], 0x0
  3a17ce:      	jne	0x3a1770 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x240>
  3a17d0:      	jmp	0x3a17f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2c0>
  3a17d2:      	xor	eax, eax
  3a17d4:      	xor	edx, edx
  3a17d6:      	test	rax, rax
  3a17d9:      	jne	0x3a1730 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x200>
  3a17df:      	cmp	dword ptr [r12 + 0x1c], 0x0
  3a17e5:      	jne	0x3a1747 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x217>
  3a17eb:      	nop	dword ptr [rax + rax]
  3a17f0:      	mov	edx, dword ptr [r12]
  3a17f4:      	lea	eax, [rdx - 0xe0]
  3a17fa:      	cmp	eax, 0xffff20
  3a17ff:      	jae	0x3a1a60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x530>
  3a1805:      	mov	eax, edx
  3a1807:      	shr	eax, 0x10
  3a180a:      	cmp	eax, 0xfb
  3a180f:      	je	0x3a19c5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x495>
  3a1815:      	cmp	eax, 0xfc
  3a181a:      	je	0x3a18a4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x374>
  3a1820:      	cmp	eax, 0xfd
  3a1825:      	jne	0x3a1a49 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x519>
  3a182b:      	mov	edi, dword ptr [r12 + 0xc]
  3a1830:      	cmp	rdi, r8
  3a1833:      	je	0x3a1cf0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7c0>
  3a1839:      	cmp	r13, rdi
  3a183c:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a1842:      	shl	rdi, 0x4
  3a1846:      	mov	rcx, qword ptr [r9 + rdi]
  3a184a:      	mov	r10, qword ptr [r9 + rdi + 0x8]
  3a184f:      	mov	edi, dword ptr [r12 + 0x10]
  3a1854:      	mov	esi, 0xffffffff
  3a1859:      	cmp	rdi, r8
  3a185c:      	je	0x3a1d08 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7d8>
  3a1862:      	cmp	r13, rdi
  3a1865:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a186b:      	shl	rdi, 0x4
  3a186f:      	mov	rax, qword ptr [r9 + rdi]
  3a1873:      	mov	r8, qword ptr [r9 + rdi + 0x8]
  3a1878:      	mov	edi, dword ptr [r12 + 0x14]
  3a187d:      	cmp	rdi, rsi
  3a1880:      	je	0x3a1d1b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7eb>
  3a1886:      	cmp	r13, rdi
  3a1889:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a188f:      	shl	rdi, 0x4
  3a1893:      	mov	rsi, r9
  3a1896:      	mov	r9, qword ptr [r9 + rdi]
  3a189a:      	mov	r11, qword ptr [rsi + rdi + 0x8]
  3a189f:      	jmp	0x3a1d21 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7f1>
  3a18a4:      	add	edx, 0xff03ffed
  3a18aa:      	cmp	edx, 0x4
  3a18ad:      	jae	0x3a1a60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x530>
  3a18b3:      	mov	rdx, qword ptr [r10 + 0xd8]
  3a18ba:      	mov	edi, dword ptr [r12 + 0x18]
  3a18bf:      	mov	eax, dword ptr [r12 + 0x1c]
  3a18c4:      	lea	esi, [rdi + rax]
  3a18c7:      	add	esi, 0x2
  3a18ca:      	cmp	rdx, rsi
  3a18cd:      	jb	0x3a36c6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2196>
  3a18d3:      	cmp	edi, esi
  3a18d5:      	ja	0x3a36b7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2187>
  3a18db:      	mov	eax, edi
  3a18dd:      	mov	ecx, esi
  3a18df:      	sub	rsi, rdi
  3a18e2:      	sub	rdx, rdi
  3a18e5:      	test	rdx, rdx
  3a18e8:      	mov	edx, 0x0
  3a18ed:      	cmovle	rdi, rdx
  3a18f1:      	cmp	ecx, eax
  3a18f3:      	je	0x3a36b0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2180>
  3a18f9:      	mov	rax, qword ptr [r10 + 0xc8]
  3a1900:      	lea	r8, [rax + 4*rdi]
  3a1904:      	mov	edi, dword ptr [r8]
  3a1907:      	cmp	r13, rdi
  3a190a:      	jbe	0x3a3678 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2148>
  3a1910:      	cmp	rsi, 0x1
  3a1914:      	jbe	0x3a36a6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2176>
  3a191a:      	mov	eax, dword ptr [r8 + 0x4]
  3a191e:      	cmp	r13, rax
  3a1921:      	jbe	0x3a3688 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2158>
  3a1927:      	shl	rdi, 0x4
  3a192b:      	mov	rcx, r9
  3a192e:      	mov	r9, qword ptr [r9 + rdi]
  3a1932:      	shl	rax, 0x4
  3a1936:      	mov	rcx, qword ptr [rcx + rax]
  3a193a:      	mov	r10d, dword ptr [r12]
  3a193e:      	lea	eax, [r10 - 0xfc0013]
  3a1945:      	cmp	eax, 0x2
  3a1948:      	jae	0x3a1f06 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9d6>
  3a194e:      	cmp	rsi, 0x2
  3a1952:      	je	0x3a3706 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21d6>
  3a1958:      	mov	edi, dword ptr [r8 + 0x8]
  3a195c:      	cmp	r13, rdi
  3a195f:      	jbe	0x3a3678 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2148>
  3a1965:      	cmp	rsi, 0x3
  3a1969:      	jbe	0x3a36fc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21cc>
  3a196f:      	mov	eax, dword ptr [r8 + 0xc]
  3a1973:      	cmp	r13, rax
  3a1976:      	jbe	0x3a3688 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2158>
  3a197c:      	shl	rdi, 0x4
  3a1980:      	mov	rdx, qword ptr [rbp - 0x48]
  3a1984:      	mov	rdi, qword ptr [rdx + rdi]
  3a1988:      	shl	rax, 0x4
  3a198c:      	mov	rdx, qword ptr [rdx + rax]
  3a1990:      	cmp	r10d, 0xfc0013
  3a1997:      	jne	0x3a1f9f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa6f>
  3a199d:      	lea	rax, [rdi + r9]
  3a19a1:      	mov	r10, rdi
  3a19a4:      	and	r10, r9
  3a19a7:      	or	rdi, r9
  3a19aa:      	mov	r9, rax
  3a19ad:      	not	r9
  3a19b0:      	and	r9, rdi
  3a19b3:      	or	r9, r10
  3a19b6:      	shr	r9, 0x3f
  3a19ba:      	add	rdx, rcx
  3a19bd:      	add	rdx, r9
  3a19c0:      	jmp	0x3a1fc4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa94>
  3a19c5:      	add	edx, -0x2
  3a19c8:      	cmp	dx, 0x21
  3a19cc:      	jae	0x3a1bd9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6a9>
  3a19d2:      	lea	rax, [r8 + 0x1c003e10]
  3a19d9:      	movzx	ecx, dx
  3a19dc:      	bt	rax, rcx
  3a19e0:      	jae	0x3a1bd9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6a9>
  3a19e6:      	mov	edi, dword ptr [r12 + 0xc]
  3a19eb:      	mov	eax, 0xffffffff
  3a19f0:      	cmp	rdi, r8
  3a19f3:      	je	0x3a2069 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb39>
  3a19f9:      	cmp	r13, rdi
  3a19fc:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a1a02:      	shl	rdi, 0x4
  3a1a06:      	mov	r8, qword ptr [r9 + rdi]
  3a1a0a:      	mov	edi, dword ptr [r12 + 0x10]
  3a1a0f:      	cmp	rdi, rax
  3a1a12:      	je	0x3a207a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb4a>
  3a1a18:      	cmp	r13, rdi
  3a1a1b:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a1a21:      	shl	rdi, 0x4
  3a1a25:      	mov	rax, qword ptr [r9 + rdi]
  3a1a29:      	mov	r9, qword ptr [r9 + rdi + 0x8]
  3a1a2e:      	mov	r10d, 0x0
  3a1a34:      	mov	r11d, 0x0
  3a1a3a:      	cmp	dx, 0xc
  3a1a3e:      	je	0x3a2091 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb61>
  3a1a44:      	jmp	0x3a20d1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xba1>
  3a1a49:      	cmp	edx, 0xe7
  3a1a4f:      	jb	0x3a2ed0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19a0>
  3a1a55:      	nop	word ptr cs:[rax + rax]
  3a1a60:      	mov	edi, dword ptr [r12 + 0xc]
  3a1a65:      	cmp	rdi, r8
  3a1a68:      	je	0x3a1b00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d0>
  3a1a6e:      	cmp	r13, rdi
  3a1a71:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a1a77:      	mov	rax, rdi
  3a1a7a:      	shl	rax, 0x4
  3a1a7e:      	mov	r15, qword ptr [r9 + rax]
  3a1a82:      	mov	ecx, dword ptr [r12 + 0x10]
  3a1a87:      	cmp	rcx, r8
  3a1a8a:      	je	0x3a1b11 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5e1>
  3a1a90:      	cmp	r13, rcx
  3a1a93:      	jbe	0x3a366d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x213d>
  3a1a99:      	mov	rax, rcx
  3a1a9c:      	shl	rax, 0x4
  3a1aa0:      	mov	rdx, qword ptr [r9 + rax]
  3a1aa4:      	mov	r9d, dword ptr [r12]
  3a1aa8:      	lea	eax, [r9 - 0x45]
  3a1aac:      	cmp	eax, 0x7f
  3a1aaf:      	ja	0x3a1b20 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5f0>
  3a1ab1:      	ja	0x3a2adf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15af>
  3a1ab7:      	jmp	qword ptr [8*rax + 0x203938]
  3a1abe:      	mov	edi, r9d
  3a1ac1:      	mov	rsi, r15
  3a1ac4:      	call	0x3d02f0 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a1ac9:      	test	dl, dl
  3a1acb:      	jne	0x3a324d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d1d>
  3a1ad1:      	xor	ecx, ecx
  3a1ad3:      	mov	r11, qword ptr [rbp - 0x30]
  3a1ad7:      	mov	r8d, 0xffffffff
  3a1add:      	mov	r10, qword ptr [rbp - 0x38]
  3a1ae1:      	mov	edi, dword ptr [r12 + 0x8]
  3a1ae6:      	cmp	rdi, r8
  3a1ae9:      	mov	r9, qword ptr [rbp - 0x48]
  3a1aed:      	jne	0x3a2640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1110>
  3a1af3:      	jmp	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a1af8:      	nop	dword ptr [rax + rax]
  3a1b00:      	xor	r15d, r15d
  3a1b03:      	mov	ecx, dword ptr [r12 + 0x10]
  3a1b08:      	cmp	rcx, r8
  3a1b0b:      	jne	0x3a1a90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x560>
  3a1b11:      	xor	edx, edx
  3a1b13:      	mov	r9d, dword ptr [r12]
  3a1b17:      	lea	eax, [r9 - 0x45]
  3a1b1b:      	cmp	eax, 0x7f
  3a1b1e:      	jbe	0x3a1ab1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x581>
  3a1b20:      	lea	eax, [r9 - 0x1]
  3a1b24:      	cmp	eax, 0xd3
  3a1b29:      	ja	0x3a1c16 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6e6>
  3a1b2f:      	jmp	qword ptr [8*rax + 0x203258]
  3a1b36:      	mov	rsi, qword ptr [r10 + 0x80]
  3a1b3d:      	mov	edi, dword ptr [r12 + 0x4]
  3a1b42:      	cmp	rsi, rdi
  3a1b45:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a1b4b:      	imul	rax, rdi, 0xe0
  3a1b52:      	add	rax, qword ptr [r10 + 0x78]
  3a1b56:      	mov	r10, qword ptr [rbp - 0x70]
  3a1b5a:      	mov	rcx, qword ptr [rbp - 0x50]
  3a1b5e:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1b64:      	mov	rsi, qword ptr [rcx + 0x78]
  3a1b68:      	mov	edi, dword ptr [rax + 0x18]
  3a1b6b:      	cmp	rsi, rdi
  3a1b6e:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a1b74:      	shl	rdi, 0x3
  3a1b78:      	add	rdi, qword ptr [rcx + 0x70]
  3a1b7c:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1b82:      	mov	rcx, qword ptr [rdi]
  3a1b85:      	test	rcx, rcx
  3a1b88:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1b8e:      	mov	rsi, qword ptr [rcx + 0x10]
  3a1b92:      	mov	ecx, dword ptr [rax + 0x14]
  3a1b95:      	mov	r8d, dword ptr [rax + 0x1c]
  3a1b99:      	shl	r8, 0x20
  3a1b9d:      	or	r8, rcx
  3a1ba0:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a1ba5:      	je	0x3a1bb2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x682>
  3a1ba7:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a1bac:      	je	0x3a1f5e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa2e>
  3a1bb2:      	xorps	xmm0, xmm0
  3a1bb5:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a1bbc:      	movups	xmmword ptr [rsp], xmm0
  3a1bc0:      	mov	rdi, r10
  3a1bc3:      	mov	edx, r9d
  3a1bc6:      	mov	rcx, r15
  3a1bc9:      	xor	r9d, r9d
  3a1bcc:      	mov	r15, r10
  3a1bcf:      	call	0x3c68f0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a1bd4:      	jmp	0x3a1f80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa50>
  3a1bd9:      	mov	r15, qword ptr [rbp - 0x60]
  3a1bdd:      	mov	dword ptr [r15 - 0x38], r14d
  3a1be1:      	mov	rdi, r11
  3a1be4:      	call	0x3a39d0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a1be9:      	mov	r9, qword ptr [rbp - 0x48]
  3a1bed:      	mov	r10, qword ptr [rbp - 0x38]
  3a1bf1:      	mov	r8d, 0xffffffff
  3a1bf7:      	mov	r11, qword ptr [rbp - 0x30]
  3a1bfb:      	test	al, al
  3a1bfd:      	mov	rcx, qword ptr [rbp - 0x58]
  3a1c01:      	jne	0x3a35fb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20cb>
  3a1c07:      	inc	r14d
  3a1c0a:      	mov	dword ptr [r15 - 0x38], r14d
  3a1c0e:      	mov	r15d, r14d
  3a1c11:      	jmp	0x3a1637 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x107>
  3a1c16:      	lea	eax, [r9 - 0xfc0000]
  3a1c1d:      	cmp	eax, 0x8
  3a1c20:      	jae	0x3a1e0b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8db>
  3a1c26:      	mov	edi, r9d
  3a1c29:      	mov	rsi, r15
  3a1c2c:      	call	0x3d02f0 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a1c31:      	mov	r10, qword ptr [rbp - 0x38]
  3a1c35:      	mov	r8d, 0xffffffff
  3a1c3b:      	mov	r11, qword ptr [rbp - 0x30]
  3a1c3f:      	mov	r15d, edx
  3a1c42:      	xor	ecx, ecx
  3a1c44:      	jmp	0x3a2625 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10f5>
  3a1c49:      	mov	rsi, qword ptr [r10 + 0x80]
  3a1c50:      	mov	edi, dword ptr [r12 + 0x4]
  3a1c55:      	cmp	rsi, rdi
  3a1c58:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a1c5e:      	imul	rax, rdi, 0xe0
  3a1c65:      	add	rax, qword ptr [r10 + 0x78]
  3a1c69:      	mov	rcx, qword ptr [rbp - 0x50]
  3a1c6d:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1c73:      	mov	rsi, qword ptr [rcx + 0x78]
  3a1c77:      	mov	edi, dword ptr [rax + 0x18]
  3a1c7a:      	cmp	rsi, rdi
  3a1c7d:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a1c83:      	shl	rdi, 0x3
  3a1c87:      	add	rdi, qword ptr [rcx + 0x70]
  3a1c8b:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1c91:      	mov	rcx, qword ptr [rdi]
  3a1c94:      	test	rcx, rcx
  3a1c97:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1c9d:      	mov	rdi, qword ptr [rcx + 0x10]
  3a1ca1:      	mov	esi, dword ptr [rax + 0x14]
  3a1ca4:      	mov	ecx, dword ptr [rax + 0x1c]
  3a1ca7:      	shl	rcx, 0x20
  3a1cab:      	or	rcx, rsi
  3a1cae:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a1cb3:      	je	0x3a1cc0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x790>
  3a1cb5:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a1cba:      	je	0x3a203e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb0e>
  3a1cc0:      	mov	qword ptr [rbp - 0x90], rdx
  3a1cc7:      	mov	qword ptr [rbp - 0x88], 0x0
  3a1cd2:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a1cd9:      	movups	xmmword ptr [rsp], xmm0
  3a1cdd:      	mov	esi, r9d
  3a1ce0:      	mov	rdx, r15
  3a1ce3:      	xor	r8d, r8d
  3a1ce6:      	call	0x3c7c60 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a1ceb:      	jmp	0x3a2610 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10e0>
  3a1cf0:      	xor	r10d, r10d
  3a1cf3:      	xor	ecx, ecx
  3a1cf5:      	mov	edi, dword ptr [r12 + 0x10]
  3a1cfa:      	mov	esi, 0xffffffff
  3a1cff:      	cmp	rdi, r8
  3a1d02:      	jne	0x3a1862 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x332>
  3a1d08:      	xor	eax, eax
  3a1d0a:      	xor	r8d, r8d
  3a1d0d:      	mov	edi, dword ptr [r12 + 0x14]
  3a1d12:      	cmp	rdi, rsi
  3a1d15:      	jne	0x3a1886 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x356>
  3a1d1b:      	xor	r9d, r9d
  3a1d1e:      	xor	r11d, r11d
  3a1d21:      	mov	r15, qword ptr [rbp - 0x38]
  3a1d25:      	mov	rsi, qword ptr [r15 + 0x80]
  3a1d2c:      	mov	edi, dword ptr [r12 + 0x4]
  3a1d31:      	cmp	rsi, rdi
  3a1d34:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a1d3a:      	imul	rdi, rdi, 0xe0
  3a1d41:      	add	rdi, qword ptr [r15 + 0x78]
  3a1d45:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1d4b:      	movzx	esi, dx
  3a1d4e:      	cmp	esi, 0xb
  3a1d51:      	setb	r15b
  3a1d55:      	mov	esi, edx
  3a1d57:      	and	esi, 0xfffc
  3a1d5d:      	cmp	esi, 0x54
  3a1d60:      	sete	r13b
  3a1d64:      	or	r13b, r15b
  3a1d67:      	jne	0x3a1d7b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x84b>
  3a1d69:      	movzx	r15d, dx
  3a1d6d:      	lea	r13d, [r15 - 0x5c]
  3a1d71:      	cmp	r13d, 0x2
  3a1d75:      	jae	0x3a1e69 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x939>
  3a1d7b:      	mov	r10, qword ptr [rbp - 0x50]
  3a1d7f:      	mov	rsi, qword ptr [r10 + 0x78]
  3a1d83:      	mov	r9d, dword ptr [rdi + 0x18]
  3a1d87:      	cmp	rsi, r9
  3a1d8a:      	jbe	0x3a3680 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2150>
  3a1d90:      	shl	r9, 0x3
  3a1d94:      	add	r9, qword ptr [r10 + 0x70]
  3a1d98:      	mov	r11, qword ptr [rbp - 0x70]
  3a1d9c:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1da2:      	mov	rsi, qword ptr [r9]
  3a1da5:      	test	rsi, rsi
  3a1da8:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1dae:      	mov	rsi, qword ptr [rsi + 0x10]
  3a1db2:      	mov	r9d, dword ptr [rdi + 0x14]
  3a1db6:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a1dba:      	shl	r10, 0x20
  3a1dbe:      	or	r10, r9
  3a1dc1:      	movzx	edi, byte ptr [rdi + 0x40]
  3a1dc5:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a1dca:      	mov	r13, qword ptr [rbp - 0x78]
  3a1dce:      	je	0x3a1ddb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8ab>
  3a1dd0:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a1dd5:      	je	0x3a1f36 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa06>
  3a1ddb:      	mov	qword ptr [rbp - 0xa0], rax
  3a1de2:      	mov	qword ptr [rbp - 0x98], r8
  3a1de9:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a1df0:      	movups	xmmword ptr [rsp], xmm0
  3a1df4:      	movzx	r9d, dil
  3a1df8:      	mov	rdi, r11
  3a1dfb:      	mov	r8, r10
  3a1dfe:      	mov	r15, r11
  3a1e01:      	call	0x3c68f0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a1e06:      	jmp	0x3a1f80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa50>
  3a1e0b:      	lea	eax, [r9 - 0x1000001]
  3a1e12:      	cmp	eax, 0x8
  3a1e15:      	ja	0x3a2b80 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1650>
  3a1e1b:      	mov	ecx, 0x117
  3a1e20:      	bt	ecx, eax
  3a1e23:      	jae	0x3a2b6e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x163e>
  3a1e29:      	movzx	eax, r9b
  3a1e2d:      	xor	edi, edi
  3a1e2f:      	add	eax, -0x2
  3a1e32:      	cmp	eax, 0x7
  3a1e35:      	ja	0x3a2b2b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15fb>
  3a1e3b:      	mov	r9, qword ptr [rbp - 0x48]
  3a1e3f:      	jmp	qword ptr [8*rax + 0x2038f8]
  3a1e46:      	xor	edi, edi
  3a1e48:      	test	r15d, r15d
  3a1e4b:      	sete	dil
  3a1e4f:      	mov	rsi, qword ptr [r10 + 0xe8]
  3a1e56:      	add	edi, dword ptr [r12 + 0x18]
  3a1e5b:      	cmp	rsi, rdi
  3a1e5e:      	ja	0x3a22cf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd9f>
  3a1e64:      	jmp	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a1e69:      	cmp	r15d, 0xb
  3a1e6d:      	je	0x3a1e78 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x948>
  3a1e6f:      	cmp	esi, 0x58
  3a1e72:      	jne	0x3a226b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd3b>
  3a1e78:      	mov	r10, qword ptr [rbp - 0x50]
  3a1e7c:      	mov	rsi, qword ptr [r10 + 0x78]
  3a1e80:      	mov	r9d, dword ptr [rdi + 0x18]
  3a1e84:      	cmp	rsi, r9
  3a1e87:      	mov	r13, qword ptr [rbp - 0x78]
  3a1e8b:      	jbe	0x3a3680 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2150>
  3a1e91:      	shl	r9, 0x3
  3a1e95:      	add	r9, qword ptr [r10 + 0x70]
  3a1e99:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1e9f:      	mov	rsi, qword ptr [r9]
  3a1ea2:      	test	rsi, rsi
  3a1ea5:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a1eab:      	mov	rsi, qword ptr [rsi + 0x10]
  3a1eaf:      	mov	r10d, dword ptr [rdi + 0x14]
  3a1eb3:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a1eb7:      	shl	r9, 0x20
  3a1ebb:      	or	r9, r10
  3a1ebe:      	movzx	edi, byte ptr [rdi + 0x40]
  3a1ec2:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a1ec7:      	je	0x3a1ed4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9a4>
  3a1ec9:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a1ece:      	je	0x3a25e3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10b3>
  3a1ed4:      	mov	qword ptr [rbp - 0x90], rax
  3a1edb:      	mov	qword ptr [rbp - 0x88], r8
  3a1ee2:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a1ee9:      	movups	xmmword ptr [rsp], xmm0
  3a1eed:      	movzx	r8d, dil
  3a1ef1:      	mov	rdi, rsi
  3a1ef4:      	mov	esi, edx
  3a1ef6:      	mov	rdx, rcx
  3a1ef9:      	mov	rcx, r9
  3a1efc:      	call	0x3c7c60 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a1f01:      	jmp	0x3a2610 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10e0>
  3a1f06:      	mov	rax, rcx
  3a1f09:      	mul	r9
  3a1f0c:      	cmp	dword ptr [r12], 0xfc0015
  3a1f14:      	jne	0x3a1fc4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa94>
  3a1f1a:      	mov	rdi, r9
  3a1f1d:      	sar	rdi, 0x3f
  3a1f21:      	and	rdi, rcx
  3a1f24:      	sar	rcx, 0x3f
  3a1f28:      	and	rcx, r9
  3a1f2b:      	add	rcx, rdi
  3a1f2e:      	sub	rdx, rcx
  3a1f31:      	jmp	0x3a1fc4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa94>
  3a1f36:      	mov	qword ptr [rbp - 0xa0], rax
  3a1f3d:      	mov	qword ptr [rbp - 0x98], r8
  3a1f44:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a1f4b:      	movups	xmmword ptr [rsp], xmm0
  3a1f4f:      	movzx	r9d, dil
  3a1f53:      	mov	rdi, r11
  3a1f56:      	mov	r8, r10
  3a1f59:      	mov	r15, r11
  3a1f5c:      	jmp	0x3a1f7b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa4b>
  3a1f5e:      	xorps	xmm0, xmm0
  3a1f61:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a1f68:      	movups	xmmword ptr [rsp], xmm0
  3a1f6c:      	mov	rdi, r10
  3a1f6f:      	mov	edx, r9d
  3a1f72:      	mov	rcx, r15
  3a1f75:      	xor	r9d, r9d
  3a1f78:      	mov	r15, r10
  3a1f7b:      	call	0x3c6a30 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a1f80:      	mov	rax, qword ptr [r15]
  3a1f83:      	mov	rcx, qword ptr [r15 + 0x8]
  3a1f87:      	mov	qword ptr [rbp - 0x70], r15
  3a1f8b:      	movzx	r15d, byte ptr [r15 + 0x10]
  3a1f90:      	jmp	0x3a2617 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10e7>
  3a1f95:      	mov	rax, qword ptr [r12 + 0x18]
  3a1f9a:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a1f9f:      	mov	rax, r9
  3a1fa2:      	sub	rax, rdi
  3a1fa5:      	not	r9
  3a1fa8:      	mov	r10, rdi
  3a1fab:      	and	r10, r9
  3a1fae:      	xor	r9, rdi
  3a1fb1:      	and	r9, rax
  3a1fb4:      	or	r9, r10
  3a1fb7:      	shr	r9, 0x3f
  3a1fbb:      	add	rdx, r9
  3a1fbe:      	sub	rcx, rdx
  3a1fc1:      	mov	rdx, rcx
  3a1fc4:      	mov	r10, qword ptr [rbp - 0x38]
  3a1fc8:      	mov	edi, dword ptr [r12 + 0x1c]
  3a1fcd:      	cmp	rsi, rdi
  3a1fd0:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a1fd6:      	lea	ecx, [rdi + 0x1]
  3a1fd9:      	cmp	rsi, rcx
  3a1fdc:      	mov	r9, qword ptr [rbp - 0x60]
  3a1fe0:      	jbe	0x3a3693 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2163>
  3a1fe6:      	mov	esi, dword ptr [r8 + 4*rdi]
  3a1fea:      	cmp	r13, rsi
  3a1fed:      	jbe	0x3a369b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x216b>
  3a1ff3:      	mov	edi, dword ptr [r8 + 4*rcx]
  3a1ff7:      	shl	rsi, 0x4
  3a1ffb:      	mov	rcx, qword ptr [rbp - 0x48]
  3a1fff:      	mov	qword ptr [rcx + rsi], rax
  3a2003:      	mov	qword ptr [rcx + rsi + 0x8], 0x0
  3a200c:      	cmp	r13, rdi
  3a200f:      	jbe	0x3a3678 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2148>
  3a2015:      	shl	rdi, 0x4
  3a2019:      	mov	qword ptr [rcx + rdi], rdx
  3a201d:      	mov	qword ptr [rcx + rdi + 0x8], 0x0
  3a2026:      	inc	r14d
  3a2029:      	mov	dword ptr [r9 - 0x38], r14d
  3a202d:      	mov	r15d, r14d
  3a2030:      	mov	r8d, 0xffffffff
  3a2036:      	mov	r9, rcx
  3a2039:      	jmp	0x3a1633 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x103>
  3a203e:      	mov	qword ptr [rbp - 0x90], rdx
  3a2045:      	mov	qword ptr [rbp - 0x88], 0x0
  3a2050:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a2057:      	movups	xmmword ptr [rsp], xmm0
  3a205b:      	mov	esi, r9d
  3a205e:      	mov	rdx, r15
  3a2061:      	xor	r8d, r8d
  3a2064:      	jmp	0x3a260b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10db>
  3a2069:      	xor	r8d, r8d
  3a206c:      	mov	edi, dword ptr [r12 + 0x10]
  3a2071:      	cmp	rdi, rax
  3a2074:      	jne	0x3a1a18 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4e8>
  3a207a:      	xor	eax, eax
  3a207c:      	xor	r9d, r9d
  3a207f:      	mov	r10d, 0x0
  3a2085:      	mov	r11d, 0x0
  3a208b:      	cmp	dx, 0xc
  3a208f:      	jne	0x3a20d1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xba1>
  3a2091:      	mov	rcx, qword ptr [rbp - 0x38]
  3a2095:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a209c:      	mov	edi, dword ptr [r12 + 0x18]
  3a20a1:      	add	edi, 0x2
  3a20a4:      	cmp	rsi, rdi
  3a20a7:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a20ad:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a20b4:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a20b7:      	cmp	r13, rdi
  3a20ba:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a20c0:      	shl	rdi, 0x4
  3a20c4:      	mov	rcx, qword ptr [rbp - 0x48]
  3a20c8:      	mov	r10, qword ptr [rcx + rdi]
  3a20cc:      	mov	r11, qword ptr [rcx + rdi + 0x8]
  3a20d1:      	cmp	qword ptr [rbp - 0x30], 0x0
  3a20d6:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a20dc:      	mov	rdx, qword ptr [rbp - 0x38]
  3a20e0:      	mov	rsi, qword ptr [rdx + 0x80]
  3a20e7:      	mov	edi, dword ptr [r12 + 0x4]
  3a20ec:      	cmp	rsi, rdi
  3a20ef:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a20f5:      	imul	rcx, rdi, 0xe0
  3a20fc:      	add	rcx, qword ptr [rdx + 0x78]
  3a2100:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2106:      	mov	rdx, qword ptr [rbp - 0x30]
  3a210a:      	mov	rsi, qword ptr [rdx]
  3a210d:      	mov	qword ptr [rbp - 0xd0], rax
  3a2114:      	mov	qword ptr [rbp - 0xc8], r9
  3a211b:      	mov	qword ptr [rbp - 0xc0], r10
  3a2122:      	mov	qword ptr [rbp - 0xb8], r11
  3a2129:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a2130:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a2135:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a213c:      	movups	xmmword ptr [rsp], xmm0
  3a2140:      	lea	rdi, [rbp - 0x160]
  3a2147:      	mov	rdx, qword ptr [rbp - 0x50]
  3a214b:      	call	0x3b6af0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a2150:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a2157:      	mov	r15, qword ptr [rbp - 0x150]
  3a215e:      	test	r15, r15
  3a2161:      	je	0x3a2181 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc51>
  3a2163:      	cmp	qword ptr [r15 + 0x8], 0x0
  3a2168:      	mov	r8d, 0xffffffff
  3a216e:      	mov	r10, qword ptr [rbp - 0x38]
  3a2172:      	mov	r9, qword ptr [rbp - 0x48]
  3a2176:      	mov	rcx, qword ptr [rbp - 0x58]
  3a217a:      	je	0x3a2193 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc63>
  3a217c:      	jmp	0x3a3255 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d25>
  3a2181:      	mov	r8d, 0xffffffff
  3a2187:      	mov	r10, qword ptr [rbp - 0x38]
  3a218b:      	mov	r9, qword ptr [rbp - 0x48]
  3a218f:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2193:      	mov	edi, dword ptr [r12 + 0x8]
  3a2198:      	cmp	rdi, r8
  3a219b:      	mov	r11, qword ptr [rbp - 0x30]
  3a219f:      	mov	r15, qword ptr [rbp - 0x60]
  3a21a3:      	je	0x3a1c07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6d7>
  3a21a9:      	cmp	r13, rdi
  3a21ac:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a21b2:      	shl	rdi, 0x4
  3a21b6:      	movups	xmmword ptr [r9 + rdi], xmm0
  3a21bb:      	jmp	0x3a1c07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6d7>
  3a21c0:      	cmp	r13, rdi
  3a21c3:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a21c9:      	mov	edx, dword ptr [r12 + 0x18]
  3a21ce:      	cmp	r13, rdx
  3a21d1:      	mov	rsi, qword ptr [rbp - 0x48]
  3a21d5:      	jbe	0x3a3736 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2206>
  3a21db:      	shl	rdi, 0x4
  3a21df:      	mov	rax, qword ptr [rsi + rdi]
  3a21e3:      	mov	rcx, qword ptr [rsi + rdi + 0x8]
  3a21e8:      	shl	rdx, 0x4
  3a21ec:      	mov	qword ptr [rsi + rdx], rax
  3a21f0:      	mov	qword ptr [rsi + rdx + 0x8], rcx
  3a21f5:      	mov	edi, dword ptr [r12 + 0x8]
  3a21fa:      	cmp	rdi, r8
  3a21fd:      	mov	r9, qword ptr [rbp - 0x48]
  3a2201:      	jne	0x3a2640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1110>
  3a2207:      	jmp	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a220c:      	mov	eax, dword ptr [r12 + 0x14]
  3a2211:      	cmp	rax, r8
  3a2214:      	je	0x3a2afa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15ca>
  3a221a:      	cmp	r13, rax
  3a221d:      	mov	rdx, qword ptr [rbp - 0x48]
  3a2221:      	jbe	0x3a3657 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2127>
  3a2227:      	cmp	r13, rcx
  3a222a:      	jbe	0x3a3741 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2211>
  3a2230:      	shl	rax, 0x4
  3a2234:      	cmp	dword ptr [rdx + rax], 0x0
  3a2238:      	je	0x3a2b07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15d7>
  3a223e:      	cmp	r13, rdi
  3a2241:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a2247:      	shl	rdi, 0x4
  3a224b:      	mov	rax, qword ptr [rdx + rdi]
  3a224f:      	mov	rcx, qword ptr [rdx + rdi + 0x8]
  3a2254:      	mov	edi, dword ptr [r12 + 0x8]
  3a2259:      	cmp	rdi, r8
  3a225c:      	mov	r9, qword ptr [rbp - 0x48]
  3a2260:      	jne	0x3a2640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1110>
  3a2266:      	jmp	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a226b:      	mov	qword ptr [rsp], r11
  3a226f:      	mov	rsi, rcx
  3a2272:      	mov	rdx, r10
  3a2275:      	mov	rcx, rax
  3a2278:      	call	0x3ce4b0 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a227d:      	mov	rcx, rdx
  3a2280:      	mov	r11, qword ptr [rbp - 0x30]
  3a2284:      	mov	r13, qword ptr [rbp - 0x78]
  3a2288:      	jmp	0x3a1ad7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5a7>
  3a228d:      	mov	edi, dword ptr [r12 + 0x1c]
  3a2292:      	dec	edi
  3a2294:      	cmp	edi, r15d
  3a2297:      	cmovae	edi, r15d
  3a229b:      	mov	rsi, qword ptr [r10 + 0xe8]
  3a22a2:      	add	edi, dword ptr [r12 + 0x18]
  3a22a7:      	cmp	rsi, rdi
  3a22aa:      	ja	0x3a22cf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd9f>
  3a22ac:      	jmp	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a22b1:      	xor	edi, edi
  3a22b3:      	test	r15, r15
  3a22b6:      	setne	dil
  3a22ba:      	mov	rsi, qword ptr [r10 + 0xe8]
  3a22c1:      	add	edi, dword ptr [r12 + 0x18]
  3a22c6:      	cmp	rsi, rdi
  3a22c9:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a22cf:      	mov	rax, qword ptr [r10 + 0xe0]
  3a22d6:      	lea	rcx, [rdi + 2*rdi]
  3a22da:      	mov	r15d, dword ptr [rax + 4*rcx]
  3a22de:      	mov	esi, dword ptr [rax + 4*rcx + 0x4]
  3a22e2:      	mov	r12d, dword ptr [rax + 4*rcx + 0x8]
  3a22e7:      	cmp	r15d, r14d
  3a22ea:      	ja	0x3a2694 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1164>
  3a22f0:      	mov	rax, qword ptr [r11 + 0x160]
  3a22f7:      	test	rax, rax
  3a22fa:      	je	0x3a230f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xddf>
  3a22fc:      	test	r11, r11
  3a22ff:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2305:      	mov	eax, dword ptr [rax]
  3a2307:      	test	eax, eax
  3a2309:      	je	0x3a2b49 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1619>
  3a230f:      	mov	rax, qword ptr [r11 + 0x8]
  3a2313:      	test	rax, rax
  3a2316:      	je	0x3a3755 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2225>
  3a231c:      	mov	rdi, qword ptr [r11 + 0x10]
  3a2320:      	mov	rcx, qword ptr [rax + 0x8]
  3a2324:      	test	byte ptr [rcx + 0x17], 0x20
  3a2328:      	jne	0x3a264e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x111e>
  3a232e:      	mov	r14, rsi
  3a2331:      	jmp	0x3a2668 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1138>
  3a2336:      	xor	eax, eax
  3a2338:      	cmp	r15, rdx
  3a233b:      	sete	al
  3a233e:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2343:      	mov	rax, qword ptr [rbp - 0x50]
  3a2347:      	mov	rsi, qword ptr [rax + 0x78]
  3a234b:      	mov	edi, dword ptr [r12 + 0x18]
  3a2350:      	cmp	rsi, rdi
  3a2353:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a2359:      	shl	rdi, 0x3
  3a235d:      	add	rdi, qword ptr [rax + 0x70]
  3a2361:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2367:      	test	r11, r11
  3a236a:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2370:      	mov	rcx, r15
  3a2373:      	mov	r15, qword ptr [rdi]
  3a2376:      	test	r15, r15
  3a2379:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a237f:      	mov	rdi, qword ptr [r11]
  3a2382:      	mov	rax, qword ptr [r15 + 0x10]
  3a2386:      	movzx	eax, byte ptr [rax + 0x31]
  3a238a:      	mov	byte ptr [rbp - 0x61], al
  3a238d:      	mov	edx, ecx
  3a238f:      	test	al, 0x1
  3a2391:      	cmovne	rdx, rcx
  3a2395:      	mov	rsi, r15
  3a2398:      	call	0x3c0d60 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a239d:      	test	dl, 0x1
  3a23a0:      	jne	0x3a1ad1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5a1>
  3a23a6:      	mov	rax, qword ptr [r15 + 0x10]
  3a23aa:      	movzx	eax, byte ptr [rax + 0x31]
  3a23ae:      	mov	byte ptr [rbp - 0x62], al
  3a23b1:      	and	al, 0x1
  3a23b3:      	neg	al
  3a23b5:      	mov	eax, 0x0
  3a23ba:      	sbb	rax, rax
  3a23bd:      	mov	r8d, 0xffffffff
  3a23c3:      	or	rax, r8
  3a23c6:      	jmp	0x3a25ce <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x109e>
  3a23cb:      	mov	rax, qword ptr [rbp - 0x50]
  3a23cf:      	mov	rsi, qword ptr [rax + 0x60]
  3a23d3:      	mov	edi, dword ptr [r12 + 0x18]
  3a23d8:      	cmp	rsi, rdi
  3a23db:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a23e1:      	mov	rax, qword ptr [rax + 0x58]
  3a23e5:      	lea	rcx, [rax + 8*rdi]
  3a23e9:      	test	rcx, rcx
  3a23ec:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a23f2:      	mov	rax, qword ptr [rax + 8*rdi]
  3a23f6:      	test	rax, rax
  3a23f9:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a23ff:      	cmp	byte ptr [rax + 0x29], 0x0
  3a2403:      	mov	ecx, r15d
  3a2406:      	cmovne	rcx, r15
  3a240a:      	cmp	rcx, qword ptr [rax + 0x40]
  3a240e:      	jae	0x3a33fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ece>
  3a2414:      	mov	rax, qword ptr [rax + 0x38]
  3a2418:      	mov	rax, qword ptr [rax + 8*rcx]
  3a241c:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2421:      	mov	rax, qword ptr [rbp - 0x50]
  3a2425:      	mov	rsi, qword ptr [rax + 0x48]
  3a2429:      	mov	edi, dword ptr [r12 + 0x18]
  3a242e:      	cmp	rsi, rdi
  3a2431:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a2437:      	mov	rax, qword ptr [rax + 0x40]
  3a243b:      	mov	rax, qword ptr [rax + 8*rdi]
  3a243f:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2444:      	test	r15, r15
  3a2447:      	jne	0x3a253c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100c>
  3a244d:      	jmp	0x3a3418 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ee8>
  3a2452:      	mov	rax, qword ptr [rbp - 0x50]
  3a2456:      	mov	rsi, qword ptr [rax + 0x90]
  3a245d:      	mov	eax, dword ptr [r12 + 0x18]
  3a2462:      	cmp	rsi, rax
  3a2465:      	jbe	0x3a374d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x221d>
  3a246b:      	cmp	r13, rdi
  3a246e:      	mov	rdx, qword ptr [rbp - 0x48]
  3a2472:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a2478:      	mov	rcx, qword ptr [rbp - 0x50]
  3a247c:      	mov	rcx, qword ptr [rcx + 0x88]
  3a2483:      	mov	rax, qword ptr [rcx + 8*rax]
  3a2487:      	shl	rdi, 0x4
  3a248b:      	movups	xmm0, xmmword ptr [rdx + rdi]
  3a248f:      	movups	xmmword ptr [rax + 0x20], xmm0
  3a2493:      	jmp	0x3a252c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xffc>
  3a2498:      	mov	rax, qword ptr [rbp - 0x50]
  3a249c:      	mov	rsi, qword ptr [rax + 0x90]
  3a24a3:      	mov	edi, dword ptr [r12 + 0x18]
  3a24a8:      	cmp	rsi, rdi
  3a24ab:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a24b1:      	mov	rax, qword ptr [rax + 0x88]
  3a24b8:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a24bc:      	mov	rax, qword ptr [rcx + 0x20]
  3a24c0:      	mov	rcx, qword ptr [rcx + 0x28]
  3a24c4:      	mov	edi, dword ptr [r12 + 0x8]
  3a24c9:      	cmp	rdi, r8
  3a24cc:      	mov	r9, qword ptr [rbp - 0x48]
  3a24d0:      	jne	0x3a2640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1110>
  3a24d6:      	jmp	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a24db:      	mov	rax, qword ptr [rbp - 0x50]
  3a24df:      	mov	rsi, qword ptr [rax + 0x60]
  3a24e3:      	mov	edi, dword ptr [r12 + 0x18]
  3a24e8:      	cmp	rsi, rdi
  3a24eb:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a24f1:      	mov	rax, qword ptr [rax + 0x58]
  3a24f5:      	lea	rcx, [rax + 8*rdi]
  3a24f9:      	test	rcx, rcx
  3a24fc:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2502:      	mov	rax, qword ptr [rax + 8*rdi]
  3a2506:      	test	rax, rax
  3a2509:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a250f:      	cmp	byte ptr [rax + 0x29], 0x0
  3a2513:      	mov	ecx, r15d
  3a2516:      	cmovne	rcx, r15
  3a251a:      	cmp	rcx, qword ptr [rax + 0x40]
  3a251e:      	jae	0x3a33fe <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ece>
  3a2524:      	mov	rax, qword ptr [rax + 0x38]
  3a2528:      	mov	qword ptr [rax + 8*rcx], rdx
  3a252c:      	xor	eax, eax
  3a252e:      	nop
  3a2530:      	xor	ecx, ecx
  3a2532:      	jmp	0x3a262e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10fe>
  3a2537:      	mov	edi, dword ptr [r12 + 0x18]
  3a253c:      	cmp	r13, rdi
  3a253f:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a2545:      	shl	rdi, 0x4
  3a2549:      	mov	rcx, qword ptr [rbp - 0x48]
  3a254d:      	mov	rax, qword ptr [rcx + rdi]
  3a2551:      	mov	rcx, qword ptr [rcx + rdi + 0x8]
  3a2556:      	mov	edi, dword ptr [r12 + 0x8]
  3a255b:      	cmp	rdi, r8
  3a255e:      	mov	r9, qword ptr [rbp - 0x48]
  3a2562:      	jne	0x3a2640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1110>
  3a2568:      	jmp	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a256d:      	xor	eax, eax
  3a256f:      	test	r15, r15
  3a2572:      	sete	al
  3a2575:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2577:      	mov	rax, qword ptr [rbp - 0x50]
  3a257b:      	mov	rsi, qword ptr [rax + 0x78]
  3a257f:      	mov	edi, dword ptr [r12 + 0x18]
  3a2584:      	cmp	rsi, rdi
  3a2587:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a258d:      	shl	rdi, 0x3
  3a2591:      	add	rdi, qword ptr [rax + 0x70]
  3a2595:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a259b:      	mov	rax, qword ptr [rdi]
  3a259e:      	test	rax, rax
  3a25a1:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a25a7:      	mov	rdi, qword ptr [rax + 0x10]
  3a25ab:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a25b0:      	mov	rax, qword ptr [rdi + 0x60]
  3a25b4:      	je	0x3a2d71 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1841>
  3a25ba:      	test	rax, rax
  3a25bd:      	jne	0x3a2d7a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x184a>
  3a25c3:      	call	0x3c45b0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a25c8:      	mov	r8d, 0xffffffff
  3a25ce:      	xor	ecx, ecx
  3a25d0:      	mov	r11, qword ptr [rbp - 0x30]
  3a25d4:      	jmp	0x3a1add <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5ad>
  3a25d9:      	mov	eax, dword ptr [r12 + 0x18]
  3a25de:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a25e3:      	mov	qword ptr [rbp - 0x90], rax
  3a25ea:      	mov	qword ptr [rbp - 0x88], r8
  3a25f1:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a25f8:      	movups	xmmword ptr [rsp], xmm0
  3a25fc:      	movzx	r8d, dil
  3a2600:      	mov	rdi, rsi
  3a2603:      	mov	esi, edx
  3a2605:      	mov	rdx, rcx
  3a2608:      	mov	rcx, r9
  3a260b:      	call	0x3c7f40 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a2610:      	mov	r15d, eax
  3a2613:      	xor	eax, eax
  3a2615:      	xor	ecx, ecx
  3a2617:      	mov	r11, qword ptr [rbp - 0x30]
  3a261b:      	mov	r8d, 0xffffffff
  3a2621:      	mov	r10, qword ptr [rbp - 0x38]
  3a2625:      	test	r15b, r15b
  3a2628:      	jne	0x3a341b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eeb>
  3a262e:      	mov	edi, dword ptr [r12 + 0x8]
  3a2633:      	cmp	rdi, r8
  3a2636:      	mov	r9, qword ptr [rbp - 0x48]
  3a263a:      	je	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a2640:      	cmp	r13, rdi
  3a2643:      	ja	0x3a1620 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf0>
  3a2649:      	jmp	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a264e:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a2652:      	and	cl, 0x1f
  3a2655:      	mov	r14, rsi
  3a2658:      	cmp	cl, 0x16
  3a265b:      	je	0x3a2668 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1138>
  3a265d:      	mov	qword ptr [rbp - 0x90], rdi
  3a2664:      	mov	rdi, qword ptr [rbp - 0x70]
  3a2668:      	call	qword ptr [rax + 0x28]
  3a266b:      	mov	r11, qword ptr [rbp - 0x30]
  3a266f:      	mov	r8d, 0xffffffff
  3a2675:      	mov	r10, qword ptr [rbp - 0x38]
  3a2679:      	mov	r9, qword ptr [rbp - 0x48]
  3a267d:      	mov	rsi, r14
  3a2680:      	test	rax, rax
  3a2683:      	je	0x3a2b56 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1626>
  3a2689:      	cmp	qword ptr [rax + 0x8], 0x0
  3a268e:      	jne	0x3a3403 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ed3>
  3a2694:      	mov	rdx, qword ptr [r10 + 0x108]
  3a269b:      	add	r12d, esi
  3a269e:      	cmp	rdx, r12
  3a26a1:      	jb	0x3a3727 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21f7>
  3a26a7:      	cmp	esi, r12d
  3a26aa:      	ja	0x3a3715 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21e5>
  3a26b0:      	sub	rdx, rsi
  3a26b3:      	test	rdx, rdx
  3a26b6:      	mov	eax, 0x0
  3a26bb:      	cmovg	rax, rsi
  3a26bf:      	sub	r12, rsi
  3a26c2:      	jle	0x3a1633 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x103>
  3a26c8:      	mov	rcx, qword ptr [r10 + 0xf8]
  3a26cf:      	lea	rcx, [rcx + 8*rax]
  3a26d3:      	add	rcx, 0x4
  3a26d7:      	xor	edx, edx
  3a26d9:      	nop	dword ptr [rax]
  3a26e0:      	cmp	r12, rdx
  3a26e3:      	je	0x3a3662 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2132>
  3a26e9:      	mov	edi, dword ptr [rcx + 8*rdx]
  3a26ec:      	cmp	r13, rdi
  3a26ef:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a26f5:      	mov	eax, dword ptr [rcx + 8*rdx - 0x4]
  3a26f9:      	cmp	r13, rax
  3a26fc:      	jbe	0x3a3657 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2127>
  3a2702:      	shl	rdi, 0x4
  3a2706:      	shl	rax, 0x4
  3a270a:      	movups	xmm0, xmmword ptr [r9 + rdi]
  3a270f:      	movups	xmmword ptr [r9 + rax], xmm0
  3a2714:      	inc	rdx
  3a2717:      	cmp	r12, rdx
  3a271a:      	jne	0x3a26e0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x11b0>
  3a271c:      	jmp	0x3a1633 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x103>
  3a2721:      	xor	rdx, r15
  3a2724:      	jmp	0x3a2a74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1544>
  3a2729:      	mov	eax, r15d
  3a272c:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2731:      	mov	ecx, edx
  3a2733:      	shr	r15, cl
  3a2736:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a273b:      	mov	ecx, edx
  3a273d:      	shr	r15d, cl
  3a2740:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a2745:      	or	rdx, r15
  3a2748:      	jmp	0x3a2a74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1544>
  3a274d:      	mov	ecx, edx
  3a274f:      	sar	r15d, cl
  3a2752:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a2757:      	xor	edx, r15d
  3a275a:      	jmp	0x3a2a74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1544>
  3a275f:      	mov	rax, r15
  3a2762:      	shr	rax
  3a2765:      	movabs	rcx, 0x5555555555555555
  3a276f:      	and	rax, rcx
  3a2772:      	and	r15, rcx
  3a2775:      	add	r15, rax
  3a2778:      	mov	rax, r15
  3a277b:      	shr	rax, 0x2
  3a277f:      	movabs	rcx, 0x3333333333333333
  3a2789:      	and	rax, rcx
  3a278c:      	and	r15, rcx
  3a278f:      	add	r15, rax
  3a2792:      	mov	rax, r15
  3a2795:      	shr	rax, 0x4
  3a2799:      	add	rax, r15
  3a279c:      	movabs	rcx, 0xf0f0f0f0f0f0f0f
  3a27a6:      	and	rcx, rax
  3a27a9:      	mov	rax, rcx
  3a27ac:      	shr	rax, 0x8
  3a27b0:      	add	rax, rcx
  3a27b3:      	mov	rcx, rax
  3a27b6:      	shr	rcx, 0x10
  3a27ba:      	add	rcx, rax
  3a27bd:      	mov	rax, rcx
  3a27c0:      	shr	rax, 0x20
  3a27c4:      	add	eax, ecx
  3a27c6:      	and	eax, 0x7f
  3a27c9:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a27ce:      	xor	eax, eax
  3a27d0:      	cmp	r15, rdx
  3a27d3:      	setb	al
  3a27d6:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a27db:      	xor	eax, eax
  3a27dd:      	cmp	r15, rdx
  3a27e0:      	setbe	al
  3a27e3:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a27e8:      	xor	eax, eax
  3a27ea:      	cmp	r15, rdx
  3a27ed:      	seta	al
  3a27f0:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a27f5:      	imul	edx, r15d
  3a27f9:      	jmp	0x3a2a74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1544>
  3a27fe:      	mov	ecx, r15d
  3a2801:      	shr	ecx, 0x10
  3a2804:      	xor	eax, eax
  3a2806:      	cmp	r15d, 0x10000
  3a280d:      	setb	al
  3a2810:      	cmovb	ecx, r15d
  3a2814:      	shl	eax, 0x4
  3a2817:      	mov	edx, ecx
  3a2819:      	shr	edx, 0x8
  3a281c:      	xor	esi, esi
  3a281e:      	cmp	ecx, 0x100
  3a2824:      	setb	sil
  3a2828:      	lea	eax, [rax + 8*rsi]
  3a282b:      	cmovb	edx, ecx
  3a282e:      	movzx	ecx, dl
  3a2831:      	movzx	ecx, byte ptr [rcx + 0x328441]
  3a2838:      	sub	rax, rcx
  3a283b:      	add	rax, 0x8
  3a283f:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2844:      	mov	ecx, edx
  3a2846:      	rol	r15d, cl
  3a2849:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a284e:      	xor	eax, eax
  3a2850:      	cmp	r15, rdx
  3a2853:      	setg	al
  3a2856:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a285b:      	xor	eax, eax
  3a285d:      	cmp	r15d, edx
  3a2860:      	setl	al
  3a2863:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2868:      	xor	eax, eax
  3a286a:      	cmp	r15, rdx
  3a286d:      	setne	al
  3a2870:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2875:      	test	r15d, r15d
  3a2878:      	je	0x3a2db8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1888>
  3a287e:      	tzcnt	eax, r15d
  3a2883:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2888:      	add	edx, r15d
  3a288b:      	jmp	0x3a2a74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1544>
  3a2890:      	xor	eax, eax
  3a2892:      	cmp	r15d, edx
  3a2895:      	setne	al
  3a2898:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a289d:      	xor	eax, eax
  3a289f:      	cmp	r15, rdx
  3a28a2:      	setle	al
  3a28a5:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a28aa:      	mov	ecx, edx
  3a28ac:      	ror	r15d, cl
  3a28af:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a28b4:      	xor	eax, eax
  3a28b6:      	cmp	r15d, edx
  3a28b9:      	setbe	al
  3a28bc:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a28c1:      	xor	eax, eax
  3a28c3:      	cmp	r15d, edx
  3a28c6:      	setge	al
  3a28c9:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a28ce:      	xor	eax, eax
  3a28d0:      	test	r15d, r15d
  3a28d3:      	sete	al
  3a28d6:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a28db:      	xor	eax, eax
  3a28dd:      	cmp	r15d, edx
  3a28e0:      	setb	al
  3a28e3:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a28e8:      	mov	rdx, r15
  3a28eb:      	mov	eax, edx
  3a28ed:      	shr	eax, 0x18
  3a28f0:      	mov	ecx, edx
  3a28f2:      	shr	ecx, 0x10
  3a28f5:      	movzx	ecx, cl
  3a28f8:      	movzx	ecx, byte ptr [rcx + 0x328541]
  3a28ff:      	add	cl, byte ptr [rax + 0x328541]
  3a2905:      	movzx	eax, dh
  3a2908:      	add	cl, byte ptr [rax + 0x328541]
  3a290e:      	movzx	eax, dl
  3a2911:      	add	cl, byte ptr [rax + 0x328541]
  3a2917:      	movzx	eax, cl
  3a291a:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a291f:      	add	rdx, r15
  3a2922:      	jmp	0x3a2a74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1544>
  3a2927:      	mov	ecx, edx
  3a2929:      	shl	r15d, cl
  3a292c:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a2931:      	xor	eax, eax
  3a2933:      	test	r15, r15
  3a2936:      	sete	al
  3a2939:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a293e:      	and	rdx, r15
  3a2941:      	jmp	0x3a2a74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1544>
  3a2946:      	sub	r15, rdx
  3a2949:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a294e:      	xor	eax, eax
  3a2950:      	cmp	r15d, edx
  3a2953:      	sete	al
  3a2956:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a295b:      	mov	ecx, edx
  3a295d:      	rol	r15, cl
  3a2960:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a2965:      	xor	eax, eax
  3a2967:      	cmp	r15d, edx
  3a296a:      	setae	al
  3a296d:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2972:      	and	r15d, edx
  3a2975:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a297a:      	xor	eax, eax
  3a297c:      	cmp	r15d, edx
  3a297f:      	setle	al
  3a2982:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2987:      	xor	eax, eax
  3a2989:      	cmp	r15d, edx
  3a298c:      	setg	al
  3a298f:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2994:      	xor	eax, eax
  3a2996:      	mov	rcx, r15
  3a2999:      	shr	rcx, 0x20
  3a299d:      	cmove	rcx, r15
  3a29a1:      	setne	al
  3a29a4:      	shl	eax, 0x5
  3a29a7:      	mov	rdx, rcx
  3a29aa:      	shr	rdx, 0x10
  3a29ae:      	lea	rsi, [rax + 0x10]
  3a29b2:      	cmp	rcx, 0x10000
  3a29b9:      	cmovb	rdx, rcx
  3a29bd:      	cmovb	rsi, rax
  3a29c1:      	mov	eax, edx
  3a29c3:      	shr	eax, 0x8
  3a29c6:      	lea	rcx, [rsi + 0x8]
  3a29ca:      	cmp	rdx, 0x100
  3a29d1:      	cmovb	rax, rdx
  3a29d5:      	cmovb	rcx, rsi
  3a29d9:      	movzx	eax, al
  3a29dc:      	movzx	edx, byte ptr [rax + 0x328441]
  3a29e3:      	add	rdx, rcx
  3a29e6:      	mov	eax, 0x40
  3a29eb:      	sub	rax, rdx
  3a29ee:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a29f3:      	xor	eax, eax
  3a29f5:      	cmp	r15, rdx
  3a29f8:      	setl	al
  3a29fb:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a00:      	mov	ecx, edx
  3a2a02:      	ror	r15, cl
  3a2a05:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a2a0a:      	xor	eax, eax
  3a2a0c:      	cmp	r15d, edx
  3a2a0f:      	seta	al
  3a2a12:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a17:      	mov	eax, r15d
  3a2a1a:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a1f:      	movsx	rax, r15w
  3a2a23:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a28:      	movsxd	rax, r15d
  3a2a2b:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a30:      	xor	eax, eax
  3a2a32:      	cmp	r15, rdx
  3a2a35:      	sete	al
  3a2a38:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a3d:      	movsxd	rax, r15d
  3a2a40:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a45:      	xor	eax, eax
  3a2a47:      	cmp	r15, rdx
  3a2a4a:      	setae	al
  3a2a4d:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a52:      	xor	eax, eax
  3a2a54:      	cmp	r15, rdx
  3a2a57:      	setge	al
  3a2a5a:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a5f:      	mov	ecx, edx
  3a2a61:      	sar	r15, cl
  3a2a64:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a2a66:      	or	edx, r15d
  3a2a69:      	jmp	0x3a2a74 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1544>
  3a2a6b:      	sub	r15d, edx
  3a2a6e:      	jmp	0x3a2aaf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157f>
  3a2a70:      	imul	rdx, r15
  3a2a74:      	xor	ecx, ecx
  3a2a76:      	mov	rax, rdx
  3a2a79:      	mov	edi, dword ptr [r12 + 0x8]
  3a2a7e:      	cmp	rdi, r8
  3a2a81:      	mov	r9, qword ptr [rbp - 0x48]
  3a2a85:      	jne	0x3a2640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1110>
  3a2a8b:      	jmp	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a2a90:      	movsx	rax, r15w
  3a2a94:      	mov	eax, eax
  3a2a96:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2a9b:      	mov	eax, 0x40
  3a2aa0:      	tzcnt	rax, r15
  3a2aa5:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2aaa:      	mov	ecx, edx
  3a2aac:      	shl	r15, cl
  3a2aaf:      	xor	ecx, ecx
  3a2ab1:      	mov	rax, r15
  3a2ab4:      	mov	edi, dword ptr [r12 + 0x8]
  3a2ab9:      	cmp	rdi, r8
  3a2abc:      	mov	r9, qword ptr [rbp - 0x48]
  3a2ac0:      	jne	0x3a2640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1110>
  3a2ac6:      	jmp	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a2acb:      	movsx	rax, r15b
  3a2acf:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2ad4:      	movsx	rax, r15b
  3a2ad8:      	mov	eax, eax
  3a2ada:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2adf:      	movzx	esi, r9b
  3a2ae3:      	add	esi, -0x61
  3a2ae6:      	cmp	esi, 0x5e
  3a2ae9:      	ja	0x3a252c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xffc>
  3a2aef:      	xor	eax, eax
  3a2af1:      	xor	ecx, ecx
  3a2af3:      	jmp	qword ptr [8*rsi + 0x203d38]
  3a2afa:      	cmp	r13, rcx
  3a2afd:      	mov	rdx, qword ptr [rbp - 0x48]
  3a2b01:      	jbe	0x3a3741 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2211>
  3a2b07:      	shl	rcx, 0x4
  3a2b0b:      	mov	rax, qword ptr [rdx + rcx]
  3a2b0f:      	mov	rcx, qword ptr [rdx + rcx + 0x8]
  3a2b14:      	mov	edi, dword ptr [r12 + 0x8]
  3a2b19:      	cmp	rdi, r8
  3a2b1c:      	mov	r9, qword ptr [rbp - 0x48]
  3a2b20:      	jne	0x3a2640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1110>
  3a2b26:      	jmp	0x3a162d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfd>
  3a2b2b:      	mov	r9, qword ptr [rbp - 0x48]
  3a2b2f:      	mov	rsi, qword ptr [r10 + 0xe8]
  3a2b36:      	add	edi, dword ptr [r12 + 0x18]
  3a2b3b:      	cmp	rsi, rdi
  3a2b3e:      	ja	0x3a22cf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xd9f>
  3a2b44:      	jmp	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a2b49:      	xor	eax, eax
  3a2b4b:      	xor	edx, edx
  3a2b4d:      	test	rax, rax
  3a2b50:      	jne	0x3a2689 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1159>
  3a2b56:      	mov	rdx, qword ptr [r10 + 0x108]
  3a2b5d:      	add	r12d, esi
  3a2b60:      	cmp	rdx, r12
  3a2b63:      	jae	0x3a26a7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1177>
  3a2b69:      	jmp	0x3a3727 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21f7>
  3a2b6e:      	cmp	eax, 0x3
  3a2b71:      	je	0x3a2f30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a00>
  3a2b77:      	cmp	eax, 0x6
  3a2b7a:      	je	0x3a2ed0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19a0>
  3a2b80:      	lea	eax, [r9 - 0x12]
  3a2b84:      	cmp	eax, 0x4
  3a2b87:      	jb	0x3a2f06 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19d6>
  3a2b8d:      	mov	edi, dword ptr [r12 + 0x14]
  3a2b92:      	cmp	rdi, r8
  3a2b95:      	je	0x3a2bee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x16be>
  3a2b97:      	cmp	r13, rdi
  3a2b9a:      	jbe	0x3a3635 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2105>
  3a2ba0:      	shl	rdi, 0x4
  3a2ba4:      	mov	rax, qword ptr [rbp - 0x48]
  3a2ba8:      	mov	r10, qword ptr [rax + rdi]
  3a2bac:      	mov	eax, r9d
  3a2baf:      	and	eax, 0xffff0000
  3a2bb4:      	cmp	eax, 0xfe0000
  3a2bb9:      	jne	0x3a2c00 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x16d0>
  3a2bbb:      	movzx	r9d, r9w
  3a2bbf:      	cmp	r9d, 0x3
  3a2bc3:      	jne	0x3a2dc2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1892>
  3a2bc9:      	test	r11, r11
  3a2bcc:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2bd2:      	mov	rax, qword ptr [r11]
  3a2bd5:      	test	rax, rax
  3a2bd8:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2bde:      	mov	rdi, qword ptr [rax + 0x8]
  3a2be2:      	call	0x3c3b90 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a2be7:      	xor	eax, eax
  3a2be9:      	jmp	0x3a1ad1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5a1>
  3a2bee:      	xor	r10d, r10d
  3a2bf1:      	mov	eax, r9d
  3a2bf4:      	and	eax, 0xffff0000
  3a2bf9:      	cmp	eax, 0xfe0000
  3a2bfe:      	je	0x3a2bbb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x168b>
  3a2c00:      	add	r9d, 0xff03fff8
  3a2c07:      	cmp	r9d, 0xa
  3a2c0b:      	jae	0x3a3773 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2243>
  3a2c11:      	mov	rcx, qword ptr [rbp - 0x38]
  3a2c15:      	mov	rsi, qword ptr [rcx + 0x80]
  3a2c1c:      	mov	edi, dword ptr [r12 + 0x4]
  3a2c21:      	cmp	rsi, rdi
  3a2c24:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a2c2a:      	imul	rax, rdi, 0xe0
  3a2c31:      	add	rax, qword ptr [rcx + 0x78]
  3a2c35:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2c3b:      	movups	xmm0, xmmword ptr [rax + 0x98]
  3a2c42:      	movups	xmm1, xmmword ptr [rax + 0x88]
  3a2c49:      	movups	xmm2, xmmword ptr [rax + 0x78]
  3a2c4d:      	movups	xmm3, xmmword ptr [rax + 0x68]
  3a2c51:      	movups	xmm4, xmmword ptr [rax + 0x58]
  3a2c55:      	movups	xmm5, xmmword ptr [rax + 0x48]
  3a2c59:      	movzx	ecx, byte ptr [rax + 0x40]
  3a2c5d:      	movups	xmm6, xmmword ptr [rax + 0x30]
  3a2c61:      	movups	xmm7, xmmword ptr [rax + 0x20]
  3a2c65:      	movups	xmm8, xmmword ptr [rax + 0x10]
  3a2c6a:      	mov	rsi, qword ptr [rax + 0x8]
  3a2c6e:      	mov	eax, dword ptr [rax]
  3a2c70:      	mov	dword ptr [rbp - 0x208], eax
  3a2c76:      	mov	qword ptr [rbp - 0x200], rsi
  3a2c7d:      	movups	xmmword ptr [rbp - 0x1f8], xmm8
  3a2c85:      	movups	xmmword ptr [rbp - 0x1e8], xmm7
  3a2c8c:      	movups	xmmword ptr [rbp - 0x1d8], xmm6
  3a2c93:      	mov	byte ptr [rbp - 0x1c8], cl
  3a2c99:      	movups	xmmword ptr [rbp - 0x1c0], xmm5
  3a2ca0:      	movups	xmmword ptr [rbp - 0x1b0], xmm4
  3a2ca7:      	movups	xmmword ptr [rbp - 0x1a0], xmm3
  3a2cae:      	movups	xmmword ptr [rbp - 0x190], xmm2
  3a2cb5:      	movups	xmmword ptr [rbp - 0x180], xmm1
  3a2cbc:      	movups	xmmword ptr [rbp - 0x170], xmm0
  3a2cc3:      	mov	qword ptr [rbp - 0xf0], rdx
  3a2cca:      	mov	qword ptr [rbp - 0xe8], 0x0
  3a2cd5:      	mov	qword ptr [rbp - 0xe0], r10
  3a2cdc:      	mov	qword ptr [rbp - 0xd8], 0x0
  3a2ce7:      	mov	ecx, 0x15
  3a2cec:      	mov	rdi, rsp
  3a2cef:      	lea	rsi, [rbp - 0x208]
  3a2cf6:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a2cf9:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a2d00:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a2d08:      	movups	xmm0, xmmword ptr [rbp - 0xf0]
  3a2d0f:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a2d17:      	lea	rdi, [rbp - 0x118]
  3a2d1e:      	mov	rsi, r11
  3a2d21:      	mov	rcx, r15
  3a2d24:      	mov	rdx, qword ptr [rbp - 0x50]
  3a2d28:      	call	0x3a9110 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a2d2d:      	mov	rax, qword ptr [rbp - 0x118]
  3a2d34:      	mov	rcx, qword ptr [rbp - 0x110]
  3a2d3b:      	movzx	r15d, byte ptr [rbp - 0x108]
  3a2d43:      	mov	rsi, qword ptr [rbp - 0x100]
  3a2d4a:      	test	rsi, rsi
  3a2d4d:      	je	0x3a2eac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x197c>
  3a2d53:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a2d58:      	mov	r11, qword ptr [rbp - 0x30]
  3a2d5c:      	mov	r8d, 0xffffffff
  3a2d62:      	mov	r10, qword ptr [rbp - 0x38]
  3a2d66:      	je	0x3a2625 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10f5>
  3a2d6c:      	jmp	0x3a3610 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20e0>
  3a2d71:      	test	rax, rax
  3a2d74:      	je	0x3a2eb1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1981>
  3a2d7a:      	mov	rax, qword ptr [rax]
  3a2d7d:      	test	rdi, rdi
  3a2d80:      	mov	r11, qword ptr [rbp - 0x30]
  3a2d84:      	mov	r10, qword ptr [rbp - 0x38]
  3a2d88:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2d8e:      	mov	rcx, qword ptr [rdi + 0x38]
  3a2d92:      	test	rcx, rcx
  3a2d95:      	mov	edx, 0x10000
  3a2d9a:      	cmove	rcx, rdx
  3a2d9e:      	mov	rdx, rax
  3a2da1:      	or	rdx, rcx
  3a2da4:      	shr	rdx, 0x20
  3a2da8:      	je	0x3a2ea3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1973>
  3a2dae:      	xor	edx, edx
  3a2db0:      	div	rcx
  3a2db3:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2db8:      	mov	eax, 0x20
  3a2dbd:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2dc2:      	mov	rax, qword ptr [rbp - 0x38]
  3a2dc6:      	mov	rsi, qword ptr [rax + 0x80]
  3a2dcd:      	mov	edi, dword ptr [r12 + 0x4]
  3a2dd2:      	cmp	rsi, rdi
  3a2dd5:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a2ddb:      	imul	r11, rdi, 0xe0
  3a2de2:      	add	r11, qword ptr [rax + 0x78]
  3a2de6:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2dec:      	mov	rax, qword ptr [rbp - 0x50]
  3a2df0:      	mov	rsi, qword ptr [rax + 0x78]
  3a2df4:      	mov	edi, dword ptr [r11 + 0x18]
  3a2df8:      	cmp	rsi, rdi
  3a2dfb:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a2e01:      	shl	rdi, 0x3
  3a2e05:      	add	rdi, qword ptr [rax + 0x70]
  3a2e09:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2e0f:      	mov	rsi, qword ptr [rdi]
  3a2e12:      	test	rsi, rsi
  3a2e15:      	je	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2e1b:      	mov	rdi, qword ptr [rbp - 0x30]
  3a2e1f:      	mov	rcx, qword ptr [rdi]
  3a2e22:      	mov	rax, qword ptr [rdi + 0x8]
  3a2e26:      	mov	r8, qword ptr [rcx + 0x8]
  3a2e2a:      	mov	rcx, qword ptr [rdi + 0x10]
  3a2e2e:      	mov	rsi, qword ptr [rsi + 0x10]
  3a2e32:      	mov	edi, dword ptr [r11 + 0x14]
  3a2e36:      	mov	r11d, dword ptr [r11 + 0x1c]
  3a2e3a:      	shl	r11, 0x20
  3a2e3e:      	or	r11, rdi
  3a2e41:      	mov	qword ptr [rsp + 0x18], r10
  3a2e46:      	mov	qword ptr [rsp + 0x10], rdx
  3a2e4b:      	mov	qword ptr [rsp + 0x8], r11
  3a2e50:      	mov	qword ptr [rsp], r15
  3a2e54:      	lea	rdi, [rbp - 0x140]
  3a2e5b:      	mov	rdx, rax
  3a2e5e:      	call	0x3c4e20 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a2e63:      	mov	rax, qword ptr [rbp - 0x140]
  3a2e6a:      	mov	rcx, qword ptr [rbp - 0x138]
  3a2e71:      	movzx	r15d, byte ptr [rbp - 0x130]
  3a2e79:      	mov	rsi, qword ptr [rbp - 0x128]
  3a2e80:      	test	rsi, rsi
  3a2e83:      	je	0x3a2ecb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x199b>
  3a2e85:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a2e8a:      	mov	r11, qword ptr [rbp - 0x30]
  3a2e8e:      	mov	r8d, 0xffffffff
  3a2e94:      	mov	r10, qword ptr [rbp - 0x38]
  3a2e98:      	je	0x3a2625 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10f5>
  3a2e9e:      	jmp	0x3a3619 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20e9>
  3a2ea3:      	xor	edx, edx
  3a2ea5:      	div	ecx
  3a2ea7:      	jmp	0x3a2530 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1000>
  3a2eac:      	jmp	0x3a2617 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10e7>
  3a2eb1:      	mov	rax, qword ptr [rdi + 0x8]
  3a2eb5:      	test	rdi, rdi
  3a2eb8:      	mov	r11, qword ptr [rbp - 0x30]
  3a2ebc:      	mov	r10, qword ptr [rbp - 0x38]
  3a2ec0:      	jne	0x3a2d8e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x185e>
  3a2ec6:      	jmp	0x3a363d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x210d>
  3a2ecb:      	jmp	0x3a2617 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10e7>
  3a2ed0:      	mov	rax, qword ptr [rbp - 0x60]
  3a2ed4:      	mov	dword ptr [rax - 0x38], r14d
  3a2ed8:      	mov	rdi, r11
  3a2edb:      	mov	r14, r11
  3a2ede:      	call	0x3a5f60 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a2ee3:      	test	al, al
  3a2ee5:      	jne	0x3a35fb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20cb>
  3a2eeb:      	mov	rcx, qword ptr [r14 + 0x48]
  3a2eef:      	test	rcx, rcx
  3a2ef2:      	mov	r11, r14
  3a2ef5:      	mov	r8d, 0xffffffff
  3a2efb:      	jg	0x3a1573 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x43>
  3a2f01:      	jmp	0x3a3764 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2234>
  3a2f06:      	mov	rax, qword ptr [rbp - 0x60]
  3a2f0a:      	mov	dword ptr [rax - 0x38], r14d
  3a2f0e:      	mov	rdi, r11
  3a2f11:      	mov	r14, r11
  3a2f14:      	call	0x3a4a50 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a2f19:      	jmp	0x3a2ee3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b3>
  3a2f1b:      	mov	rax, qword ptr [rbp - 0x60]
  3a2f1f:      	mov	dword ptr [rax - 0x38], r14d
  3a2f23:      	mov	rdi, r11
  3a2f26:      	mov	r14, r11
  3a2f29:      	call	0x3a5960 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a2f2e:      	jmp	0x3a2ee3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19b3>
  3a2f30:      	mov	rax, qword ptr [rbp - 0x60]
  3a2f34:      	mov	dword ptr [rax - 0x38], r14d
  3a2f38:      	mov	rdi, r11
  3a2f3b:      	mov	r14, r11
  3a2f3e:      	call	0x3a45f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a2f43:      	test	al, al
  3a2f45:      	je	0x3a2eeb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x19bb>
  3a2f47:      	mov	al, 0x1
  3a2f49:      	jmp	0x3a35fb <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20cb>
  3a2f4e:      	mov	edi, 0x28
  3a2f53:      	call	0x5fa400 <GC_malloc@plt>
  3a2f58:      	test	rax, rax
  3a2f5b:      	je	0x3a3782 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2252>
  3a2f61:      	mov	r14, rax
  3a2f64:      	add	r15, r13
  3a2f67:      	mov	edi, 0x28
  3a2f6c:      	call	0x5e01d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a2f71:      	mov	rax, qword ptr [r15 + 0x8]
  3a2f75:      	movaps	xmm0, xmmword ptr [rip + 0x30fa24] # 0x6b29a0 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a2f7c:      	movups	xmmword ptr [r14], xmm0
  3a2f80:      	mov	qword ptr [r14 + 0x10], rax
  3a2f84:      	mov	qword ptr [r14 + 0x18], 0x29a078
  3a2f8c:      	mov	qword ptr [r14 + 0x20], 0xe
  3a2f94:      	mov	r15, qword ptr [rbp - 0x30]
  3a2f98:      	mov	qword ptr [r15 + 0x140], 0x2baa78
  3a2fa3:      	mov	qword ptr [r15 + 0x148], r14
  3a2faa:      	mov	byte ptr [r15 + 0x138], 0x3
  3a2fb2:      	lea	rdi, [r15 + 0x168]
  3a2fb9:      	mov	edx, 0x158
  3a2fbe:      	xor	esi, esi
  3a2fc0:      	call	0x5fa1d0 <memset@plt>
  3a2fc5:      	mov	rdi, qword ptr [r15 + 0x40]
  3a2fc9:      	mov	rdx, qword ptr [r15 + 0x48]
  3a2fcd:      	shl	rdx, 0x6
  3a2fd1:      	xor	esi, esi
  3a2fd3:      	call	0x5fa1d0 <memset@plt>
  3a2fd8:      	mov	qword ptr [r15 + 0x48], 0x0
  3a2fe0:      	mov	rdi, qword ptr [r15 + 0x58]
  3a2fe4:      	mov	rax, qword ptr [r15 + 0x60]
  3a2fe8:      	shl	rax, 0x4
  3a2fec:      	lea	rdx, [rax + 2*rax]
  3a2ff0:      	xor	esi, esi
  3a2ff2:      	call	0x5fa1d0 <memset@plt>
  3a2ff7:      	mov	qword ptr [r15 + 0x60], 0x0
  3a2fff:      	mov	qword ptr [r15 + 0x70], 0x0
  3a3007:      	mov	qword ptr [r15 + 0x80], 0x0
  3a3012:      	mov	qword ptr [r15 + 0x118], 0x0
  3a301d:      	xorps	xmm0, xmm0
  3a3020:      	movups	xmmword ptr [r15 + 0xd8], xmm0
  3a3028:      	movups	xmmword ptr [r15 + 0xe8], xmm0
  3a3030:      	mov	qword ptr [r15 + 0xf8], 0x0
  3a303b:      	mov	rdi, qword ptr [r15 + 0x100]
  3a3042:      	mov	rdx, qword ptr [r15 + 0x108]
  3a3049:      	shl	rdx, 0x4
  3a304d:      	xor	esi, esi
  3a304f:      	call	0x5fa1d0 <memset@plt>
  3a3054:      	cmp	qword ptr [r15 + 0x110], 0x41
  3a305c:      	jl	0x3a3074 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1b44>
  3a305e:      	mov	qword ptr [r15 + 0x100], 0x0
  3a3069:      	mov	qword ptr [r15 + 0x110], 0x0
  3a3074:      	mov	qword ptr [r15 + 0x108], 0x0
  3a307f:      	xorps	xmm0, xmm0
  3a3082:      	movups	xmmword ptr [r15 + 0x120], xmm0
  3a308a:      	mov	qword ptr [r15 + 0x130], 0x0
  3a3095:      	mov	rdi, qword ptr [r15 + 0xa8]
  3a309c:      	mov	rdx, qword ptr [r15 + 0xb0]
  3a30a3:      	shl	rdx, 0x4
  3a30a7:      	xor	esi, esi
  3a30a9:      	call	0x5fa1d0 <memset@plt>
  3a30ae:      	mov	rax, qword ptr [rbp - 0x30]
  3a30b2:      	mov	qword ptr [rax + 0xb0], 0x0
  3a30bd:      	mov	rax, qword ptr [rbp - 0x30]
  3a30c1:      	mov	rdi, qword ptr [rax + 0xc0]
  3a30c8:      	mov	rax, qword ptr [rbp - 0x30]
  3a30cc:      	mov	rdx, qword ptr [rax + 0xc8]
  3a30d3:      	jmp	0x3a35bf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x208f>
  3a30d8:      	mov	qword ptr [r11 + 0x140], rax
  3a30df:      	mov	qword ptr [r11 + 0x148], rdx
  3a30e6:      	mov	byte ptr [r11 + 0x138], 0x3
  3a30ee:      	lea	rdi, [r11 + 0x168]
  3a30f5:      	mov	edx, 0x158
  3a30fa:      	xor	esi, esi
  3a30fc:      	call	0x5fa1d0 <memset@plt>
  3a3101:      	mov	rax, qword ptr [rbp - 0x30]
  3a3105:      	mov	rdi, qword ptr [rax + 0x40]
  3a3109:      	mov	rax, qword ptr [rbp - 0x30]
  3a310d:      	mov	rdx, qword ptr [rax + 0x48]
  3a3111:      	shl	rdx, 0x6
  3a3115:      	xor	esi, esi
  3a3117:      	call	0x5fa1d0 <memset@plt>
  3a311c:      	mov	rax, qword ptr [rbp - 0x30]
  3a3120:      	mov	qword ptr [rax + 0x48], 0x0
  3a3128:      	mov	rax, qword ptr [rbp - 0x30]
  3a312c:      	mov	rdi, qword ptr [rax + 0x58]
  3a3130:      	mov	rax, qword ptr [rbp - 0x30]
  3a3134:      	mov	rax, qword ptr [rax + 0x60]
  3a3138:      	shl	rax, 0x4
  3a313c:      	lea	rdx, [rax + 2*rax]
  3a3140:      	xor	esi, esi
  3a3142:      	call	0x5fa1d0 <memset@plt>
  3a3147:      	mov	rax, qword ptr [rbp - 0x30]
  3a314b:      	mov	qword ptr [rax + 0x60], 0x0
  3a3153:      	mov	rax, qword ptr [rbp - 0x30]
  3a3157:      	mov	qword ptr [rax + 0x70], 0x0
  3a315f:      	mov	rax, qword ptr [rbp - 0x30]
  3a3163:      	mov	qword ptr [rax + 0x80], 0x0
  3a316e:      	mov	rax, qword ptr [rbp - 0x30]
  3a3172:      	mov	qword ptr [rax + 0x118], 0x0
  3a317d:      	xorps	xmm0, xmm0
  3a3180:      	mov	rax, qword ptr [rbp - 0x30]
  3a3184:      	movups	xmmword ptr [rax + 0xd8], xmm0
  3a318b:      	mov	rax, qword ptr [rbp - 0x30]
  3a318f:      	movups	xmmword ptr [rax + 0xe8], xmm0
  3a3196:      	mov	rax, qword ptr [rbp - 0x30]
  3a319a:      	mov	qword ptr [rax + 0xf8], 0x0
  3a31a5:      	mov	rax, qword ptr [rbp - 0x30]
  3a31a9:      	mov	rdi, qword ptr [rax + 0x100]
  3a31b0:      	mov	rax, qword ptr [rbp - 0x30]
  3a31b4:      	mov	rdx, qword ptr [rax + 0x108]
  3a31bb:      	shl	rdx, 0x4
  3a31bf:      	xor	esi, esi
  3a31c1:      	call	0x5fa1d0 <memset@plt>
  3a31c6:      	mov	rax, qword ptr [rbp - 0x30]
  3a31ca:      	cmp	qword ptr [rax + 0x110], 0x41
  3a31d2:      	jl	0x3a31ea <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1cba>
  3a31d4:      	mov	qword ptr [rax + 0x100], 0x0
  3a31df:      	mov	qword ptr [rax + 0x110], 0x0
  3a31ea:      	mov	qword ptr [rax + 0x108], 0x0
  3a31f5:      	xorps	xmm0, xmm0
  3a31f8:      	movups	xmmword ptr [rax + 0x120], xmm0
  3a31ff:      	mov	qword ptr [rax + 0x130], 0x0
  3a320a:      	mov	rdi, qword ptr [rax + 0xa8]
  3a3211:      	mov	rdx, qword ptr [rax + 0xb0]
  3a3218:      	shl	rdx, 0x4
  3a321c:      	xor	esi, esi
  3a321e:      	call	0x5fa1d0 <memset@plt>
  3a3223:      	mov	rax, qword ptr [rbp - 0x30]
  3a3227:      	mov	qword ptr [rax + 0xb0], 0x0
  3a3232:      	mov	rax, qword ptr [rbp - 0x30]
  3a3236:      	mov	rdi, qword ptr [rax + 0xc0]
  3a323d:      	mov	rax, qword ptr [rbp - 0x30]
  3a3241:      	mov	rdx, qword ptr [rax + 0xc8]
  3a3248:      	jmp	0x3a35bf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x208f>
  3a324d:      	mov	r15d, edx
  3a3250:      	jmp	0x3a341b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eeb>
  3a3255:      	mov	r14, qword ptr [rbp - 0x148]
  3a325c:      	mov	edi, 0x8
  3a3261:      	call	0x5fa400 <GC_malloc@plt>
  3a3266:      	test	rax, rax
  3a3269:      	je	0x3a3782 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2252>
  3a326f:      	mov	r12, rax
  3a3272:      	mov	edi, 0x8
  3a3277:      	call	0x5e01d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a327c:      	mov	edx, 0x299fe8
  3a3281:      	mov	rdi, r15
  3a3284:      	mov	rsi, r14
  3a3287:      	mov	rcx, r12
  3a328a:      	call	0x5b7210 <errors.As>
  3a328f:      	test	al, 0x1
  3a3291:      	je	0x3a32a4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d74>
  3a3293:      	mov	rax, qword ptr [r12]
  3a3297:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a329e:      	mov	ecx, dword ptr [rcx + 0x50]
  3a32a1:      	mov	dword ptr [rax + 0x4], ecx
  3a32a4:      	mov	r12, qword ptr [rbp - 0x30]
  3a32a8:      	mov	qword ptr [r12 + 0x140], r15
  3a32b0:      	mov	qword ptr [r12 + 0x148], r14
  3a32b8:      	mov	byte ptr [r12 + 0x138], 0x3
  3a32c1:      	lea	rdi, [r12 + 0x168]
  3a32c9:      	mov	edx, 0x158
  3a32ce:      	xor	esi, esi
  3a32d0:      	call	0x5fa1d0 <memset@plt>
  3a32d5:      	mov	rdi, qword ptr [r12 + 0x40]
  3a32da:      	mov	rdx, qword ptr [r12 + 0x48]
  3a32df:      	shl	rdx, 0x6
  3a32e3:      	xor	esi, esi
  3a32e5:      	call	0x5fa1d0 <memset@plt>
  3a32ea:      	mov	qword ptr [r12 + 0x48], 0x0
  3a32f3:      	mov	rdi, qword ptr [r12 + 0x58]
  3a32f8:      	mov	rax, qword ptr [r12 + 0x60]
  3a32fd:      	shl	rax, 0x4
  3a3301:      	lea	rdx, [rax + 2*rax]
  3a3305:      	xor	esi, esi
  3a3307:      	call	0x5fa1d0 <memset@plt>
  3a330c:      	mov	qword ptr [r12 + 0x60], 0x0
  3a3315:      	mov	qword ptr [r12 + 0x70], 0x0
  3a331e:      	mov	qword ptr [r12 + 0x80], 0x0
  3a332a:      	mov	qword ptr [r12 + 0x118], 0x0
  3a3336:      	xorps	xmm0, xmm0
  3a3339:      	movups	xmmword ptr [r12 + 0xd8], xmm0
  3a3342:      	movups	xmmword ptr [r12 + 0xe8], xmm0
  3a334b:      	mov	qword ptr [r12 + 0xf8], 0x0
  3a3357:      	mov	rdi, qword ptr [r12 + 0x100]
  3a335f:      	mov	rdx, qword ptr [r12 + 0x108]
  3a3367:      	shl	rdx, 0x4
  3a336b:      	xor	esi, esi
  3a336d:      	call	0x5fa1d0 <memset@plt>
  3a3372:      	cmp	qword ptr [r12 + 0x110], 0x41
  3a337b:      	jl	0x3a3395 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1e65>
  3a337d:      	mov	qword ptr [r12 + 0x100], 0x0
  3a3389:      	mov	qword ptr [r12 + 0x110], 0x0
  3a3395:      	mov	qword ptr [r12 + 0x108], 0x0
  3a33a1:      	xorps	xmm0, xmm0
  3a33a4:      	movups	xmmword ptr [r12 + 0x120], xmm0
  3a33ad:      	mov	qword ptr [r12 + 0x130], 0x0
  3a33b9:      	mov	rdi, qword ptr [r12 + 0xa8]
  3a33c1:      	mov	rdx, qword ptr [r12 + 0xb0]
  3a33c9:      	shl	rdx, 0x4
  3a33cd:      	xor	esi, esi
  3a33cf:      	call	0x5fa1d0 <memset@plt>
  3a33d4:      	mov	rax, qword ptr [rbp - 0x30]
  3a33d8:      	mov	qword ptr [rax + 0xb0], 0x0
  3a33e3:      	mov	rax, qword ptr [rbp - 0x30]
  3a33e7:      	mov	rdi, qword ptr [rax + 0xc0]
  3a33ee:      	mov	rax, qword ptr [rbp - 0x30]
  3a33f2:      	mov	rdx, qword ptr [rax + 0xc8]
  3a33f9:      	jmp	0x3a35bf <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x208f>
  3a33fe:      	mov	r15b, 0x3
  3a3401:      	jmp	0x3a341b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eeb>
  3a3403:      	mov	rdi, r11
  3a3406:      	mov	rsi, rax
  3a3409:      	call	0x3a0ba0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a340e:      	jmp	0x3a35f9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20c9>
  3a3413:      	mov	r15b, 0x1
  3a3416:      	jmp	0x3a341b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1eeb>
  3a3418:      	mov	r15b, 0x9
  3a341b:      	mov	edi, 0x10
  3a3420:      	call	0x5fa4b0 <GC_malloc_atomic@plt>
  3a3425:      	test	rax, rax
  3a3428:      	je	0x3a3782 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2252>
  3a342e:      	mov	r14, rax
  3a3431:      	mov	edi, 0x10
  3a3436:      	call	0x5e01d0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a343b:      	xorps	xmm0, xmm0
  3a343e:      	movups	xmmword ptr [r14], xmm0
  3a3442:      	mov	rax, qword ptr [rbp - 0x38]
  3a3446:      	mov	rsi, qword ptr [rax + 0x80]
  3a344d:      	mov	edi, dword ptr [r12 + 0x4]
  3a3452:      	cmp	rsi, rdi
  3a3455:      	jbe	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2117>
  3a345b:      	mov	rax, qword ptr [rax + 0x78]
  3a345f:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a3466:      	mov	ecx, dword ptr [rcx + 0x50]
  3a3469:      	imul	rdx, rdi, 0xe0
  3a3470:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a3475:      	mov	byte ptr [r14], r15b
  3a3478:      	mov	dword ptr [r14 + 0x4], ecx
  3a347c:      	mov	qword ptr [r14 + 0x8], rax
  3a3480:      	mov	r15, qword ptr [rbp - 0x30]
  3a3484:      	mov	qword ptr [r15 + 0x140], 0x29a400
  3a348f:      	mov	qword ptr [r15 + 0x148], r14
  3a3496:      	mov	byte ptr [r15 + 0x138], 0x3
  3a349e:      	lea	rdi, [r15 + 0x168]
  3a34a5:      	mov	edx, 0x158
  3a34aa:      	xor	esi, esi
  3a34ac:      	call	0x5fa1d0 <memset@plt>
  3a34b1:      	mov	rdi, qword ptr [r15 + 0x40]
  3a34b5:      	mov	rdx, qword ptr [r15 + 0x48]
  3a34b9:      	shl	rdx, 0x6
  3a34bd:      	xor	esi, esi
  3a34bf:      	call	0x5fa1d0 <memset@plt>
  3a34c4:      	mov	qword ptr [r15 + 0x48], 0x0
  3a34cc:      	mov	rdi, qword ptr [r15 + 0x58]
  3a34d0:      	mov	rax, qword ptr [r15 + 0x60]
  3a34d4:      	shl	rax, 0x4
  3a34d8:      	lea	rdx, [rax + 2*rax]
  3a34dc:      	xor	esi, esi
  3a34de:      	call	0x5fa1d0 <memset@plt>
  3a34e3:      	mov	qword ptr [r15 + 0x60], 0x0
  3a34eb:      	mov	qword ptr [r15 + 0x70], 0x0
  3a34f3:      	mov	qword ptr [r15 + 0x80], 0x0
  3a34fe:      	mov	qword ptr [r15 + 0x118], 0x0
  3a3509:      	xorps	xmm0, xmm0
  3a350c:      	movups	xmmword ptr [r15 + 0xd8], xmm0
  3a3514:      	movups	xmmword ptr [r15 + 0xe8], xmm0
  3a351c:      	mov	qword ptr [r15 + 0xf8], 0x0
  3a3527:      	mov	rdi, qword ptr [r15 + 0x100]
  3a352e:      	mov	rdx, qword ptr [r15 + 0x108]
  3a3535:      	shl	rdx, 0x4
  3a3539:      	xor	esi, esi
  3a353b:      	call	0x5fa1d0 <memset@plt>
  3a3540:      	cmp	qword ptr [r15 + 0x110], 0x41
  3a3548:      	jl	0x3a3560 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2030>
  3a354a:      	mov	qword ptr [r15 + 0x100], 0x0
  3a3555:      	mov	qword ptr [r15 + 0x110], 0x0
  3a3560:      	mov	qword ptr [r15 + 0x108], 0x0
  3a356b:      	xorps	xmm0, xmm0
  3a356e:      	movups	xmmword ptr [r15 + 0x120], xmm0
  3a3576:      	mov	qword ptr [r15 + 0x130], 0x0
  3a3581:      	mov	rdi, qword ptr [r15 + 0xa8]
  3a3588:      	mov	rdx, qword ptr [r15 + 0xb0]
  3a358f:      	shl	rdx, 0x4
  3a3593:      	xor	esi, esi
  3a3595:      	call	0x5fa1d0 <memset@plt>
  3a359a:      	mov	rax, qword ptr [rbp - 0x30]
  3a359e:      	mov	qword ptr [rax + 0xb0], 0x0
  3a35a9:      	mov	rax, qword ptr [rbp - 0x30]
  3a35ad:      	mov	rdi, qword ptr [rax + 0xc0]
  3a35b4:      	mov	rax, qword ptr [rbp - 0x30]
  3a35b8:      	mov	rdx, qword ptr [rax + 0xc8]
  3a35bf:      	shl	rdx, 0x4
  3a35c3:      	xor	esi, esi
  3a35c5:      	call	0x5fa1d0 <memset@plt>
  3a35ca:      	mov	rax, qword ptr [rbp - 0x30]
  3a35ce:      	cmp	qword ptr [rax + 0xd0], 0x41
  3a35d6:      	jl	0x3a35ee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20be>
  3a35d8:      	mov	qword ptr [rax + 0xc0], 0x0
  3a35e3:      	mov	qword ptr [rax + 0xd0], 0x0
  3a35ee:      	mov	qword ptr [rax + 0xc8], 0x0
  3a35f9:      	mov	al, 0x3
  3a35fb:      	mov	rdx, rbx
  3a35fe:      	add	rsp, 0x2a8
  3a3605:      	pop	rbx
  3a3606:      	pop	r12
  3a3608:      	pop	r13
  3a360a:      	pop	r14
  3a360c:      	pop	r15
  3a360e:      	pop	rbp
  3a360f:      	ret
  3a3610:      	mov	rdx, qword ptr [rbp - 0xf8]
  3a3617:      	jmp	0x3a3620 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20f0>
  3a3619:      	mov	rdx, qword ptr [rbp - 0x120]
  3a3620:      	mov	rdi, r11
  3a3623:      	call	0x3a0ba0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a3628:      	jmp	0x3a35f9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20c9>
  3a362a:      	mov	rdi, r14
  3a362d:      	mov	rsi, r15
  3a3630:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3635:      	mov	rsi, r13
  3a3638:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a363d:      	mov	edi, 0x1
  3a3642:      	call	0x5e4540 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a3647:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a364c:      	mov	rdi, r12
  3a364f:      	mov	rsi, r15
  3a3652:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3657:      	mov	rdi, rax
  3a365a:      	mov	rsi, r13
  3a365d:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3662:      	mov	rdi, r12
  3a3665:      	mov	rsi, r12
  3a3668:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a366d:      	mov	rdi, rcx
  3a3670:      	mov	rsi, r13
  3a3673:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3678:      	mov	rsi, r13
  3a367b:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3680:      	mov	rdi, r9
  3a3683:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3688:      	mov	rdi, rax
  3a368b:      	mov	rsi, r13
  3a368e:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3693:      	mov	rdi, rcx
  3a3696:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a369b:      	mov	rdi, rsi
  3a369e:      	mov	rsi, r13
  3a36a1:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a36a6:      	mov	edi, 0x1
  3a36ab:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a36b0:      	xor	edi, edi
  3a36b2:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a36b7:      	mov	rdx, rsi
  3a36ba:      	xor	esi, esi
  3a36bc:      	mov	ecx, 0x3
  3a36c1:      	call	0x5eb140 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a36c6:      	mov	rdi, rsi
  3a36c9:      	xor	esi, esi
  3a36cb:      	mov	ecx, 0x2
  3a36d0:      	call	0x5eb140 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a36d5:      	mov	rdi, rax
  3a36d8:      	mov	esi, 0x1
  3a36dd:      	mov	ecx, 0x2
  3a36e2:      	call	0x5eb140 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a36e7:      	mov	rdi, r9
  3a36ea:      	mov	esi, 0x1
  3a36ef:      	mov	rdx, rax
  3a36f2:      	mov	ecx, 0x3
  3a36f7:      	call	0x5eb140 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a36fc:      	mov	edi, 0x3
  3a3701:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3706:      	mov	edi, 0x2
  3a370b:      	mov	esi, 0x2
  3a3710:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3715:      	mov	rdi, rsi
  3a3718:      	xor	esi, esi
  3a371a:      	mov	rdx, r12
  3a371d:      	mov	ecx, 0x3
  3a3722:      	call	0x5eb140 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a3727:      	mov	rdi, r12
  3a372a:      	xor	esi, esi
  3a372c:      	mov	ecx, 0x2
  3a3731:      	call	0x5eb140 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a3736:      	mov	rdi, rdx
  3a3739:      	mov	rsi, r13
  3a373c:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3741:      	mov	rdi, rcx
  3a3744:      	mov	rsi, qword ptr [rbp - 0x78]
  3a3748:      	call	0x5e02b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a374d:      	mov	rdi, rax
  3a3750:      	call	0x5e17e0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3755:      	mov	edi, 0x329520
  3a375a:      	mov	esi, 0x329850
  3a375f:      	call	0x5dfef0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3764:      	mov	edi, 0x3339d8
  3a3769:      	mov	esi, 0x29a068
  3a376e:      	call	0x5dfef0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3773:      	mov	edi, 0x3339d8
  3a3778:      	mov	esi, 0x29a088
  3a377d:      	call	0x5dfef0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3782:      	mov	edi, 0x3339d8
  3a3787:      	mov	esi, 0x3294f8
  3a378c:      	call	0x5dfef0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3791:      	int3
  3a3792:      	int3
  3a3793:      	int3
  3a3794:      	int3
  3a3795:      	int3
  3a3796:      	int3
  3a3797:      	int3
  3a3798:      	int3
  3a3799:      	int3
  3a379a:      	int3
  3a379b:      	int3
  3a379c:      	int3
  3a379d:      	int3
  3a379e:      	int3
  3a379f:      	int3
