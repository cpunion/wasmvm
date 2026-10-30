package preview2

import (
	"net/netip"
	"testing"

	"github.com/cpunion/wasmvm/pkg/component"
)

func TestSocketAddressRoundTrip(t *testing.T) {
	for _, input := range []SocketAddress{
		{Address: netip.MustParseAddrPort("0.0.0.0:0")},
		{Address: netip.MustParseAddrPort("127.0.0.1:80")},
		{Address: netip.MustParseAddrPort("255.255.255.255:65535")},
		{Address: netip.MustParseAddrPort("[::]:0")},
		{Address: netip.MustParseAddrPort("[2001:db8::1]:443")},
		{Address: netip.MustParseAddrPort("[fe80::a:b:c:d]:65535"), FlowInfo: 0xffffffff, ScopeID: 0xffffffff},
		{Address: netip.MustParseAddrPort("[::ffff:127.0.0.1]:10")},
	} {
		value, err := socketAddressValue(input)
		if err != nil {
			t.Fatal(input, err)
		}
		out, err := parseSocketAddress(value)
		if err != nil || out != input {
			t.Fatal(input, out, err)
		}
		if input.Address.Addr().Is4() && value.Case != 0 || input.Address.Addr().Is6() && value.Case != 1 {
			t.Fatal(input, value)
		}
	}
}
func TestSocketAddressInvalidValues(t *testing.T) {
	for _, input := range []SocketAddress{{}, {Address: netip.MustParseAddrPort("127.0.0.1:80"), FlowInfo: 1}, {Address: netip.MustParseAddrPort("127.0.0.1:80"), ScopeID: 1}, {Address: netip.MustParseAddrPort("[fe80::1%en0]:80")}} {
		if _, err := socketAddressValue(input); networkError(err) != NetworkInvalidArgument {
			t.Fatal(input, err)
		}
	}
	good4, _ := socketAddressValue(SocketAddress{Address: netip.MustParseAddrPort("127.0.0.1:80")})
	good6, _ := socketAddressValue(SocketAddress{Address: netip.MustParseAddrPort("[::1]:80")})
	for _, mutate := range []func(*component.Value){
		func(v *component.Value) { v.Case = 2 },
		func(v *component.Value) { v.Elements = nil },
		func(v *component.Value) { v.Elements[0].Elements = nil },
		func(v *component.Value) { v.Elements[0].Elements[0].Bits = 65536 },
		func(v *component.Value) { v.Elements[0].Elements[1].Elements = nil },
		func(v *component.Value) { v.Elements[0].Elements[1].Elements[0].Bits = 256 },
	} {
		copy := component.Value{Case: good4.Case, Elements: []component.Value{{Elements: []component.Value{{Bits: 80}, {Elements: []component.Value{{Bits: 127}, {}, {}, {Bits: 1}}}}}}}
		mutate(&copy)
		if _, err := parseSocketAddress(copy); networkError(err) != NetworkInvalidArgument {
			t.Fatal(copy, err)
		}
	}
	for _, index := range []int{1, 3} {
		value, _ := socketAddressValue(SocketAddress{Address: netip.MustParseAddrPort("[::1]:80")})
		value.Elements[0].Elements[index].Bits = 1 << 32
		if _, err := parseSocketAddress(value); networkError(err) != NetworkInvalidArgument {
			t.Fatal(value, err)
		}
	}
	good6.Elements[0].Elements[2].Elements[0].Bits = 65536
	if _, err := parseSocketAddress(good6); networkError(err) != NetworkInvalidArgument {
		t.Fatal(err)
	}
}
