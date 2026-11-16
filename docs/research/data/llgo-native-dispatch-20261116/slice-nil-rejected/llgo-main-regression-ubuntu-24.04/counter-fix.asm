
/home/runner/work/wasmvm/wasmvm/main-regression/counter-fix:	file format elf64-x86-64

Disassembly of section .text:

00000000003a1540 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop>:
  3a1540:      	push	rbp
  3a1541:      	mov	rbp, rsp
  3a1544:      	push	r15
  3a1546:      	push	r14
  3a1548:      	push	r13
  3a154a:      	push	r12
  3a154c:      	push	rbx
  3a154d:      	sub	rsp, 0x2a8
  3a1554:      	mov	rcx, qword ptr [rdi + 0x48]
  3a1558:      	test	rcx, rcx
  3a155b:      	jle	0x3a3638 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20f8>
  3a1561:      	mov	rbx, rsi
  3a1564:      	mov	r11, rdi
  3a1567:      	movzx	r12d, byte ptr [rdi + 0x20]
  3a156c:      	mov	r8d, 0xffffffff
  3a1572:      	lea	rax, [rbp - 0x90]
  3a1579:      	mov	qword ptr [rbp - 0x78], rax
  3a157d:      	mov	qword ptr [rbp - 0x30], rdi
  3a1581:      	mov	byte ptr [rbp - 0x31], r12b
  3a1585:      	mov	r10, qword ptr [r11 + 0x40]
  3a1589:      	shl	rcx, 0x6
  3a158d:      	mov	byte ptr [r10 + rcx - 0x30], 0x0
  3a1593:      	mov	rsi, qword ptr [r10 + rcx - 0x40]
  3a1598:      	mov	rax, qword ptr [rsi + 0x48]
  3a159c:      	mov	qword ptr [rbp - 0x50], rax
  3a15a0:      	mov	r9, qword ptr [rax + 0x98]
  3a15a7:      	mov	rdi, qword ptr [r10 + rcx - 0x28]
  3a15ac:      	mov	rdx, qword ptr [r11 + 0x88]
  3a15b3:      	lea	rax, [r9 + rdi]
  3a15b7:      	cmp	rax, rdx
  3a15ba:      	ja	0x3a35b6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2076>
  3a15c0:      	cmp	rdi, rax
  3a15c3:      	ja	0x3a35c8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2088>
  3a15c9:      	mov	r14, qword ptr [rbp - 0x50]
  3a15cd:      	mov	r13, qword ptr [r14 + 0xb8]
  3a15d4:      	mov	r15d, dword ptr [r10 + rcx - 0x38]
  3a15d9:      	sub	rdx, rdi
  3a15dc:      	test	rdx, rdx
  3a15df:      	mov	eax, 0x0
  3a15e4:      	cmovle	rdi, rax
  3a15e8:      	mov	qword ptr [rbp - 0x60], r13
  3a15ec:      	cmp	r13, r15
  3a15ef:      	jbe	0x3a35da <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x209a>
  3a15f5:      	add	r10, rcx
  3a15f8:      	mov	qword ptr [rbp - 0x68], r10
  3a15fc:      	mov	qword ptr [rbp - 0xa8], rsi
  3a1603:      	mov	rax, qword ptr [rsi + 0x40]
  3a1607:      	mov	qword ptr [rbp - 0x58], rax
  3a160b:      	mov	r13, qword ptr [r14 + 0xb0]
  3a1612:      	shl	rdi, 0x4
  3a1616:      	add	rdi, qword ptr [r11 + 0x78]
  3a161a:      	mov	r14d, r15d
  3a161d:      	mov	qword ptr [rbp - 0x48], r9
  3a1621:      	mov	rsi, rdi
  3a1624:      	mov	qword ptr [rbp - 0x40], rdi
  3a1628:      	mov	r10, qword ptr [rbp - 0x60]
  3a162c:      	mov	qword ptr [rbp - 0xb0], r13
  3a1633:      	jmp	0x3a1669 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x129>
  3a1635:      	nop	word ptr cs:[rax + rax]
  3a1640:      	shl	rdi, 0x4
  3a1644:      	mov	qword ptr [rsi + rdi], rax
  3a1648:      	mov	qword ptr [rsi + rdi + 0x8], rdx
  3a164d:      	inc	r14d
  3a1650:      	mov	r13d, r14d
  3a1653:      	mov	r15d, r13d
  3a1656:      	mov	r14d, r13d
  3a1659:      	cmp	r10, r15
  3a165c:      	mov	r13, qword ptr [rbp - 0xb0]
  3a1663:      	jbe	0x3a35de <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x209e>
  3a1669:      	shl	r15, 0x5
  3a166d:      	add	r15, r13
  3a1670:      	test	r12b, 0x1
  3a1674:      	je	0x3a16d6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x196>
  3a1676:      	mov	rax, qword ptr [rbp - 0x50]
  3a167a:      	mov	rsi, qword ptr [rax + 0x80]
  3a1681:      	mov	edi, dword ptr [r15 + 0x4]
  3a1685:      	cmp	rsi, rdi
  3a1688:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a168e:      	mov	r13, qword ptr [rax + 0x78]
  3a1692:      	imul	r12, rdi, 0xe0
  3a1699:      	mov	eax, dword ptr [r13 + r12]
  3a169e:      	cmp	eax, 0x5
  3a16a1:      	setne	cl
  3a16a4:      	cmp	eax, 0xb
  3a16a7:      	setne	al
  3a16aa:      	and	al, cl
  3a16ac:      	movzx	ecx, al
  3a16af:      	mov	rax, qword ptr [r11 + 0x150]
  3a16b6:      	sub	rax, rcx
  3a16b9:      	jb	0x3a2e57 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1917>
  3a16bf:      	mov	qword ptr [r11 + 0x150], rax
  3a16c6:      	movzx	r12d, byte ptr [rbp - 0x31]
  3a16cb:      	mov	rsi, qword ptr [rbp - 0x40]
  3a16cf:      	mov	r13, qword ptr [rbp - 0xb0]
  3a16d6:      	inc	rbx
  3a16d9:      	test	bl, 0x3f
  3a16dc:      	jne	0x3a175c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x21c>
  3a16de:      	mov	rax, qword ptr [r11 + 0x160]
  3a16e5:      	test	rax, rax
  3a16e8:      	je	0x3a16fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bd>
  3a16ea:      	test	r11, r11
  3a16ed:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a16f3:      	mov	eax, dword ptr [rax]
  3a16f5:      	test	eax, eax
  3a16f7:      	je	0x3a17f0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2b0>
  3a16fd:      	mov	rax, qword ptr [r11 + 0x8]
  3a1701:      	test	rax, rax
  3a1704:      	je	0x3a3629 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20e9>
  3a170a:      	mov	rdi, qword ptr [r11 + 0x10]
  3a170e:      	mov	rcx, qword ptr [rax + 0x8]
  3a1712:      	test	byte ptr [rcx + 0x17], 0x20
  3a1716:      	je	0x3a172f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ef>
  3a1718:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a171c:      	and	cl, 0x1f
  3a171f:      	cmp	cl, 0x16
  3a1722:      	je	0x3a172f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ef>
  3a1724:      	mov	qword ptr [rbp - 0x90], rdi
  3a172b:      	mov	rdi, qword ptr [rbp - 0x78]
  3a172f:      	call	qword ptr [rax + 0x28]
  3a1732:      	mov	r11, qword ptr [rbp - 0x30]
  3a1736:      	mov	r8d, 0xffffffff
  3a173c:      	mov	r9, qword ptr [rbp - 0x48]
  3a1740:      	mov	rsi, qword ptr [rbp - 0x40]
  3a1744:      	mov	r10, qword ptr [rbp - 0x60]
  3a1748:      	test	rax, rax
  3a174b:      	je	0x3a17fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2bd>
  3a1751:      	cmp	qword ptr [rax + 0x8], 0x0
  3a1756:      	jne	0x3a2fe1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1aa1>
  3a175c:      	cmp	dword ptr [r15 + 0x1c], 0x0
  3a1761:      	je	0x3a1810 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d0>
  3a1767:      	test	r12b, 0x1
  3a176b:      	je	0x3a1790 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x250>
  3a176d:      	mov	edx, dword ptr [r15]
  3a1770:      	cmp	edx, 0x20
  3a1773:      	je	0x3a1810 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d0>
  3a1779:      	cmp	edx, 0x41
  3a177c:      	je	0x3a1810 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d0>
  3a1782:      	jmp	0x3a1813 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d3>
  3a1787:      	nop	word ptr [rax + rax]
  3a1790:      	mov	edx, dword ptr [r15]
  3a1793:      	cmp	edx, 0x20
  3a1796:      	je	0x3a179d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x25d>
  3a1798:      	cmp	edx, 0x41
  3a179b:      	jne	0x3a1813 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d3>
  3a179d:      	mov	eax, ebx
  3a179f:      	not	eax
  3a17a1:      	test	al, 0x3f
  3a17a3:      	je	0x3a1810 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d0>
  3a17a5:      	cmp	edx, 0x41
  3a17a8:      	jne	0x3a17cc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x28c>
  3a17aa:      	mov	edi, dword ptr [r15 + 0x8]
  3a17ae:      	cmp	r9, rdi
  3a17b1:      	jbe	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a17b7:      	mov	eax, dword ptr [r15 + 0x18]
  3a17bb:      	shl	rdi, 0x4
  3a17bf:      	mov	qword ptr [rsi + rdi], rax
  3a17c3:      	mov	qword ptr [rsi + rdi + 0x8], 0x0
  3a17cc:      	inc	r14d
  3a17cf:      	cmp	r10, r14
  3a17d2:      	jbe	0x3a351f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fdf>
  3a17d8:      	mov	rax, r14
  3a17db:      	shl	rax, 0x5
  3a17df:      	lea	r15, [rax + r13]
  3a17e3:      	inc	rbx
  3a17e6:      	cmp	dword ptr [r13 + rax + 0x1c], 0x0
  3a17ec:      	jne	0x3a1790 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x250>
  3a17ee:      	jmp	0x3a1810 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2d0>
  3a17f0:      	xor	eax, eax
  3a17f2:      	xor	edx, edx
  3a17f4:      	test	rax, rax
  3a17f7:      	jne	0x3a1751 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x211>
  3a17fd:      	cmp	dword ptr [r15 + 0x1c], 0x0
  3a1802:      	jne	0x3a1767 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x227>
  3a1808:      	nop	dword ptr [rax + rax]
  3a1810:      	mov	edx, dword ptr [r15]
  3a1813:      	lea	eax, [rdx - 0xe0]
  3a1819:      	cmp	eax, 0xffff20
  3a181e:      	jae	0x3a1aa0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x560>
  3a1824:      	mov	eax, edx
  3a1826:      	shr	eax, 0x10
  3a1829:      	cmp	eax, 0xfb
  3a182e:      	je	0x3a1a07 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x4c7>
  3a1834:      	cmp	eax, 0xfc
  3a1839:      	je	0x3a18de <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x39e>
  3a183f:      	cmp	eax, 0xfd
  3a1844:      	jne	0x3a1a89 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x549>
  3a184a:      	mov	edi, dword ptr [r15 + 0xc]
  3a184e:      	cmp	rdi, r8
  3a1851:      	je	0x3a1d1e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7de>
  3a1857:      	cmp	r9, rdi
  3a185a:      	mov	r12, qword ptr [rbp - 0x50]
  3a185e:      	jbe	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a1864:      	shl	rdi, 0x4
  3a1868:      	mov	rcx, qword ptr [rsi + rdi]
  3a186c:      	mov	r10, qword ptr [rsi + rdi + 0x8]
  3a1871:      	mov	edi, dword ptr [r15 + 0x10]
  3a1875:      	mov	esi, 0xffffffff
  3a187a:      	cmp	rdi, r8
  3a187d:      	je	0x3a1d39 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7f9>
  3a1883:      	cmp	r9, rdi
  3a1886:      	jbe	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a188c:      	shl	rdi, 0x4
  3a1890:      	mov	r8, qword ptr [rbp - 0x40]
  3a1894:      	mov	rax, qword ptr [r8 + rdi]
  3a1898:      	mov	r8, qword ptr [r8 + rdi + 0x8]
  3a189d:      	mov	edi, dword ptr [r15 + 0x14]
  3a18a1:      	cmp	rdi, rsi
  3a18a4:      	je	0x3a1d4b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x80b>
  3a18aa:      	cmp	r9, rdi
  3a18ad:      	jbe	0x3a354f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x200f>
  3a18b3:      	shl	rdi, 0x4
  3a18b7:      	mov	rsi, qword ptr [rbp - 0x40]
  3a18bb:      	mov	r9, qword ptr [rsi + rdi]
  3a18bf:      	mov	r11, qword ptr [rsi + rdi + 0x8]
  3a18c4:      	mov	rsi, qword ptr [r12 + 0x80]
  3a18cc:      	mov	edi, dword ptr [r15 + 0x4]
  3a18d0:      	cmp	rsi, rdi
  3a18d3:      	ja	0x3a1d66 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x826>
  3a18d9:      	jmp	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a18de:      	add	edx, 0xff03ffed
  3a18e4:      	cmp	edx, 0x4
  3a18e7:      	jae	0x3a1aa0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x560>
  3a18ed:      	mov	rax, qword ptr [rbp - 0x50]
  3a18f1:      	mov	rdx, qword ptr [rax + 0xd8]
  3a18f8:      	mov	edi, dword ptr [r15 + 0x18]
  3a18fc:      	mov	eax, dword ptr [r15 + 0x1c]
  3a1900:      	lea	esi, [rdi + rax]
  3a1903:      	add	esi, 0x2
  3a1906:      	cmp	rdx, rsi
  3a1909:      	jb	0x3a35a7 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2067>
  3a190f:      	cmp	edi, esi
  3a1911:      	ja	0x3a3598 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2058>
  3a1917:      	mov	eax, edi
  3a1919:      	mov	ecx, esi
  3a191b:      	sub	rsi, rdi
  3a191e:      	sub	rdx, rdi
  3a1921:      	test	rdx, rdx
  3a1924:      	mov	edx, 0x0
  3a1929:      	cmovle	rdi, rdx
  3a192d:      	cmp	ecx, eax
  3a192f:      	je	0x3a3591 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2051>
  3a1935:      	mov	rax, qword ptr [rbp - 0x50]
  3a1939:      	mov	rax, qword ptr [rax + 0xc8]
  3a1940:      	lea	r8, [rax + 4*rdi]
  3a1944:      	mov	edi, dword ptr [r8]
  3a1947:      	cmp	r9, rdi
  3a194a:      	jbe	0x3a3558 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2018>
  3a1950:      	cmp	rsi, 0x1
  3a1954:      	jbe	0x3a3587 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2047>
  3a195a:      	mov	eax, dword ptr [r8 + 0x4]
  3a195e:      	cmp	r9, rax
  3a1961:      	jbe	0x3a3568 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2028>
  3a1967:      	shl	rdi, 0x4
  3a196b:      	mov	rcx, qword ptr [rbp - 0x40]
  3a196f:      	mov	r9, qword ptr [rcx + rdi]
  3a1973:      	shl	rax, 0x4
  3a1977:      	mov	rcx, qword ptr [rcx + rax]
  3a197b:      	mov	r10d, dword ptr [r15]
  3a197e:      	lea	eax, [r10 - 0xfc0013]
  3a1985:      	cmp	eax, 0x2
  3a1988:      	jae	0x3a1f25 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9e5>
  3a198e:      	cmp	rsi, 0x2
  3a1992:      	je	0x3a35fc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20bc>
  3a1998:      	mov	edi, dword ptr [r8 + 0x8]
  3a199c:      	cmp	qword ptr [rbp - 0x48], rdi
  3a19a0:      	jbe	0x3a35e9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20a9>
  3a19a6:      	cmp	rsi, 0x3
  3a19aa:      	jbe	0x3a35f2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20b2>
  3a19b0:      	mov	eax, dword ptr [r8 + 0xc]
  3a19b4:      	cmp	qword ptr [rbp - 0x48], rax
  3a19b8:      	jbe	0x3a3568 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2028>
  3a19be:      	shl	rdi, 0x4
  3a19c2:      	mov	rdx, qword ptr [rbp - 0x40]
  3a19c6:      	mov	rdi, qword ptr [rdx + rdi]
  3a19ca:      	shl	rax, 0x4
  3a19ce:      	mov	rdx, qword ptr [rdx + rax]
  3a19d2:      	cmp	r10d, 0xfc0013
  3a19d9:      	jne	0x3a1fbd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa7d>
  3a19df:      	lea	rax, [rdi + r9]
  3a19e3:      	mov	r10, rdi
  3a19e6:      	and	r10, r9
  3a19e9:      	or	rdi, r9
  3a19ec:      	mov	r9, rax
  3a19ef:      	not	r9
  3a19f2:      	and	r9, rdi
  3a19f5:      	or	r9, r10
  3a19f8:      	shr	r9, 0x3f
  3a19fc:      	add	rdx, rcx
  3a19ff:      	add	rdx, r9
  3a1a02:      	jmp	0x3a1fe2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa2>
  3a1a07:      	add	edx, -0x2
  3a1a0a:      	cmp	dx, 0x21
  3a1a0e:      	jae	0x3a1c0c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6cc>
  3a1a14:      	lea	rax, [r8 + 0x1c003e10]
  3a1a1b:      	movzx	ecx, dx
  3a1a1e:      	bt	rax, rcx
  3a1a22:      	jae	0x3a1c0c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6cc>
  3a1a28:      	mov	edi, dword ptr [r15 + 0xc]
  3a1a2c:      	mov	eax, 0xffffffff
  3a1a31:      	cmp	rdi, r8
  3a1a34:      	je	0x3a2086 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb46>
  3a1a3a:      	cmp	r9, rdi
  3a1a3d:      	jbe	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a1a43:      	shl	rdi, 0x4
  3a1a47:      	mov	r8, qword ptr [rsi + rdi]
  3a1a4b:      	mov	edi, dword ptr [r15 + 0x10]
  3a1a4f:      	cmp	rdi, rax
  3a1a52:      	je	0x3a2096 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb56>
  3a1a58:      	cmp	r9, rdi
  3a1a5b:      	jbe	0x3a354f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x200f>
  3a1a61:      	shl	rdi, 0x4
  3a1a65:      	mov	rax, qword ptr [rsi + rdi]
  3a1a69:      	mov	r9, qword ptr [rsi + rdi + 0x8]
  3a1a6e:      	mov	r10d, 0x0
  3a1a74:      	mov	r11d, 0x0
  3a1a7a:      	cmp	dx, 0xc
  3a1a7e:      	je	0x3a20ad <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb6d>
  3a1a84:      	jmp	0x3a20ed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbad>
  3a1a89:      	cmp	edx, 0xe7
  3a1a8f:      	jb	0x3a2dd9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a1a95:      	nop	word ptr cs:[rax + rax]
  3a1aa0:      	mov	edi, dword ptr [r15 + 0xc]
  3a1aa4:      	cmp	rdi, r8
  3a1aa7:      	je	0x3a1b30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5f0>
  3a1aad:      	cmp	r9, rdi
  3a1ab0:      	jbe	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a1ab6:      	shl	rdi, 0x4
  3a1aba:      	mov	r13, qword ptr [rsi + rdi]
  3a1abe:      	mov	rdx, qword ptr [rsi + rdi + 0x8]
  3a1ac3:      	mov	edi, dword ptr [r15 + 0x10]
  3a1ac7:      	cmp	rdi, r8
  3a1aca:      	je	0x3a1b3e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5fe>
  3a1acc:      	cmp	r9, rdi
  3a1acf:      	jbe	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a1ad5:      	shl	rdi, 0x4
  3a1ad9:      	mov	r10, qword ptr [rsi + rdi]
  3a1add:      	mov	rax, qword ptr [rsi + rdi + 0x8]
  3a1ae2:      	mov	r9d, dword ptr [r15]
  3a1ae5:      	lea	ecx, [r9 - 0x45]
  3a1ae9:      	cmp	ecx, 0x7f
  3a1aec:      	ja	0x3a1b4f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x60f>
  3a1aee:      	ja	0x3a28f1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13b1>
  3a1af4:      	jmp	qword ptr [8*rcx + 0x203938]
  3a1afb:      	mov	edi, r9d
  3a1afe:      	mov	rsi, r13
  3a1b01:      	mov	rdx, r10
  3a1b04:      	call	0x3d01c0 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a1b09:      	test	dl, dl
  3a1b0b:      	jne	0x3a3156 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c16>
  3a1b11:      	xor	edx, edx
  3a1b13:      	mov	r11, qword ptr [rbp - 0x30]
  3a1b17:      	mov	r8d, 0xffffffff
  3a1b1d:      	mov	rsi, qword ptr [rbp - 0x40]
  3a1b21:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a1b26:      	nop	word ptr cs:[rax + rax]
  3a1b30:      	xor	r13d, r13d
  3a1b33:      	xor	edx, edx
  3a1b35:      	mov	edi, dword ptr [r15 + 0x10]
  3a1b39:      	cmp	rdi, r8
  3a1b3c:      	jne	0x3a1acc <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x58c>
  3a1b3e:      	xor	r10d, r10d
  3a1b41:      	xor	eax, eax
  3a1b43:      	mov	r9d, dword ptr [r15]
  3a1b46:      	lea	ecx, [r9 - 0x45]
  3a1b4a:      	cmp	ecx, 0x7f
  3a1b4d:      	jbe	0x3a1aee <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5ae>
  3a1b4f:      	lea	ecx, [r9 - 0x1]
  3a1b53:      	cmp	ecx, 0xd3
  3a1b59:      	ja	0x3a1c46 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x706>
  3a1b5f:      	jmp	qword ptr [8*rcx + 0x203258]
  3a1b66:      	mov	rcx, qword ptr [rbp - 0x50]
  3a1b6a:      	mov	rsi, qword ptr [rcx + 0x80]
  3a1b71:      	mov	edi, dword ptr [r15 + 0x4]
  3a1b75:      	cmp	rsi, rdi
  3a1b78:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a1b7e:      	imul	rax, rdi, 0xe0
  3a1b85:      	add	rax, qword ptr [rcx + 0x78]
  3a1b89:      	mov	r10, qword ptr [rbp - 0x78]
  3a1b8d:      	mov	rcx, qword ptr [rbp - 0x58]
  3a1b91:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1b97:      	mov	rsi, qword ptr [rcx + 0x78]
  3a1b9b:      	mov	edi, dword ptr [rax + 0x18]
  3a1b9e:      	cmp	rsi, rdi
  3a1ba1:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a1ba7:      	shl	rdi, 0x3
  3a1bab:      	add	rdi, qword ptr [rcx + 0x70]
  3a1baf:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1bb5:      	mov	rcx, qword ptr [rdi]
  3a1bb8:      	test	rcx, rcx
  3a1bbb:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1bc1:      	mov	rsi, qword ptr [rcx + 0x10]
  3a1bc5:      	mov	ecx, dword ptr [rax + 0x14]
  3a1bc8:      	mov	r8d, dword ptr [rax + 0x1c]
  3a1bcc:      	shl	r8, 0x20
  3a1bd0:      	or	r8, rcx
  3a1bd3:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a1bd8:      	je	0x3a1be5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6a5>
  3a1bda:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a1bdf:      	je	0x3a1f7c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa3c>
  3a1be5:      	xorps	xmm0, xmm0
  3a1be8:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a1bef:      	movups	xmmword ptr [rsp], xmm0
  3a1bf3:      	mov	rdi, r10
  3a1bf6:      	mov	edx, r9d
  3a1bf9:      	mov	rcx, r13
  3a1bfc:      	xor	r9d, r9d
  3a1bff:      	mov	r13, r10
  3a1c02:      	call	0x3c67c0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a1c07:      	jmp	0x3a1f9e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa5e>
  3a1c0c:      	mov	r15, qword ptr [rbp - 0x68]
  3a1c10:      	mov	dword ptr [r15 - 0x38], r14d
  3a1c14:      	mov	rdi, r11
  3a1c17:      	call	0x3a38a0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runGC>
  3a1c1c:      	mov	r10, qword ptr [rbp - 0x60]
  3a1c20:      	mov	rsi, qword ptr [rbp - 0x40]
  3a1c24:      	mov	r9, qword ptr [rbp - 0x48]
  3a1c28:      	mov	r8d, 0xffffffff
  3a1c2e:      	mov	r11, qword ptr [rbp - 0x30]
  3a1c32:      	test	al, al
  3a1c34:      	jne	0x3a34e8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa8>
  3a1c3a:      	inc	r14d
  3a1c3d:      	mov	dword ptr [r15 - 0x38], r14d
  3a1c41:      	jmp	0x3a1650 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x110>
  3a1c46:      	lea	ecx, [r9 - 0xfc0000]
  3a1c4d:      	cmp	ecx, 0x8
  3a1c50:      	jae	0x3a1e3a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8fa>
  3a1c56:      	mov	edi, r9d
  3a1c59:      	mov	rsi, r13
  3a1c5c:      	mov	rdx, r10
  3a1c5f:      	call	0x3d01c0 <github.com/cpunion/wasmvm/internal/engine/numeric.Eval>
  3a1c64:      	mov	r8d, 0xffffffff
  3a1c6a:      	mov	r11, qword ptr [rbp - 0x30]
  3a1c6e:      	mov	r13d, edx
  3a1c71:      	xor	edx, edx
  3a1c73:      	jmp	0x3a2533 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xff3>
  3a1c78:      	mov	rdx, qword ptr [rbp - 0x50]
  3a1c7c:      	mov	rsi, qword ptr [rdx + 0x80]
  3a1c83:      	mov	edi, dword ptr [r15 + 0x4]
  3a1c87:      	cmp	rsi, rdi
  3a1c8a:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a1c90:      	imul	rcx, rdi, 0xe0
  3a1c97:      	add	rcx, qword ptr [rdx + 0x78]
  3a1c9b:      	mov	rdx, qword ptr [rbp - 0x58]
  3a1c9f:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1ca5:      	mov	rsi, qword ptr [rdx + 0x78]
  3a1ca9:      	mov	edi, dword ptr [rcx + 0x18]
  3a1cac:      	cmp	rsi, rdi
  3a1caf:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a1cb5:      	shl	rdi, 0x3
  3a1cb9:      	add	rdi, qword ptr [rdx + 0x70]
  3a1cbd:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1cc3:      	mov	rdx, qword ptr [rdi]
  3a1cc6:      	test	rdx, rdx
  3a1cc9:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1ccf:      	mov	rdi, qword ptr [rdx + 0x10]
  3a1cd3:      	mov	edx, dword ptr [rcx + 0x14]
  3a1cd6:      	mov	ecx, dword ptr [rcx + 0x1c]
  3a1cd9:      	shl	rcx, 0x20
  3a1cdd:      	or	rcx, rdx
  3a1ce0:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a1ce5:      	je	0x3a1cf2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x7b2>
  3a1ce7:      	cmp	qword ptr [rdi + 0x60], 0x0
  3a1cec:      	je	0x3a205f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xb1f>
  3a1cf2:      	mov	qword ptr [rbp - 0x90], r10
  3a1cf9:      	mov	qword ptr [rbp - 0x88], rax
  3a1d00:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a1d07:      	movups	xmmword ptr [rsp], xmm0
  3a1d0b:      	mov	esi, r9d
  3a1d0e:      	mov	rdx, r13
  3a1d11:      	xor	r8d, r8d
  3a1d14:      	call	0x3c7b30 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a1d19:      	jmp	0x3a2522 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe2>
  3a1d1e:      	xor	r10d, r10d
  3a1d21:      	xor	ecx, ecx
  3a1d23:      	mov	r12, qword ptr [rbp - 0x50]
  3a1d27:      	mov	edi, dword ptr [r15 + 0x10]
  3a1d2b:      	mov	esi, 0xffffffff
  3a1d30:      	cmp	rdi, r8
  3a1d33:      	jne	0x3a1883 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x343>
  3a1d39:      	xor	eax, eax
  3a1d3b:      	xor	r8d, r8d
  3a1d3e:      	mov	edi, dword ptr [r15 + 0x14]
  3a1d42:      	cmp	rdi, rsi
  3a1d45:      	jne	0x3a18aa <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x36a>
  3a1d4b:      	xor	r9d, r9d
  3a1d4e:      	xor	r11d, r11d
  3a1d51:      	mov	rsi, qword ptr [r12 + 0x80]
  3a1d59:      	mov	edi, dword ptr [r15 + 0x4]
  3a1d5d:      	cmp	rsi, rdi
  3a1d60:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a1d66:      	imul	rdi, rdi, 0xe0
  3a1d6d:      	add	rdi, qword ptr [r12 + 0x78]
  3a1d72:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1d78:      	movzx	esi, dx
  3a1d7b:      	cmp	esi, 0xb
  3a1d7e:      	setb	r12b
  3a1d82:      	mov	esi, edx
  3a1d84:      	and	esi, 0xfffc
  3a1d8a:      	cmp	esi, 0x54
  3a1d8d:      	sete	r13b
  3a1d91:      	or	r13b, r12b
  3a1d94:      	jne	0x3a1da9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x869>
  3a1d96:      	movzx	r12d, dx
  3a1d9a:      	lea	r13d, [r12 - 0x5c]
  3a1d9f:      	cmp	r13d, 0x2
  3a1da3:      	jae	0x3a1e87 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x947>
  3a1da9:      	mov	r10, qword ptr [rbp - 0x58]
  3a1dad:      	mov	rsi, qword ptr [r10 + 0x78]
  3a1db1:      	mov	r9d, dword ptr [rdi + 0x18]
  3a1db5:      	cmp	rsi, r9
  3a1db8:      	jbe	0x3a3560 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2020>
  3a1dbe:      	shl	r9, 0x3
  3a1dc2:      	add	r9, qword ptr [r10 + 0x70]
  3a1dc6:      	movzx	r12d, byte ptr [rbp - 0x31]
  3a1dcb:      	mov	r11, qword ptr [rbp - 0x78]
  3a1dcf:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1dd5:      	mov	rsi, qword ptr [r9]
  3a1dd8:      	test	rsi, rsi
  3a1ddb:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1de1:      	mov	rsi, qword ptr [rsi + 0x10]
  3a1de5:      	mov	r9d, dword ptr [rdi + 0x14]
  3a1de9:      	mov	r10d, dword ptr [rdi + 0x1c]
  3a1ded:      	shl	r10, 0x20
  3a1df1:      	or	r10, r9
  3a1df4:      	movzx	edi, byte ptr [rdi + 0x40]
  3a1df8:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a1dfd:      	je	0x3a1e0a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x8ca>
  3a1dff:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a1e04:      	je	0x3a1f54 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa14>
  3a1e0a:      	mov	qword ptr [rbp - 0xa0], rax
  3a1e11:      	mov	qword ptr [rbp - 0x98], r8
  3a1e18:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a1e1f:      	movups	xmmword ptr [rsp], xmm0
  3a1e23:      	movzx	r9d, dil
  3a1e27:      	mov	rdi, r11
  3a1e2a:      	mov	r8, r10
  3a1e2d:      	mov	r13, r11
  3a1e30:      	call	0x3c67c0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).load>
  3a1e35:      	jmp	0x3a1f9e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa5e>
  3a1e3a:      	lea	ecx, [r9 - 0x1000001]
  3a1e41:      	cmp	ecx, 0x8
  3a1e44:      	ja	0x3a2a81 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1541>
  3a1e4a:      	mov	edx, 0x117
  3a1e4f:      	bt	edx, ecx
  3a1e52:      	jae	0x3a2a6f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x152f>
  3a1e58:      	movzx	eax, r9b
  3a1e5c:      	xor	edi, edi
  3a1e5e:      	add	eax, -0x2
  3a1e61:      	cmp	eax, 0x7
  3a1e64:      	ja	0x3a2917 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13d7>
  3a1e6a:      	mov	r9, qword ptr [rbp - 0x48]
  3a1e6e:      	mov	r10, qword ptr [rbp - 0x60]
  3a1e72:      	jmp	qword ptr [8*rax + 0x2038f8]
  3a1e79:      	xor	edi, edi
  3a1e7b:      	test	r13d, r13d
  3a1e7e:      	sete	dil
  3a1e82:      	jmp	0x3a291f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13df>
  3a1e87:      	cmp	r12d, 0xb
  3a1e8b:      	je	0x3a1e96 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x956>
  3a1e8d:      	cmp	esi, 0x58
  3a1e90:      	jne	0x3a222d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xced>
  3a1e96:      	mov	r10, qword ptr [rbp - 0x58]
  3a1e9a:      	mov	rsi, qword ptr [r10 + 0x78]
  3a1e9e:      	mov	r9d, dword ptr [rdi + 0x18]
  3a1ea2:      	cmp	rsi, r9
  3a1ea5:      	movzx	r12d, byte ptr [rbp - 0x31]
  3a1eaa:      	jbe	0x3a3560 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2020>
  3a1eb0:      	shl	r9, 0x3
  3a1eb4:      	add	r9, qword ptr [r10 + 0x70]
  3a1eb8:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1ebe:      	mov	rsi, qword ptr [r9]
  3a1ec1:      	test	rsi, rsi
  3a1ec4:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a1eca:      	mov	rsi, qword ptr [rsi + 0x10]
  3a1ece:      	mov	r10d, dword ptr [rdi + 0x14]
  3a1ed2:      	mov	r9d, dword ptr [rdi + 0x1c]
  3a1ed6:      	shl	r9, 0x20
  3a1eda:      	or	r9, r10
  3a1edd:      	movzx	edi, byte ptr [rdi + 0x40]
  3a1ee1:      	cmp	qword ptr [rsi + 0x40], 0x0
  3a1ee6:      	je	0x3a1ef3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x9b3>
  3a1ee8:      	cmp	qword ptr [rsi + 0x60], 0x0
  3a1eed:      	je	0x3a24f5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfb5>
  3a1ef3:      	mov	qword ptr [rbp - 0x90], rax
  3a1efa:      	mov	qword ptr [rbp - 0x88], r8
  3a1f01:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a1f08:      	movups	xmmword ptr [rsp], xmm0
  3a1f0c:      	movzx	r8d, dil
  3a1f10:      	mov	rdi, rsi
  3a1f13:      	mov	esi, edx
  3a1f15:      	mov	rdx, rcx
  3a1f18:      	mov	rcx, r9
  3a1f1b:      	call	0x3c7b30 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).store>
  3a1f20:      	jmp	0x3a2522 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe2>
  3a1f25:      	mov	rax, rcx
  3a1f28:      	mul	r9
  3a1f2b:      	cmp	dword ptr [r15], 0xfc0015
  3a1f32:      	jne	0x3a1fe2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa2>
  3a1f38:      	mov	rdi, r9
  3a1f3b:      	sar	rdi, 0x3f
  3a1f3f:      	and	rdi, rcx
  3a1f42:      	sar	rcx, 0x3f
  3a1f46:      	and	rcx, r9
  3a1f49:      	add	rcx, rdi
  3a1f4c:      	sub	rdx, rcx
  3a1f4f:      	jmp	0x3a1fe2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xaa2>
  3a1f54:      	mov	qword ptr [rbp - 0xa0], rax
  3a1f5b:      	mov	qword ptr [rbp - 0x98], r8
  3a1f62:      	movups	xmm0, xmmword ptr [rbp - 0xa0]
  3a1f69:      	movups	xmmword ptr [rsp], xmm0
  3a1f6d:      	movzx	r9d, dil
  3a1f71:      	mov	rdi, r11
  3a1f74:      	mov	r8, r10
  3a1f77:      	mov	r13, r11
  3a1f7a:      	jmp	0x3a1f99 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xa59>
  3a1f7c:      	xorps	xmm0, xmm0
  3a1f7f:      	movaps	xmmword ptr [rbp - 0xa0], xmm0
  3a1f86:      	movups	xmmword ptr [rsp], xmm0
  3a1f8a:      	mov	rdi, r10
  3a1f8d:      	mov	edx, r9d
  3a1f90:      	mov	rcx, r13
  3a1f93:      	xor	r9d, r9d
  3a1f96:      	mov	r13, r10
  3a1f99:      	call	0x3c6900 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).loadLocked>
  3a1f9e:      	mov	rax, qword ptr [r13]
  3a1fa2:      	mov	rdx, qword ptr [r13 + 0x8]
  3a1fa6:      	mov	qword ptr [rbp - 0x78], r13
  3a1faa:      	movzx	r13d, byte ptr [r13 + 0x10]
  3a1faf:      	jmp	0x3a2529 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe9>
  3a1fb4:      	mov	rax, qword ptr [r15 + 0x18]
  3a1fb8:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a1fbd:      	mov	rax, r9
  3a1fc0:      	sub	rax, rdi
  3a1fc3:      	not	r9
  3a1fc6:      	mov	r10, rdi
  3a1fc9:      	and	r10, r9
  3a1fcc:      	xor	r9, rdi
  3a1fcf:      	and	r9, rax
  3a1fd2:      	or	r9, r10
  3a1fd5:      	shr	r9, 0x3f
  3a1fd9:      	add	rdx, r9
  3a1fdc:      	sub	rcx, rdx
  3a1fdf:      	mov	rdx, rcx
  3a1fe2:      	mov	edi, dword ptr [r15 + 0x1c]
  3a1fe6:      	cmp	rsi, rdi
  3a1fe9:      	mov	r9, qword ptr [rbp - 0x48]
  3a1fed:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a1ff3:      	lea	ecx, [rdi + 0x1]
  3a1ff6:      	cmp	rsi, rcx
  3a1ff9:      	mov	r10, qword ptr [rbp - 0x68]
  3a1ffd:      	jbe	0x3a3574 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2034>
  3a2003:      	mov	esi, dword ptr [r8 + 4*rdi]
  3a2007:      	cmp	r9, rsi
  3a200a:      	jbe	0x3a357c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x203c>
  3a2010:      	mov	edi, dword ptr [r8 + 4*rcx]
  3a2014:      	shl	rsi, 0x4
  3a2018:      	mov	rcx, qword ptr [rbp - 0x40]
  3a201c:      	mov	qword ptr [rcx + rsi], rax
  3a2020:      	mov	qword ptr [rcx + rsi + 0x8], 0x0
  3a2029:      	cmp	r9, rdi
  3a202c:      	jbe	0x3a3558 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2018>
  3a2032:      	mov	rsi, rcx
  3a2035:      	shl	rdi, 0x4
  3a2039:      	mov	qword ptr [rcx + rdi], rdx
  3a203d:      	mov	qword ptr [rcx + rdi + 0x8], 0x0
  3a2046:      	inc	r14d
  3a2049:      	mov	dword ptr [r10 - 0x38], r14d
  3a204d:      	mov	r13d, r14d
  3a2050:      	mov	r8d, 0xffffffff
  3a2056:      	mov	r10, qword ptr [rbp - 0x60]
  3a205a:      	jmp	0x3a1653 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x113>
  3a205f:      	mov	qword ptr [rbp - 0x90], r10
  3a2066:      	mov	qword ptr [rbp - 0x88], rax
  3a206d:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a2074:      	movups	xmmword ptr [rsp], xmm0
  3a2078:      	mov	esi, r9d
  3a207b:      	mov	rdx, r13
  3a207e:      	xor	r8d, r8d
  3a2081:      	jmp	0x3a251d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfdd>
  3a2086:      	xor	r8d, r8d
  3a2089:      	mov	edi, dword ptr [r15 + 0x10]
  3a208d:      	cmp	rdi, rax
  3a2090:      	jne	0x3a1a58 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x518>
  3a2096:      	xor	eax, eax
  3a2098:      	xor	r9d, r9d
  3a209b:      	mov	r10d, 0x0
  3a20a1:      	mov	r11d, 0x0
  3a20a7:      	cmp	dx, 0xc
  3a20ab:      	jne	0x3a20ed <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xbad>
  3a20ad:      	mov	rcx, qword ptr [rbp - 0x50]
  3a20b1:      	mov	rsi, qword ptr [rcx + 0xd0]
  3a20b8:      	mov	edi, dword ptr [r15 + 0x18]
  3a20bc:      	add	edi, 0x2
  3a20bf:      	cmp	rsi, rdi
  3a20c2:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a20c8:      	mov	rcx, qword ptr [rcx + 0xc8]
  3a20cf:      	mov	edi, dword ptr [rcx + 4*rdi]
  3a20d2:      	cmp	qword ptr [rbp - 0x48], rdi
  3a20d6:      	jbe	0x3a354f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x200f>
  3a20dc:      	shl	rdi, 0x4
  3a20e0:      	mov	rcx, qword ptr [rbp - 0x40]
  3a20e4:      	mov	r10, qword ptr [rcx + rdi]
  3a20e8:      	mov	r11, qword ptr [rcx + rdi + 0x8]
  3a20ed:      	cmp	qword ptr [rbp - 0x30], 0x0
  3a20f2:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a20f8:      	mov	rdx, qword ptr [rbp - 0x50]
  3a20fc:      	mov	rsi, qword ptr [rdx + 0x80]
  3a2103:      	mov	edi, dword ptr [r15 + 0x4]
  3a2107:      	cmp	rsi, rdi
  3a210a:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a2110:      	imul	rcx, rdi, 0xe0
  3a2117:      	add	rcx, qword ptr [rdx + 0x78]
  3a211b:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2121:      	mov	rdx, qword ptr [rbp - 0x30]
  3a2125:      	mov	rsi, qword ptr [rdx]
  3a2128:      	mov	qword ptr [rbp - 0xd0], rax
  3a212f:      	mov	qword ptr [rbp - 0xc8], r9
  3a2136:      	mov	qword ptr [rbp - 0xc0], r10
  3a213d:      	mov	qword ptr [rbp - 0xb8], r11
  3a2144:      	movups	xmm0, xmmword ptr [rbp - 0xc0]
  3a214b:      	movups	xmmword ptr [rsp + 0x10], xmm0
  3a2150:      	movups	xmm0, xmmword ptr [rbp - 0xd0]
  3a2157:      	movups	xmmword ptr [rsp], xmm0
  3a215b:      	lea	rdi, [rbp - 0x150]
  3a2162:      	mov	rdx, qword ptr [rbp - 0x58]
  3a2166:      	call	0x3b69c0 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GCAccess>
  3a216b:      	movaps	xmm0, xmmword ptr [rbp - 0x150]
  3a2172:      	mov	r13, qword ptr [rbp - 0x140]
  3a2179:      	test	r13, r13
  3a217c:      	je	0x3a219c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc5c>
  3a217e:      	cmp	qword ptr [r13 + 0x8], 0x0
  3a2183:      	mov	r8d, 0xffffffff
  3a2189:      	mov	r9, qword ptr [rbp - 0x48]
  3a218d:      	mov	rsi, qword ptr [rbp - 0x40]
  3a2191:      	mov	r10, qword ptr [rbp - 0x60]
  3a2195:      	je	0x3a21ae <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xc6e>
  3a2197:      	jmp	0x3a315e <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c1e>
  3a219c:      	mov	r8d, 0xffffffff
  3a21a2:      	mov	r9, qword ptr [rbp - 0x48]
  3a21a6:      	mov	rsi, qword ptr [rbp - 0x40]
  3a21aa:      	mov	r10, qword ptr [rbp - 0x60]
  3a21ae:      	mov	edi, dword ptr [r15 + 0x8]
  3a21b2:      	cmp	rdi, r8
  3a21b5:      	mov	r11, qword ptr [rbp - 0x30]
  3a21b9:      	mov	r15, qword ptr [rbp - 0x68]
  3a21bd:      	je	0x3a1c3a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6fa>
  3a21c3:      	cmp	r9, rdi
  3a21c6:      	jbe	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a21cc:      	shl	rdi, 0x4
  3a21d0:      	movups	xmmword ptr [rsi + rdi], xmm0
  3a21d4:      	jmp	0x3a1c3a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x6fa>
  3a21d9:      	mov	edi, dword ptr [r15 + 0x18]
  3a21dd:      	cmp	qword ptr [rbp - 0x48], rdi
  3a21e1:      	jbe	0x3a354f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x200f>
  3a21e7:      	shl	rdi, 0x4
  3a21eb:      	mov	qword ptr [rsi + rdi], r13
  3a21ef:      	mov	qword ptr [rsi + rdi + 0x8], rdx
  3a21f4:      	mov	rax, r13
  3a21f7:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a21fc:      	mov	edi, dword ptr [r15 + 0x14]
  3a2200:      	cmp	rdi, r8
  3a2203:      	je	0x3a290c <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13cc>
  3a2209:      	cmp	qword ptr [rbp - 0x48], rdi
  3a220d:      	jbe	0x3a354f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x200f>
  3a2213:      	shl	rdi, 0x4
  3a2217:      	mov	rcx, qword ptr [rsi + rdi]
  3a221b:      	test	ecx, ecx
  3a221d:      	cmove	rdx, rax
  3a2221:      	cmove	r13, r10
  3a2225:      	mov	rax, r13
  3a2228:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a222d:      	mov	qword ptr [rsp], r11
  3a2231:      	mov	rsi, rcx
  3a2234:      	mov	rdx, r10
  3a2237:      	mov	rcx, rax
  3a223a:      	call	0x3ce380 <github.com/cpunion/wasmvm/internal/engine/simd.Eval>
  3a223f:      	mov	r11, qword ptr [rbp - 0x30]
  3a2243:      	movzx	r12d, byte ptr [rbp - 0x31]
  3a2248:      	jmp	0x3a1b17 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d7>
  3a224d:      	mov	edi, dword ptr [r15 + 0x1c]
  3a2251:      	dec	edi
  3a2253:      	cmp	edi, r13d
  3a2256:      	cmovae	edi, r13d
  3a225a:      	jmp	0x3a291f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13df>
  3a225f:      	xor	edi, edi
  3a2261:      	test	r13, r13
  3a2264:      	setne	dil
  3a2268:      	jmp	0x3a291f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x13df>
  3a226d:      	xor	eax, eax
  3a226f:      	cmp	r13, r10
  3a2272:      	sete	al
  3a2275:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a227a:      	mov	rax, qword ptr [rbp - 0x58]
  3a227e:      	mov	rsi, qword ptr [rax + 0x78]
  3a2282:      	mov	edi, dword ptr [r15 + 0x18]
  3a2286:      	cmp	rsi, rdi
  3a2289:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a228f:      	shl	rdi, 0x3
  3a2293:      	add	rdi, qword ptr [rax + 0x70]
  3a2297:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a229d:      	test	r11, r11
  3a22a0:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a22a6:      	mov	rsi, qword ptr [rdi]
  3a22a9:      	test	rsi, rsi
  3a22ac:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a22b2:      	mov	rdi, qword ptr [r11]
  3a22b5:      	mov	rax, qword ptr [rsi + 0x10]
  3a22b9:      	movzx	eax, byte ptr [rax + 0x31]
  3a22bd:      	mov	byte ptr [rbp - 0x69], al
  3a22c0:      	mov	edx, r13d
  3a22c3:      	test	al, 0x1
  3a22c5:      	cmovne	rdx, r13
  3a22c9:      	mov	r13, rsi
  3a22cc:      	call	0x3c0c30 <github.com/cpunion/wasmvm/internal/engine/state.(*Store).GrowMemory>
  3a22d1:      	test	dl, 0x1
  3a22d4:      	jne	0x3a1b11 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d1>
  3a22da:      	mov	rax, qword ptr [r13 + 0x10]
  3a22de:      	movzx	eax, byte ptr [rax + 0x31]
  3a22e2:      	mov	byte ptr [rbp - 0x6a], al
  3a22e5:      	and	al, 0x1
  3a22e7:      	neg	al
  3a22e9:      	mov	eax, 0x0
  3a22ee:      	sbb	rax, rax
  3a22f1:      	mov	r8d, 0xffffffff
  3a22f7:      	or	rax, r8
  3a22fa:      	jmp	0x3a24dd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf9d>
  3a22ff:      	mov	rax, qword ptr [rbp - 0x58]
  3a2303:      	mov	rsi, qword ptr [rax + 0x60]
  3a2307:      	mov	edi, dword ptr [r15 + 0x18]
  3a230b:      	cmp	rsi, rdi
  3a230e:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a2314:      	mov	rax, qword ptr [rax + 0x58]
  3a2318:      	lea	rcx, [rax + 8*rdi]
  3a231c:      	test	rcx, rcx
  3a231f:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2325:      	mov	rax, qword ptr [rax + 8*rdi]
  3a2329:      	test	rax, rax
  3a232c:      	mov	rsi, qword ptr [rbp - 0x40]
  3a2330:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2336:      	cmp	byte ptr [rax + 0x29], 0x0
  3a233a:      	mov	ecx, r13d
  3a233d:      	cmovne	rcx, r13
  3a2341:      	cmp	rcx, qword ptr [rax + 0x40]
  3a2345:      	jae	0x3a32ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dac>
  3a234b:      	mov	rax, qword ptr [rax + 0x38]
  3a234f:      	mov	rax, qword ptr [rax + 8*rcx]
  3a2353:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2358:      	mov	rax, qword ptr [rbp - 0x58]
  3a235c:      	mov	rsi, qword ptr [rax + 0x48]
  3a2360:      	mov	edi, dword ptr [r15 + 0x18]
  3a2364:      	cmp	rsi, rdi
  3a2367:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a236d:      	mov	rax, qword ptr [rax + 0x40]
  3a2371:      	mov	rax, qword ptr [rax + 8*rdi]
  3a2375:      	xor	edx, edx
  3a2377:      	mov	rsi, qword ptr [rbp - 0x40]
  3a237b:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a2380:      	mov	rax, r13
  3a2383:      	test	r13, r13
  3a2386:      	jne	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a238c:      	jmp	0x3a3306 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc6>
  3a2391:      	mov	rax, qword ptr [rbp - 0x58]
  3a2395:      	mov	rsi, qword ptr [rax + 0x90]
  3a239c:      	mov	edi, dword ptr [r15 + 0x18]
  3a23a0:      	cmp	rsi, rdi
  3a23a3:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a23a9:      	mov	rax, qword ptr [rax + 0x88]
  3a23b0:      	mov	rax, qword ptr [rax + 8*rdi]
  3a23b4:      	mov	qword ptr [rax + 0x20], r13
  3a23b8:      	mov	qword ptr [rax + 0x28], rdx
  3a23bc:      	xor	edx, edx
  3a23be:      	xor	eax, eax
  3a23c0:      	mov	rsi, qword ptr [rbp - 0x40]
  3a23c4:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a23c9:      	mov	rax, qword ptr [rbp - 0x58]
  3a23cd:      	mov	rsi, qword ptr [rax + 0x90]
  3a23d4:      	mov	edi, dword ptr [r15 + 0x18]
  3a23d8:      	cmp	rsi, rdi
  3a23db:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a23e1:      	mov	rax, qword ptr [rax + 0x88]
  3a23e8:      	mov	rcx, qword ptr [rax + 8*rdi]
  3a23ec:      	mov	rax, qword ptr [rcx + 0x20]
  3a23f0:      	mov	rdx, qword ptr [rcx + 0x28]
  3a23f4:      	mov	rsi, qword ptr [rbp - 0x40]
  3a23f8:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a23fd:      	mov	rax, qword ptr [rbp - 0x58]
  3a2401:      	mov	rsi, qword ptr [rax + 0x60]
  3a2405:      	mov	edi, dword ptr [r15 + 0x18]
  3a2409:      	cmp	rsi, rdi
  3a240c:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a2412:      	mov	rax, qword ptr [rax + 0x58]
  3a2416:      	lea	rcx, [rax + 8*rdi]
  3a241a:      	test	rcx, rcx
  3a241d:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2423:      	mov	rax, qword ptr [rax + 8*rdi]
  3a2427:      	test	rax, rax
  3a242a:      	mov	rsi, qword ptr [rbp - 0x40]
  3a242e:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2434:      	cmp	byte ptr [rax + 0x29], 0x0
  3a2438:      	mov	ecx, r13d
  3a243b:      	cmovne	rcx, r13
  3a243f:      	cmp	rcx, qword ptr [rax + 0x40]
  3a2443:      	jae	0x3a32ec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dac>
  3a2449:      	mov	rax, qword ptr [rax + 0x38]
  3a244d:      	mov	qword ptr [rax + 8*rcx], r10
  3a2451:      	xor	edx, edx
  3a2453:      	xor	eax, eax
  3a2455:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a245a:      	mov	edi, dword ptr [r15 + 0x18]
  3a245e:      	cmp	qword ptr [rbp - 0x48], rdi
  3a2462:      	jbe	0x3a354f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x200f>
  3a2468:      	shl	rdi, 0x4
  3a246c:      	mov	rax, qword ptr [rsi + rdi]
  3a2470:      	mov	rdx, qword ptr [rsi + rdi + 0x8]
  3a2475:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a247a:      	xor	eax, eax
  3a247c:      	test	r13, r13
  3a247f:      	sete	al
  3a2482:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2487:      	mov	rax, qword ptr [rbp - 0x58]
  3a248b:      	mov	rsi, qword ptr [rax + 0x78]
  3a248f:      	mov	edi, dword ptr [r15 + 0x18]
  3a2493:      	cmp	rsi, rdi
  3a2496:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a249c:      	shl	rdi, 0x3
  3a24a0:      	add	rdi, qword ptr [rax + 0x70]
  3a24a4:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a24aa:      	mov	rax, qword ptr [rdi]
  3a24ad:      	test	rax, rax
  3a24b0:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a24b6:      	mov	rdi, qword ptr [rax + 0x10]
  3a24ba:      	cmp	qword ptr [rdi + 0x40], 0x0
  3a24bf:      	mov	rax, qword ptr [rdi + 0x60]
  3a24c3:      	je	0x3a2c62 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1722>
  3a24c9:      	test	rax, rax
  3a24cc:      	jne	0x3a2c6b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x172b>
  3a24d2:      	call	0x3c4480 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).pagesLocked>
  3a24d7:      	mov	r8d, 0xffffffff
  3a24dd:      	xor	edx, edx
  3a24df:      	mov	r11, qword ptr [rbp - 0x30]
  3a24e3:      	mov	rsi, qword ptr [rbp - 0x40]
  3a24e7:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a24ec:      	mov	eax, dword ptr [r15 + 0x18]
  3a24f0:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a24f5:      	mov	qword ptr [rbp - 0x90], rax
  3a24fc:      	mov	qword ptr [rbp - 0x88], r8
  3a2503:      	movups	xmm0, xmmword ptr [rbp - 0x90]
  3a250a:      	movups	xmmword ptr [rsp], xmm0
  3a250e:      	movzx	r8d, dil
  3a2512:      	mov	rdi, rsi
  3a2515:      	mov	esi, edx
  3a2517:      	mov	rdx, rcx
  3a251a:      	mov	rcx, r9
  3a251d:      	call	0x3c7e10 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).storeLocked>
  3a2522:      	mov	r13d, eax
  3a2525:      	xor	edx, edx
  3a2527:      	xor	eax, eax
  3a2529:      	mov	r11, qword ptr [rbp - 0x30]
  3a252d:      	mov	r8d, 0xffffffff
  3a2533:      	test	r13b, r13b
  3a2536:      	mov	rsi, qword ptr [rbp - 0x40]
  3a253a:      	je	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a2540:      	jmp	0x3a3309 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc9>
  3a2545:      	xor	r10, r13
  3a2548:      	jmp	0x3a28a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1369>
  3a254d:      	mov	eax, r13d
  3a2550:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2555:      	mov	ecx, r10d
  3a2558:      	shr	r13, cl
  3a255b:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a2560:      	mov	ecx, r10d
  3a2563:      	shr	r13d, cl
  3a2566:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a256b:      	or	r10, r13
  3a256e:      	jmp	0x3a28a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1369>
  3a2573:      	mov	ecx, r10d
  3a2576:      	sar	r13d, cl
  3a2579:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a257e:      	xor	r10d, r13d
  3a2581:      	jmp	0x3a28a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1369>
  3a2586:      	mov	rax, r13
  3a2589:      	shr	rax
  3a258c:      	movabs	rcx, 0x5555555555555555
  3a2596:      	and	rax, rcx
  3a2599:      	and	r13, rcx
  3a259c:      	add	r13, rax
  3a259f:      	mov	rax, r13
  3a25a2:      	shr	rax, 0x2
  3a25a6:      	movabs	rcx, 0x3333333333333333
  3a25b0:      	and	rax, rcx
  3a25b3:      	and	r13, rcx
  3a25b6:      	add	r13, rax
  3a25b9:      	mov	rax, r13
  3a25bc:      	shr	rax, 0x4
  3a25c0:      	add	rax, r13
  3a25c3:      	movabs	rcx, 0xf0f0f0f0f0f0f0f
  3a25cd:      	and	rcx, rax
  3a25d0:      	mov	rax, rcx
  3a25d3:      	shr	rax, 0x8
  3a25d7:      	add	rax, rcx
  3a25da:      	mov	rcx, rax
  3a25dd:      	shr	rcx, 0x10
  3a25e1:      	add	rcx, rax
  3a25e4:      	mov	rax, rcx
  3a25e7:      	shr	rax, 0x20
  3a25eb:      	add	eax, ecx
  3a25ed:      	and	eax, 0x7f
  3a25f0:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a25f5:      	xor	eax, eax
  3a25f7:      	cmp	r13, r10
  3a25fa:      	setb	al
  3a25fd:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2602:      	xor	eax, eax
  3a2604:      	cmp	r13, r10
  3a2607:      	setbe	al
  3a260a:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a260f:      	xor	eax, eax
  3a2611:      	cmp	r13, r10
  3a2614:      	seta	al
  3a2617:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a261c:      	imul	r10d, r13d
  3a2620:      	jmp	0x3a28a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1369>
  3a2625:      	mov	ecx, r13d
  3a2628:      	shr	ecx, 0x10
  3a262b:      	xor	eax, eax
  3a262d:      	cmp	r13d, 0x10000
  3a2634:      	setb	al
  3a2637:      	cmovb	ecx, r13d
  3a263b:      	shl	eax, 0x4
  3a263e:      	mov	edx, ecx
  3a2640:      	shr	edx, 0x8
  3a2643:      	xor	esi, esi
  3a2645:      	cmp	ecx, 0x100
  3a264b:      	setb	sil
  3a264f:      	lea	eax, [rax + 8*rsi]
  3a2652:      	mov	rsi, qword ptr [rbp - 0x40]
  3a2656:      	cmovb	edx, ecx
  3a2659:      	movzx	ecx, dl
  3a265c:      	movzx	ecx, byte ptr [rcx + 0x328451]
  3a2663:      	sub	rax, rcx
  3a2666:      	add	rax, 0x8
  3a266a:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a266f:      	mov	ecx, r10d
  3a2672:      	rol	r13d, cl
  3a2675:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a267a:      	xor	eax, eax
  3a267c:      	cmp	r13, r10
  3a267f:      	setg	al
  3a2682:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2687:      	xor	eax, eax
  3a2689:      	cmp	r13d, r10d
  3a268c:      	setl	al
  3a268f:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2694:      	xor	eax, eax
  3a2696:      	cmp	r13, r10
  3a2699:      	setne	al
  3a269c:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a26a1:      	test	r13d, r13d
  3a26a4:      	je	0x3a2ca9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1769>
  3a26aa:      	tzcnt	eax, r13d
  3a26af:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a26b4:      	add	r10d, r13d
  3a26b7:      	jmp	0x3a28a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1369>
  3a26bc:      	xor	eax, eax
  3a26be:      	cmp	r13d, r10d
  3a26c1:      	setne	al
  3a26c4:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a26c9:      	xor	eax, eax
  3a26cb:      	cmp	r13, r10
  3a26ce:      	setle	al
  3a26d1:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a26d6:      	mov	ecx, r10d
  3a26d9:      	ror	r13d, cl
  3a26dc:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a26e1:      	xor	eax, eax
  3a26e3:      	cmp	r13d, r10d
  3a26e6:      	setbe	al
  3a26e9:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a26ee:      	xor	eax, eax
  3a26f0:      	cmp	r13d, r10d
  3a26f3:      	setge	al
  3a26f6:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a26fb:      	xor	eax, eax
  3a26fd:      	test	r13d, r13d
  3a2700:      	sete	al
  3a2703:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2708:      	xor	eax, eax
  3a270a:      	cmp	r13d, r10d
  3a270d:      	setb	al
  3a2710:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2715:      	mov	rdx, r13
  3a2718:      	mov	eax, edx
  3a271a:      	shr	eax, 0x18
  3a271d:      	mov	ecx, edx
  3a271f:      	shr	ecx, 0x10
  3a2722:      	movzx	ecx, cl
  3a2725:      	movzx	ecx, byte ptr [rcx + 0x328551]
  3a272c:      	add	cl, byte ptr [rax + 0x328551]
  3a2732:      	movzx	eax, dh
  3a2735:      	add	cl, byte ptr [rax + 0x328551]
  3a273b:      	movzx	eax, dl
  3a273e:      	add	cl, byte ptr [rax + 0x328551]
  3a2744:      	movzx	eax, cl
  3a2747:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a274c:      	add	r10, r13
  3a274f:      	jmp	0x3a28a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1369>
  3a2754:      	mov	ecx, r10d
  3a2757:      	shl	r13d, cl
  3a275a:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a275f:      	xor	eax, eax
  3a2761:      	test	r13, r13
  3a2764:      	sete	al
  3a2767:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a276c:      	and	r10, r13
  3a276f:      	jmp	0x3a28a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1369>
  3a2774:      	sub	r13, r10
  3a2777:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a277c:      	xor	eax, eax
  3a277e:      	cmp	r13d, r10d
  3a2781:      	sete	al
  3a2784:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2789:      	mov	ecx, r10d
  3a278c:      	rol	r13, cl
  3a278f:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a2794:      	xor	eax, eax
  3a2796:      	cmp	r13d, r10d
  3a2799:      	setae	al
  3a279c:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a27a1:      	and	r13d, r10d
  3a27a4:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a27a9:      	xor	eax, eax
  3a27ab:      	cmp	r13d, r10d
  3a27ae:      	setle	al
  3a27b1:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a27b6:      	xor	eax, eax
  3a27b8:      	cmp	r13d, r10d
  3a27bb:      	setg	al
  3a27be:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a27c3:      	xor	eax, eax
  3a27c5:      	mov	rcx, r13
  3a27c8:      	shr	rcx, 0x20
  3a27cc:      	cmove	rcx, r13
  3a27d0:      	setne	al
  3a27d3:      	shl	eax, 0x5
  3a27d6:      	mov	rdx, rcx
  3a27d9:      	shr	rdx, 0x10
  3a27dd:      	lea	rsi, [rax + 0x10]
  3a27e1:      	cmp	rcx, 0x10000
  3a27e8:      	cmovb	rdx, rcx
  3a27ec:      	cmovb	rsi, rax
  3a27f0:      	mov	eax, edx
  3a27f2:      	shr	eax, 0x8
  3a27f5:      	lea	rcx, [rsi + 0x8]
  3a27f9:      	cmp	rdx, 0x100
  3a2800:      	cmovb	rax, rdx
  3a2804:      	cmovb	rcx, rsi
  3a2808:      	mov	rsi, qword ptr [rbp - 0x40]
  3a280c:      	movzx	eax, al
  3a280f:      	movzx	edx, byte ptr [rax + 0x328451]
  3a2816:      	add	rdx, rcx
  3a2819:      	mov	eax, 0x40
  3a281e:      	sub	rax, rdx
  3a2821:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2826:      	xor	eax, eax
  3a2828:      	cmp	r13, r10
  3a282b:      	setl	al
  3a282e:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2833:      	mov	ecx, r10d
  3a2836:      	ror	r13, cl
  3a2839:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a283e:      	xor	eax, eax
  3a2840:      	cmp	r13d, r10d
  3a2843:      	seta	al
  3a2846:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a284b:      	mov	eax, r13d
  3a284e:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2853:      	movsx	rax, r13w
  3a2857:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a285c:      	movsxd	rax, r13d
  3a285f:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2864:      	xor	eax, eax
  3a2866:      	cmp	r13, r10
  3a2869:      	sete	al
  3a286c:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2871:      	movsxd	rax, r13d
  3a2874:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2879:      	xor	eax, eax
  3a287b:      	cmp	r13, r10
  3a287e:      	setae	al
  3a2881:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2886:      	xor	eax, eax
  3a2888:      	cmp	r13, r10
  3a288b:      	setge	al
  3a288e:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2893:      	mov	ecx, r10d
  3a2896:      	sar	r13, cl
  3a2899:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a289b:      	or	r10d, r13d
  3a289e:      	jmp	0x3a28a9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1369>
  3a28a0:      	sub	r13d, r10d
  3a28a3:      	jmp	0x3a28d3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1393>
  3a28a5:      	imul	r10, r13
  3a28a9:      	xor	edx, edx
  3a28ab:      	mov	rax, r10
  3a28ae:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a28b3:      	movsx	rax, r13w
  3a28b7:      	mov	eax, eax
  3a28b9:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a28be:      	mov	eax, 0x40
  3a28c3:      	tzcnt	rax, r13
  3a28c8:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a28cd:      	mov	ecx, r10d
  3a28d0:      	shl	r13, cl
  3a28d3:      	xor	edx, edx
  3a28d5:      	mov	rax, r13
  3a28d8:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a28dd:      	movsx	rax, r13b
  3a28e1:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a28e6:      	movsx	rax, r13b
  3a28ea:      	mov	eax, eax
  3a28ec:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a28f1:      	movzx	ecx, r9b
  3a28f5:      	add	ecx, -0x61
  3a28f8:      	cmp	ecx, 0x5e
  3a28fb:      	ja	0x3a2451 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xf11>
  3a2901:      	xor	edx, edx
  3a2903:      	xor	eax, eax
  3a2905:      	jmp	qword ptr [8*rcx + 0x203d38]
  3a290c:      	mov	rdx, rax
  3a290f:      	mov	rax, r10
  3a2912:      	jmp	0x3a2d92 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1852>
  3a2917:      	mov	r9, qword ptr [rbp - 0x48]
  3a291b:      	mov	r10, qword ptr [rbp - 0x60]
  3a291f:      	mov	rax, qword ptr [rbp - 0x50]
  3a2923:      	mov	rsi, qword ptr [rax + 0xe8]
  3a292a:      	add	edi, dword ptr [r15 + 0x18]
  3a292e:      	cmp	rsi, rdi
  3a2931:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a2937:      	mov	rax, qword ptr [rax + 0xe0]
  3a293e:      	lea	rcx, [rdi + 2*rdi]
  3a2942:      	mov	r13d, dword ptr [rax + 4*rcx]
  3a2946:      	mov	edi, dword ptr [rax + 4*rcx + 0x4]
  3a294a:      	mov	r15d, dword ptr [rax + 4*rcx + 0x8]
  3a294f:      	cmp	r13d, r14d
  3a2952:      	mov	rsi, qword ptr [rbp - 0x40]
  3a2956:      	ja	0x3a29e2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14a2>
  3a295c:      	mov	rax, qword ptr [r11 + 0x160]
  3a2963:      	test	rax, rax
  3a2966:      	je	0x3a2977 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1437>
  3a2968:      	test	r11, r11
  3a296b:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2971:      	mov	eax, dword ptr [rax]
  3a2973:      	test	eax, eax
  3a2975:      	je	0x3a29d9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1499>
  3a2977:      	mov	rax, qword ptr [r11 + 0x8]
  3a297b:      	test	rax, rax
  3a297e:      	je	0x3a3629 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20e9>
  3a2984:      	mov	r14, rdi
  3a2987:      	mov	rdi, qword ptr [r11 + 0x10]
  3a298b:      	mov	rcx, qword ptr [rax + 0x8]
  3a298f:      	test	byte ptr [rcx + 0x17], 0x20
  3a2993:      	je	0x3a29ac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x146c>
  3a2995:      	movzx	ecx, byte ptr [rcx + 0x17]
  3a2999:      	and	cl, 0x1f
  3a299c:      	cmp	cl, 0x16
  3a299f:      	je	0x3a29ac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x146c>
  3a29a1:      	mov	qword ptr [rbp - 0x90], rdi
  3a29a8:      	mov	rdi, qword ptr [rbp - 0x78]
  3a29ac:      	call	qword ptr [rax + 0x28]
  3a29af:      	mov	r11, qword ptr [rbp - 0x30]
  3a29b3:      	mov	r8d, 0xffffffff
  3a29b9:      	mov	r9, qword ptr [rbp - 0x48]
  3a29bd:      	mov	rsi, qword ptr [rbp - 0x40]
  3a29c1:      	mov	r10, qword ptr [rbp - 0x60]
  3a29c5:      	mov	rdi, r14
  3a29c8:      	test	rax, rax
  3a29cb:      	je	0x3a29e2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14a2>
  3a29cd:      	cmp	qword ptr [rax + 0x8], 0x0
  3a29d2:      	je	0x3a29e2 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14a2>
  3a29d4:      	jmp	0x3a32f1 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1db1>
  3a29d9:      	xor	eax, eax
  3a29db:      	xor	edx, edx
  3a29dd:      	test	rax, rax
  3a29e0:      	jne	0x3a29cd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x148d>
  3a29e2:      	mov	rax, qword ptr [rbp - 0x50]
  3a29e6:      	mov	rdx, qword ptr [rax + 0x108]
  3a29ed:      	add	r15d, edi
  3a29f0:      	cmp	rdx, r15
  3a29f3:      	jb	0x3a361a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20da>
  3a29f9:      	cmp	edi, r15d
  3a29fc:      	ja	0x3a360b <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20cb>
  3a2a02:      	sub	rdx, rdi
  3a2a05:      	test	rdx, rdx
  3a2a08:      	mov	eax, 0x0
  3a2a0d:      	cmovg	rax, rdi
  3a2a11:      	sub	r15, rdi
  3a2a14:      	jle	0x3a1653 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x113>
  3a2a1a:      	mov	rcx, qword ptr [rbp - 0x50]
  3a2a1e:      	mov	rcx, qword ptr [rcx + 0xf8]
  3a2a25:      	lea	rcx, [rcx + 8*rax]
  3a2a29:      	add	rcx, 0x4
  3a2a2d:      	xor	edx, edx
  3a2a2f:      	nop
  3a2a30:      	cmp	r15, rdx
  3a2a33:      	je	0x3a3539 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff9>
  3a2a39:      	mov	edi, dword ptr [rcx + 8*rdx]
  3a2a3c:      	cmp	r9, rdi
  3a2a3f:      	jbe	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a2a45:      	mov	eax, dword ptr [rcx + 8*rdx - 0x4]
  3a2a49:      	cmp	r9, rax
  3a2a4c:      	jbe	0x3a3544 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2004>
  3a2a52:      	shl	rdi, 0x4
  3a2a56:      	shl	rax, 0x4
  3a2a5a:      	movups	xmm0, xmmword ptr [rsi + rdi]
  3a2a5e:      	movups	xmmword ptr [rsi + rax], xmm0
  3a2a62:      	inc	rdx
  3a2a65:      	cmp	r15, rdx
  3a2a68:      	jne	0x3a2a30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x14f0>
  3a2a6a:      	jmp	0x3a1653 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x113>
  3a2a6f:      	cmp	ecx, 0x3
  3a2a72:      	je	0x3a2e39 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18f9>
  3a2a78:      	cmp	ecx, 0x6
  3a2a7b:      	je	0x3a2dd9 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1899>
  3a2a81:      	lea	ecx, [r9 - 0x12]
  3a2a85:      	cmp	ecx, 0x4
  3a2a88:      	jb	0x3a2e0f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18cf>
  3a2a8e:      	mov	edi, dword ptr [r15 + 0x14]
  3a2a92:      	cmp	rdi, r8
  3a2a95:      	je	0x3a2aef <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15af>
  3a2a97:      	cmp	qword ptr [rbp - 0x48], rdi
  3a2a9b:      	jbe	0x3a354f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x200f>
  3a2aa1:      	shl	rdi, 0x4
  3a2aa5:      	movups	xmm0, xmmword ptr [rsi + rdi]
  3a2aa9:      	mov	ecx, r9d
  3a2aac:      	and	ecx, 0xffff0000
  3a2ab2:      	cmp	ecx, 0xfe0000
  3a2ab8:      	jne	0x3a2b03 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x15c3>
  3a2aba:      	movzx	r9d, r9w
  3a2abe:      	cmp	r9d, 0x3
  3a2ac2:      	jne	0x3a2cb3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1773>
  3a2ac8:      	test	r11, r11
  3a2acb:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2ad1:      	mov	rax, qword ptr [r11]
  3a2ad4:      	test	rax, rax
  3a2ad7:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2add:      	mov	rdi, qword ptr [rax + 0x8]
  3a2ae1:      	call	0x3c3a60 <github.com/cpunion/wasmvm/internal/engine/memory.(*Domain).Fence>
  3a2ae6:      	xor	edx, edx
  3a2ae8:      	xor	eax, eax
  3a2aea:      	jmp	0x3a1b13 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x5d3>
  3a2aef:      	xorps	xmm0, xmm0
  3a2af2:      	mov	ecx, r9d
  3a2af5:      	and	ecx, 0xffff0000
  3a2afb:      	cmp	ecx, 0xfe0000
  3a2b01:      	je	0x3a2aba <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x157a>
  3a2b03:      	add	r9d, 0xff03fff8
  3a2b0a:      	cmp	r9d, 0xa
  3a2b0e:      	jae	0x3a3647 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2107>
  3a2b14:      	mov	rdx, qword ptr [rbp - 0x50]
  3a2b18:      	mov	rsi, qword ptr [rdx + 0x80]
  3a2b1f:      	mov	edi, dword ptr [r15 + 0x4]
  3a2b23:      	cmp	rsi, rdi
  3a2b26:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a2b2c:      	imul	rcx, rdi, 0xe0
  3a2b33:      	add	rcx, qword ptr [rdx + 0x78]
  3a2b37:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2b3d:      	movups	xmm1, xmmword ptr [rcx + 0x98]
  3a2b44:      	movups	xmm2, xmmword ptr [rcx + 0x88]
  3a2b4b:      	movups	xmm3, xmmword ptr [rcx + 0x78]
  3a2b4f:      	movups	xmm4, xmmword ptr [rcx + 0x68]
  3a2b53:      	movups	xmm5, xmmword ptr [rcx + 0x58]
  3a2b57:      	movups	xmm6, xmmword ptr [rcx + 0x48]
  3a2b5b:      	movzx	edx, byte ptr [rcx + 0x40]
  3a2b5f:      	movups	xmm7, xmmword ptr [rcx + 0x30]
  3a2b63:      	movups	xmm8, xmmword ptr [rcx + 0x20]
  3a2b68:      	movups	xmm9, xmmword ptr [rcx + 0x10]
  3a2b6d:      	mov	rsi, qword ptr [rcx + 0x8]
  3a2b71:      	mov	ecx, dword ptr [rcx]
  3a2b73:      	mov	dword ptr [rbp - 0x208], ecx
  3a2b79:      	mov	qword ptr [rbp - 0x200], rsi
  3a2b80:      	movups	xmmword ptr [rbp - 0x1f8], xmm9
  3a2b88:      	movups	xmmword ptr [rbp - 0x1e8], xmm8
  3a2b90:      	movups	xmmword ptr [rbp - 0x1d8], xmm7
  3a2b97:      	mov	byte ptr [rbp - 0x1c8], dl
  3a2b9d:      	movups	xmmword ptr [rbp - 0x1c0], xmm6
  3a2ba4:      	movups	xmmword ptr [rbp - 0x1b0], xmm5
  3a2bab:      	movups	xmmword ptr [rbp - 0x1a0], xmm4
  3a2bb2:      	movups	xmmword ptr [rbp - 0x190], xmm3
  3a2bb9:      	movups	xmmword ptr [rbp - 0x180], xmm2
  3a2bc0:      	movups	xmmword ptr [rbp - 0x170], xmm1
  3a2bc7:      	mov	qword ptr [rbp - 0xe0], r10
  3a2bce:      	mov	qword ptr [rbp - 0xd8], rax
  3a2bd5:      	movaps	xmmword ptr [rbp - 0x160], xmm0
  3a2bdc:      	mov	ecx, 0x15
  3a2be1:      	mov	rdi, rsp
  3a2be4:      	lea	rsi, [rbp - 0x208]
  3a2beb:      	rep		movsq	qword ptr es:[rdi], qword ptr [rsi]
  3a2bee:      	movaps	xmm0, xmmword ptr [rbp - 0x160]
  3a2bf5:      	movups	xmmword ptr [rsp + 0xb8], xmm0
  3a2bfd:      	movups	xmm0, xmmword ptr [rbp - 0xe0]
  3a2c04:      	movups	xmmword ptr [rsp + 0xa8], xmm0
  3a2c0c:      	lea	rdi, [rbp - 0x108]
  3a2c13:      	mov	rsi, r11
  3a2c16:      	mov	rcx, r13
  3a2c19:      	mov	rdx, qword ptr [rbp - 0x58]
  3a2c1d:      	call	0x3a8fe0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).bulk>
  3a2c22:      	mov	rax, qword ptr [rbp - 0x108]
  3a2c29:      	mov	rdx, qword ptr [rbp - 0x100]
  3a2c30:      	movzx	r13d, byte ptr [rbp - 0xf8]
  3a2c38:      	mov	rsi, qword ptr [rbp - 0xf0]
  3a2c3f:      	test	rsi, rsi
  3a2c42:      	je	0x3a2db5 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1875>
  3a2c48:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a2c4d:      	mov	r11, qword ptr [rbp - 0x30]
  3a2c51:      	mov	r8d, 0xffffffff
  3a2c57:      	je	0x3a2533 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xff3>
  3a2c5d:      	jmp	0x3a34fd <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fbd>
  3a2c62:      	test	rax, rax
  3a2c65:      	je	0x3a2dba <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x187a>
  3a2c6b:      	mov	rax, qword ptr [rax]
  3a2c6e:      	test	rdi, rdi
  3a2c71:      	mov	r11, qword ptr [rbp - 0x30]
  3a2c75:      	mov	rsi, qword ptr [rbp - 0x40]
  3a2c79:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2c7f:      	mov	rcx, qword ptr [rdi + 0x38]
  3a2c83:      	test	rcx, rcx
  3a2c86:      	mov	edx, 0x10000
  3a2c8b:      	cmove	rcx, rdx
  3a2c8f:      	mov	rdx, rax
  3a2c92:      	or	rdx, rcx
  3a2c95:      	shr	rdx, 0x20
  3a2c99:      	je	0x3a2d85 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1845>
  3a2c9f:      	xor	edx, edx
  3a2ca1:      	div	rcx
  3a2ca4:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2ca9:      	mov	eax, 0x20
  3a2cae:      	jmp	0x3a2d90 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1850>
  3a2cb3:      	mov	rcx, qword ptr [rbp - 0x50]
  3a2cb7:      	mov	rsi, qword ptr [rcx + 0x80]
  3a2cbe:      	mov	edi, dword ptr [r15 + 0x4]
  3a2cc2:      	cmp	rsi, rdi
  3a2cc5:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a2ccb:      	imul	rax, rdi, 0xe0
  3a2cd2:      	add	rax, qword ptr [rcx + 0x78]
  3a2cd6:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2cdc:      	mov	rcx, qword ptr [rbp - 0x58]
  3a2ce0:      	mov	rsi, qword ptr [rcx + 0x78]
  3a2ce4:      	mov	edi, dword ptr [rax + 0x18]
  3a2ce7:      	cmp	rsi, rdi
  3a2cea:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a2cf0:      	shl	rdi, 0x3
  3a2cf4:      	add	rdi, qword ptr [rcx + 0x70]
  3a2cf8:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2cfe:      	mov	rsi, qword ptr [rdi]
  3a2d01:      	test	rsi, rsi
  3a2d04:      	je	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2d0a:      	mov	rcx, qword ptr [r11]
  3a2d0d:      	mov	rdx, qword ptr [r11 + 0x8]
  3a2d11:      	mov	r8, qword ptr [rcx + 0x8]
  3a2d15:      	mov	rcx, qword ptr [r11 + 0x10]
  3a2d19:      	mov	rsi, qword ptr [rsi + 0x10]
  3a2d1d:      	mov	edi, dword ptr [rax + 0x14]
  3a2d20:      	mov	eax, dword ptr [rax + 0x1c]
  3a2d23:      	shl	rax, 0x20
  3a2d27:      	or	rax, rdi
  3a2d2a:      	movlps	qword ptr [rsp + 0x18], xmm0
  3a2d2f:      	mov	qword ptr [rsp + 0x10], r10
  3a2d34:      	mov	qword ptr [rsp + 0x8], rax
  3a2d39:      	mov	qword ptr [rsp], r13
  3a2d3d:      	lea	rdi, [rbp - 0x130]
  3a2d44:      	call	0x3c4cf0 <github.com/cpunion/wasmvm/internal/engine/memory.(*Memory).Atomic>
  3a2d49:      	mov	rax, qword ptr [rbp - 0x130]
  3a2d50:      	mov	rdx, qword ptr [rbp - 0x128]
  3a2d57:      	movzx	r13d, byte ptr [rbp - 0x120]
  3a2d5f:      	mov	rsi, qword ptr [rbp - 0x118]
  3a2d66:      	test	rsi, rsi
  3a2d69:      	je	0x3a2dd4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1894>
  3a2d6b:      	cmp	qword ptr [rsi + 0x8], 0x0
  3a2d70:      	mov	r11, qword ptr [rbp - 0x30]
  3a2d74:      	mov	r8d, 0xffffffff
  3a2d7a:      	je	0x3a2533 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xff3>
  3a2d80:      	jmp	0x3a3506 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fc6>
  3a2d85:      	xor	edx, edx
  3a2d87:      	div	ecx
  3a2d89:      	nop	dword ptr [rax]
  3a2d90:      	xor	edx, edx
  3a2d92:      	mov	edi, dword ptr [r15 + 0x8]
  3a2d96:      	cmp	rdi, r8
  3a2d99:      	mov	r9, qword ptr [rbp - 0x48]
  3a2d9d:      	mov	r10, qword ptr [rbp - 0x60]
  3a2da1:      	je	0x3a164d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x10d>
  3a2da7:      	cmp	r9, rdi
  3a2daa:      	ja	0x3a1640 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x100>
  3a2db0:      	jmp	0x3a3517 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fd7>
  3a2db5:      	jmp	0x3a2529 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe9>
  3a2dba:      	mov	rax, qword ptr [rdi + 0x8]
  3a2dbe:      	test	rdi, rdi
  3a2dc1:      	mov	r11, qword ptr [rbp - 0x30]
  3a2dc5:      	mov	rsi, qword ptr [rbp - 0x40]
  3a2dc9:      	jne	0x3a2c7f <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x173f>
  3a2dcf:      	jmp	0x3a352a <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fea>
  3a2dd4:      	jmp	0x3a2529 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0xfe9>
  3a2dd9:      	mov	rax, qword ptr [rbp - 0x68]
  3a2ddd:      	mov	dword ptr [rax - 0x38], r14d
  3a2de1:      	mov	rdi, r11
  3a2de4:      	mov	r14, r11
  3a2de7:      	call	0x3a5e30 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runRuntime>
  3a2dec:      	test	al, al
  3a2dee:      	jne	0x3a34e8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa8>
  3a2df4:      	mov	rcx, qword ptr [r14 + 0x48]
  3a2df8:      	test	rcx, rcx
  3a2dfb:      	mov	r11, r14
  3a2dfe:      	mov	r8d, 0xffffffff
  3a2e04:      	jg	0x3a1585 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x45>
  3a2e0a:      	jmp	0x3a3638 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x20f8>
  3a2e0f:      	mov	rax, qword ptr [rbp - 0x68]
  3a2e13:      	mov	dword ptr [rax - 0x38], r14d
  3a2e17:      	mov	rdi, r11
  3a2e1a:      	mov	r14, r11
  3a2e1d:      	call	0x3a4920 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runCall>
  3a2e22:      	jmp	0x3a2dec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18ac>
  3a2e24:      	mov	rax, qword ptr [rbp - 0x68]
  3a2e28:      	mov	dword ptr [rax - 0x38], r14d
  3a2e2c:      	mov	rdi, r11
  3a2e2f:      	mov	r14, r11
  3a2e32:      	call	0x3a5830 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runThrow>
  3a2e37:      	jmp	0x3a2dec <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18ac>
  3a2e39:      	mov	rax, qword ptr [rbp - 0x68]
  3a2e3d:      	mov	dword ptr [rax - 0x38], r14d
  3a2e41:      	mov	rdi, r11
  3a2e44:      	mov	r14, r11
  3a2e47:      	call	0x3a44c0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).returnFrame>
  3a2e4c:      	test	al, al
  3a2e4e:      	je	0x3a2df4 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x18b4>
  3a2e50:      	mov	al, 0x1
  3a2e52:      	jmp	0x3a34e8 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa8>
  3a2e57:      	mov	edi, 0x28
  3a2e5c:      	call	0x5fa2d0 <GC_malloc@plt>
  3a2e61:      	test	rax, rax
  3a2e64:      	je	0x3a3656 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2116>
  3a2e6a:      	mov	r14, rax
  3a2e6d:      	add	r13, r12
  3a2e70:      	mov	edi, 0x28
  3a2e75:      	call	0x5e00a0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a2e7a:      	mov	rax, qword ptr [r13 + 0x8]
  3a2e7e:      	movaps	xmm0, xmmword ptr [rip + 0x30f9eb] # 0x6b2870 <github.com/cpunion/wasmvm/pkg/wasm.ErrLimit>
  3a2e85:      	movups	xmmword ptr [r14], xmm0
  3a2e89:      	mov	qword ptr [r14 + 0x10], rax
  3a2e8d:      	mov	qword ptr [r14 + 0x18], 0x29a088
  3a2e95:      	mov	qword ptr [r14 + 0x20], 0xe
  3a2e9d:      	mov	r15, qword ptr [rbp - 0x30]
  3a2ea1:      	mov	qword ptr [r15 + 0x140], 0x2baa88
  3a2eac:      	mov	qword ptr [r15 + 0x148], r14
  3a2eb3:      	mov	byte ptr [r15 + 0x138], 0x3
  3a2ebb:      	lea	rdi, [r15 + 0x168]
  3a2ec2:      	mov	edx, 0x158
  3a2ec7:      	xor	esi, esi
  3a2ec9:      	call	0x5fa0a0 <memset@plt>
  3a2ece:      	mov	rdi, qword ptr [r15 + 0x40]
  3a2ed2:      	mov	rdx, qword ptr [r15 + 0x48]
  3a2ed6:      	shl	rdx, 0x6
  3a2eda:      	xor	esi, esi
  3a2edc:      	call	0x5fa0a0 <memset@plt>
  3a2ee1:      	mov	qword ptr [r15 + 0x48], 0x0
  3a2ee9:      	mov	rdi, qword ptr [r15 + 0x58]
  3a2eed:      	mov	rax, qword ptr [r15 + 0x60]
  3a2ef1:      	shl	rax, 0x4
  3a2ef5:      	lea	rdx, [rax + 2*rax]
  3a2ef9:      	xor	esi, esi
  3a2efb:      	call	0x5fa0a0 <memset@plt>
  3a2f00:      	mov	qword ptr [r15 + 0x60], 0x0
  3a2f08:      	mov	qword ptr [r15 + 0x70], 0x0
  3a2f10:      	mov	qword ptr [r15 + 0x80], 0x0
  3a2f1b:      	mov	qword ptr [r15 + 0x118], 0x0
  3a2f26:      	xorps	xmm0, xmm0
  3a2f29:      	movups	xmmword ptr [r15 + 0xd8], xmm0
  3a2f31:      	movups	xmmword ptr [r15 + 0xe8], xmm0
  3a2f39:      	mov	qword ptr [r15 + 0xf8], 0x0
  3a2f44:      	mov	rdi, qword ptr [r15 + 0x100]
  3a2f4b:      	mov	rdx, qword ptr [r15 + 0x108]
  3a2f52:      	shl	rdx, 0x4
  3a2f56:      	xor	esi, esi
  3a2f58:      	call	0x5fa0a0 <memset@plt>
  3a2f5d:      	cmp	qword ptr [r15 + 0x110], 0x41
  3a2f65:      	jl	0x3a2f7d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1a3d>
  3a2f67:      	mov	qword ptr [r15 + 0x100], 0x0
  3a2f72:      	mov	qword ptr [r15 + 0x110], 0x0
  3a2f7d:      	mov	qword ptr [r15 + 0x108], 0x0
  3a2f88:      	xorps	xmm0, xmm0
  3a2f8b:      	movups	xmmword ptr [r15 + 0x120], xmm0
  3a2f93:      	mov	qword ptr [r15 + 0x130], 0x0
  3a2f9e:      	mov	rdi, qword ptr [r15 + 0xa8]
  3a2fa5:      	mov	rdx, qword ptr [r15 + 0xb0]
  3a2fac:      	shl	rdx, 0x4
  3a2fb0:      	xor	esi, esi
  3a2fb2:      	call	0x5fa0a0 <memset@plt>
  3a2fb7:      	mov	rax, qword ptr [rbp - 0x30]
  3a2fbb:      	mov	qword ptr [rax + 0xb0], 0x0
  3a2fc6:      	mov	rax, qword ptr [rbp - 0x30]
  3a2fca:      	mov	rdi, qword ptr [rax + 0xc0]
  3a2fd1:      	mov	rax, qword ptr [rbp - 0x30]
  3a2fd5:      	mov	rdx, qword ptr [rax + 0xc8]
  3a2fdc:      	jmp	0x3a34ac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f6c>
  3a2fe1:      	mov	qword ptr [r11 + 0x140], rax
  3a2fe8:      	mov	qword ptr [r11 + 0x148], rdx
  3a2fef:      	mov	byte ptr [r11 + 0x138], 0x3
  3a2ff7:      	lea	rdi, [r11 + 0x168]
  3a2ffe:      	mov	edx, 0x158
  3a3003:      	xor	esi, esi
  3a3005:      	call	0x5fa0a0 <memset@plt>
  3a300a:      	mov	rax, qword ptr [rbp - 0x30]
  3a300e:      	mov	rdi, qword ptr [rax + 0x40]
  3a3012:      	mov	rax, qword ptr [rbp - 0x30]
  3a3016:      	mov	rdx, qword ptr [rax + 0x48]
  3a301a:      	shl	rdx, 0x6
  3a301e:      	xor	esi, esi
  3a3020:      	call	0x5fa0a0 <memset@plt>
  3a3025:      	mov	rax, qword ptr [rbp - 0x30]
  3a3029:      	mov	qword ptr [rax + 0x48], 0x0
  3a3031:      	mov	rax, qword ptr [rbp - 0x30]
  3a3035:      	mov	rdi, qword ptr [rax + 0x58]
  3a3039:      	mov	rax, qword ptr [rbp - 0x30]
  3a303d:      	mov	rax, qword ptr [rax + 0x60]
  3a3041:      	shl	rax, 0x4
  3a3045:      	lea	rdx, [rax + 2*rax]
  3a3049:      	xor	esi, esi
  3a304b:      	call	0x5fa0a0 <memset@plt>
  3a3050:      	mov	rax, qword ptr [rbp - 0x30]
  3a3054:      	mov	qword ptr [rax + 0x60], 0x0
  3a305c:      	mov	rax, qword ptr [rbp - 0x30]
  3a3060:      	mov	qword ptr [rax + 0x70], 0x0
  3a3068:      	mov	rax, qword ptr [rbp - 0x30]
  3a306c:      	mov	qword ptr [rax + 0x80], 0x0
  3a3077:      	mov	rax, qword ptr [rbp - 0x30]
  3a307b:      	mov	qword ptr [rax + 0x118], 0x0
  3a3086:      	xorps	xmm0, xmm0
  3a3089:      	mov	rax, qword ptr [rbp - 0x30]
  3a308d:      	movups	xmmword ptr [rax + 0xd8], xmm0
  3a3094:      	mov	rax, qword ptr [rbp - 0x30]
  3a3098:      	movups	xmmword ptr [rax + 0xe8], xmm0
  3a309f:      	mov	rax, qword ptr [rbp - 0x30]
  3a30a3:      	mov	qword ptr [rax + 0xf8], 0x0
  3a30ae:      	mov	rax, qword ptr [rbp - 0x30]
  3a30b2:      	mov	rdi, qword ptr [rax + 0x100]
  3a30b9:      	mov	rax, qword ptr [rbp - 0x30]
  3a30bd:      	mov	rdx, qword ptr [rax + 0x108]
  3a30c4:      	shl	rdx, 0x4
  3a30c8:      	xor	esi, esi
  3a30ca:      	call	0x5fa0a0 <memset@plt>
  3a30cf:      	mov	rax, qword ptr [rbp - 0x30]
  3a30d3:      	cmp	qword ptr [rax + 0x110], 0x41
  3a30db:      	jl	0x3a30f3 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1bb3>
  3a30dd:      	mov	qword ptr [rax + 0x100], 0x0
  3a30e8:      	mov	qword ptr [rax + 0x110], 0x0
  3a30f3:      	mov	qword ptr [rax + 0x108], 0x0
  3a30fe:      	xorps	xmm0, xmm0
  3a3101:      	movups	xmmword ptr [rax + 0x120], xmm0
  3a3108:      	mov	qword ptr [rax + 0x130], 0x0
  3a3113:      	mov	rdi, qword ptr [rax + 0xa8]
  3a311a:      	mov	rdx, qword ptr [rax + 0xb0]
  3a3121:      	shl	rdx, 0x4
  3a3125:      	xor	esi, esi
  3a3127:      	call	0x5fa0a0 <memset@plt>
  3a312c:      	mov	rax, qword ptr [rbp - 0x30]
  3a3130:      	mov	qword ptr [rax + 0xb0], 0x0
  3a313b:      	mov	rax, qword ptr [rbp - 0x30]
  3a313f:      	mov	rdi, qword ptr [rax + 0xc0]
  3a3146:      	mov	rax, qword ptr [rbp - 0x30]
  3a314a:      	mov	rdx, qword ptr [rax + 0xc8]
  3a3151:      	jmp	0x3a34ac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f6c>
  3a3156:      	mov	r13d, edx
  3a3159:      	jmp	0x3a3309 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc9>
  3a315e:      	mov	r14, qword ptr [rbp - 0x138]
  3a3165:      	mov	edi, 0x8
  3a316a:      	call	0x5fa2d0 <GC_malloc@plt>
  3a316f:      	test	rax, rax
  3a3172:      	je	0x3a3656 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2116>
  3a3178:      	mov	r15, rax
  3a317b:      	mov	edi, 0x8
  3a3180:      	call	0x5e00a0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a3185:      	mov	edx, 0x299ff8
  3a318a:      	mov	rdi, r13
  3a318d:      	mov	rsi, r14
  3a3190:      	mov	rcx, r15
  3a3193:      	call	0x5b70e0 <errors.As>
  3a3198:      	test	al, 0x1
  3a319a:      	je	0x3a31ac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1c6c>
  3a319c:      	mov	rax, qword ptr [r15]
  3a319f:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a31a6:      	mov	ecx, dword ptr [rcx + 0x50]
  3a31a9:      	mov	dword ptr [rax + 0x4], ecx
  3a31ac:      	mov	r15, qword ptr [rbp - 0x30]
  3a31b0:      	mov	qword ptr [r15 + 0x140], r13
  3a31b7:      	mov	qword ptr [r15 + 0x148], r14
  3a31be:      	mov	byte ptr [r15 + 0x138], 0x3
  3a31c6:      	lea	rdi, [r15 + 0x168]
  3a31cd:      	mov	edx, 0x158
  3a31d2:      	xor	esi, esi
  3a31d4:      	call	0x5fa0a0 <memset@plt>
  3a31d9:      	mov	rdi, qword ptr [r15 + 0x40]
  3a31dd:      	mov	rdx, qword ptr [r15 + 0x48]
  3a31e1:      	shl	rdx, 0x6
  3a31e5:      	xor	esi, esi
  3a31e7:      	call	0x5fa0a0 <memset@plt>
  3a31ec:      	mov	qword ptr [r15 + 0x48], 0x0
  3a31f4:      	mov	rdi, qword ptr [r15 + 0x58]
  3a31f8:      	mov	rax, qword ptr [r15 + 0x60]
  3a31fc:      	shl	rax, 0x4
  3a3200:      	lea	rdx, [rax + 2*rax]
  3a3204:      	xor	esi, esi
  3a3206:      	call	0x5fa0a0 <memset@plt>
  3a320b:      	mov	qword ptr [r15 + 0x60], 0x0
  3a3213:      	mov	qword ptr [r15 + 0x70], 0x0
  3a321b:      	mov	qword ptr [r15 + 0x80], 0x0
  3a3226:      	mov	qword ptr [r15 + 0x118], 0x0
  3a3231:      	xorps	xmm0, xmm0
  3a3234:      	movups	xmmword ptr [r15 + 0xd8], xmm0
  3a323c:      	movups	xmmword ptr [r15 + 0xe8], xmm0
  3a3244:      	mov	qword ptr [r15 + 0xf8], 0x0
  3a324f:      	mov	rdi, qword ptr [r15 + 0x100]
  3a3256:      	mov	rdx, qword ptr [r15 + 0x108]
  3a325d:      	shl	rdx, 0x4
  3a3261:      	xor	esi, esi
  3a3263:      	call	0x5fa0a0 <memset@plt>
  3a3268:      	cmp	qword ptr [r15 + 0x110], 0x41
  3a3270:      	jl	0x3a3288 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1d48>
  3a3272:      	mov	qword ptr [r15 + 0x100], 0x0
  3a327d:      	mov	qword ptr [r15 + 0x110], 0x0
  3a3288:      	mov	qword ptr [r15 + 0x108], 0x0
  3a3293:      	xorps	xmm0, xmm0
  3a3296:      	movups	xmmword ptr [r15 + 0x120], xmm0
  3a329e:      	mov	qword ptr [r15 + 0x130], 0x0
  3a32a9:      	mov	rdi, qword ptr [r15 + 0xa8]
  3a32b0:      	mov	rdx, qword ptr [r15 + 0xb0]
  3a32b7:      	shl	rdx, 0x4
  3a32bb:      	xor	esi, esi
  3a32bd:      	call	0x5fa0a0 <memset@plt>
  3a32c2:      	mov	rax, qword ptr [rbp - 0x30]
  3a32c6:      	mov	qword ptr [rax + 0xb0], 0x0
  3a32d1:      	mov	rax, qword ptr [rbp - 0x30]
  3a32d5:      	mov	rdi, qword ptr [rax + 0xc0]
  3a32dc:      	mov	rax, qword ptr [rbp - 0x30]
  3a32e0:      	mov	rdx, qword ptr [rax + 0xc8]
  3a32e7:      	jmp	0x3a34ac <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f6c>
  3a32ec:      	mov	r13b, 0x3
  3a32ef:      	jmp	0x3a3309 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc9>
  3a32f1:      	mov	rdi, r11
  3a32f4:      	mov	rsi, rax
  3a32f7:      	call	0x3a0bb0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a32fc:      	jmp	0x3a34e6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa6>
  3a3301:      	mov	r13b, 0x1
  3a3304:      	jmp	0x3a3309 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1dc9>
  3a3306:      	mov	r13b, 0x9
  3a3309:      	mov	edi, 0x10
  3a330e:      	call	0x5fa380 <GC_malloc_atomic@plt>
  3a3313:      	test	rax, rax
  3a3316:      	je	0x3a3656 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x2116>
  3a331c:      	mov	r14, rax
  3a331f:      	mov	edi, 0x10
  3a3324:      	call	0x5e00a0 <github.com/xgo-dev/llgo/runtime/internal/runtime.recordMemProfileAlloc>
  3a3329:      	xorps	xmm0, xmm0
  3a332c:      	movups	xmmword ptr [r14], xmm0
  3a3330:      	mov	rax, qword ptr [rbp - 0x50]
  3a3334:      	mov	rsi, qword ptr [rax + 0x80]
  3a333b:      	mov	edi, dword ptr [r15 + 0x4]
  3a333f:      	cmp	rsi, rdi
  3a3342:      	jbe	0x3a3534 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1ff4>
  3a3348:      	mov	rax, qword ptr [rax + 0x78]
  3a334c:      	mov	rcx, qword ptr [rbp - 0xa8]
  3a3353:      	mov	ecx, dword ptr [rcx + 0x50]
  3a3356:      	imul	rdx, rdi, 0xe0
  3a335d:      	mov	rax, qword ptr [rax + rdx + 0x8]
  3a3362:      	mov	byte ptr [r14], r13b
  3a3365:      	mov	dword ptr [r14 + 0x4], ecx
  3a3369:      	mov	qword ptr [r14 + 0x8], rax
  3a336d:      	mov	r15, qword ptr [rbp - 0x30]
  3a3371:      	mov	qword ptr [r15 + 0x140], 0x29a410
  3a337c:      	mov	qword ptr [r15 + 0x148], r14
  3a3383:      	mov	byte ptr [r15 + 0x138], 0x3
  3a338b:      	lea	rdi, [r15 + 0x168]
  3a3392:      	mov	edx, 0x158
  3a3397:      	xor	esi, esi
  3a3399:      	call	0x5fa0a0 <memset@plt>
  3a339e:      	mov	rdi, qword ptr [r15 + 0x40]
  3a33a2:      	mov	rdx, qword ptr [r15 + 0x48]
  3a33a6:      	shl	rdx, 0x6
  3a33aa:      	xor	esi, esi
  3a33ac:      	call	0x5fa0a0 <memset@plt>
  3a33b1:      	mov	qword ptr [r15 + 0x48], 0x0
  3a33b9:      	mov	rdi, qword ptr [r15 + 0x58]
  3a33bd:      	mov	rax, qword ptr [r15 + 0x60]
  3a33c1:      	shl	rax, 0x4
  3a33c5:      	lea	rdx, [rax + 2*rax]
  3a33c9:      	xor	esi, esi
  3a33cb:      	call	0x5fa0a0 <memset@plt>
  3a33d0:      	mov	qword ptr [r15 + 0x60], 0x0
  3a33d8:      	mov	qword ptr [r15 + 0x70], 0x0
  3a33e0:      	mov	qword ptr [r15 + 0x80], 0x0
  3a33eb:      	mov	qword ptr [r15 + 0x118], 0x0
  3a33f6:      	xorps	xmm0, xmm0
  3a33f9:      	movups	xmmword ptr [r15 + 0xd8], xmm0
  3a3401:      	movups	xmmword ptr [r15 + 0xe8], xmm0
  3a3409:      	mov	qword ptr [r15 + 0xf8], 0x0
  3a3414:      	mov	rdi, qword ptr [r15 + 0x100]
  3a341b:      	mov	rdx, qword ptr [r15 + 0x108]
  3a3422:      	shl	rdx, 0x4
  3a3426:      	xor	esi, esi
  3a3428:      	call	0x5fa0a0 <memset@plt>
  3a342d:      	cmp	qword ptr [r15 + 0x110], 0x41
  3a3435:      	jl	0x3a344d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f0d>
  3a3437:      	mov	qword ptr [r15 + 0x100], 0x0
  3a3442:      	mov	qword ptr [r15 + 0x110], 0x0
  3a344d:      	mov	qword ptr [r15 + 0x108], 0x0
  3a3458:      	xorps	xmm0, xmm0
  3a345b:      	movups	xmmword ptr [r15 + 0x120], xmm0
  3a3463:      	mov	qword ptr [r15 + 0x130], 0x0
  3a346e:      	mov	rdi, qword ptr [r15 + 0xa8]
  3a3475:      	mov	rdx, qword ptr [r15 + 0xb0]
  3a347c:      	shl	rdx, 0x4
  3a3480:      	xor	esi, esi
  3a3482:      	call	0x5fa0a0 <memset@plt>
  3a3487:      	mov	rax, qword ptr [rbp - 0x30]
  3a348b:      	mov	qword ptr [rax + 0xb0], 0x0
  3a3496:      	mov	rax, qword ptr [rbp - 0x30]
  3a349a:      	mov	rdi, qword ptr [rax + 0xc0]
  3a34a1:      	mov	rax, qword ptr [rbp - 0x30]
  3a34a5:      	mov	rdx, qword ptr [rax + 0xc8]
  3a34ac:      	shl	rdx, 0x4
  3a34b0:      	xor	esi, esi
  3a34b2:      	call	0x5fa0a0 <memset@plt>
  3a34b7:      	mov	rax, qword ptr [rbp - 0x30]
  3a34bb:      	cmp	qword ptr [rax + 0xd0], 0x41
  3a34c3:      	jl	0x3a34db <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1f9b>
  3a34c5:      	mov	qword ptr [rax + 0xc0], 0x0
  3a34d0:      	mov	qword ptr [rax + 0xd0], 0x0
  3a34db:      	mov	qword ptr [rax + 0xc8], 0x0
  3a34e6:      	mov	al, 0x3
  3a34e8:      	mov	rdx, rbx
  3a34eb:      	add	rsp, 0x2a8
  3a34f2:      	pop	rbx
  3a34f3:      	pop	r12
  3a34f5:      	pop	r13
  3a34f7:      	pop	r14
  3a34f9:      	pop	r15
  3a34fb:      	pop	rbp
  3a34fc:      	ret
  3a34fd:      	mov	rdx, qword ptr [rbp - 0xe8]
  3a3504:      	jmp	0x3a350d <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fcd>
  3a3506:      	mov	rdx, qword ptr [rbp - 0x110]
  3a350d:      	mov	rdi, r11
  3a3510:      	call	0x3a0bb0 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).fail>
  3a3515:      	jmp	0x3a34e6 <github.com/cpunion/wasmvm/internal/engine/execute.(*Execution).runSlotsLoop+0x1fa6>
  3a3517:      	mov	rsi, r9
  3a351a:      	call	0x5e0180 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a351f:      	mov	rdi, r14
  3a3522:      	mov	rsi, r10
  3a3525:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a352a:      	mov	edi, 0x1
  3a352f:      	call	0x5e4410 <github.com/xgo-dev/llgo/runtime/internal/runtime.AssertNilDeref>
  3a3534:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3539:      	mov	rdi, r15
  3a353c:      	mov	rsi, r15
  3a353f:      	call	0x5e0180 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3544:      	mov	rdi, rax
  3a3547:      	mov	rsi, r9
  3a354a:      	call	0x5e0180 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a354f:      	mov	rsi, qword ptr [rbp - 0x48]
  3a3553:      	call	0x5e0180 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3558:      	mov	rsi, r9
  3a355b:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3560:      	mov	rdi, r9
  3a3563:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3568:      	mov	rdi, rax
  3a356b:      	mov	rsi, qword ptr [rbp - 0x48]
  3a356f:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3574:      	mov	rdi, rcx
  3a3577:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a357c:      	mov	rdi, rsi
  3a357f:      	mov	rsi, r9
  3a3582:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a3587:      	mov	edi, 0x1
  3a358c:      	call	0x5e0180 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3591:      	xor	edi, edi
  3a3593:      	call	0x5e0180 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a3598:      	mov	rdx, rsi
  3a359b:      	xor	esi, esi
  3a359d:      	mov	ecx, 0x3
  3a35a2:      	call	0x5eb010 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a35a7:      	mov	rdi, rsi
  3a35aa:      	xor	esi, esi
  3a35ac:      	mov	ecx, 0x2
  3a35b1:      	call	0x5eb010 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a35b6:      	mov	rdi, rax
  3a35b9:      	mov	esi, 0x1
  3a35be:      	mov	ecx, 0x2
  3a35c3:      	call	0x5eb010 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a35c8:      	mov	esi, 0x1
  3a35cd:      	mov	rdx, rax
  3a35d0:      	mov	ecx, 0x3
  3a35d5:      	call	0x5eb010 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a35da:      	mov	r10, qword ptr [rbp - 0x60]
  3a35de:      	mov	rdi, r15
  3a35e1:      	mov	rsi, r10
  3a35e4:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a35e9:      	mov	rsi, qword ptr [rbp - 0x48]
  3a35ed:      	call	0x5e16b0 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndexU>
  3a35f2:      	mov	edi, 0x3
  3a35f7:      	call	0x5e0180 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a35fc:      	mov	edi, 0x2
  3a3601:      	mov	esi, 0x2
  3a3606:      	call	0x5e0180 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicIndex>
  3a360b:      	xor	esi, esi
  3a360d:      	mov	rdx, r15
  3a3610:      	mov	ecx, 0x3
  3a3615:      	call	0x5eb010 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a361a:      	mov	rdi, r15
  3a361d:      	xor	esi, esi
  3a361f:      	mov	ecx, 0x2
  3a3624:      	call	0x5eb010 <github.com/xgo-dev/llgo/runtime/internal/runtime.PanicBounds>
  3a3629:      	mov	edi, 0x329530
  3a362e:      	mov	esi, 0x329860
  3a3633:      	call	0x5dfdc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3638:      	mov	edi, 0x3339e8
  3a363d:      	mov	esi, 0x29a078
  3a3642:      	call	0x5dfdc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3647:      	mov	edi, 0x3339e8
  3a364c:      	mov	esi, 0x29a098
  3a3651:      	call	0x5dfdc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3656:      	mov	edi, 0x3339e8
  3a365b:      	mov	esi, 0x329508
  3a3660:      	call	0x5dfdc0 <github.com/xgo-dev/llgo/runtime/internal/runtime.Panic>
  3a3665:      	int3
  3a3666:      	int3
  3a3667:      	int3
  3a3668:      	int3
  3a3669:      	int3
  3a366a:      	int3
  3a366b:      	int3
  3a366c:      	int3
  3a366d:      	int3
  3a366e:      	int3
  3a366f:      	int3
