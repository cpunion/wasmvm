//go:build linux || darwin

package preview1

// Errno is the wasi_snapshot_preview1 error namespace.
type Errno uint16

const (
	ErrSuccess        Errno = 0
	Err2Big           Errno = 1
	ErrAcces          Errno = 2
	ErrAddrinuse      Errno = 3
	ErrAddrnotavail   Errno = 4
	ErrAfnosupport    Errno = 5
	ErrAgain          Errno = 6
	ErrAlready        Errno = 7
	ErrBadf           Errno = 8
	ErrBadmsg         Errno = 9
	ErrBusy           Errno = 10
	ErrCanceled       Errno = 11
	ErrChild          Errno = 12
	ErrConnaborted    Errno = 13
	ErrConnrefused    Errno = 14
	ErrConnreset      Errno = 15
	ErrDeadlk         Errno = 16
	ErrDestaddrreq    Errno = 17
	ErrDom            Errno = 18
	ErrDquot          Errno = 19
	ErrExist          Errno = 20
	ErrFault          Errno = 21
	ErrFbig           Errno = 22
	ErrHostunreach    Errno = 23
	ErrIdrm           Errno = 24
	ErrIlseq          Errno = 25
	ErrInprogress     Errno = 26
	ErrIntr           Errno = 27
	ErrInval          Errno = 28
	ErrIo             Errno = 29
	ErrIsconn         Errno = 30
	ErrIsdir          Errno = 31
	ErrLoop           Errno = 32
	ErrMfile          Errno = 33
	ErrMlink          Errno = 34
	ErrMsgsize        Errno = 35
	ErrMultihop       Errno = 36
	ErrNametoolong    Errno = 37
	ErrNetdown        Errno = 38
	ErrNetreset       Errno = 39
	ErrNetunreach     Errno = 40
	ErrNfile          Errno = 41
	ErrNobufs         Errno = 42
	ErrNodev          Errno = 43
	ErrNoent          Errno = 44
	ErrNoexec         Errno = 45
	ErrNolck          Errno = 46
	ErrNolink         Errno = 47
	ErrNomem          Errno = 48
	ErrNomsg          Errno = 49
	ErrNoprotoopt     Errno = 50
	ErrNospc          Errno = 51
	ErrNosys          Errno = 52
	ErrNotconn        Errno = 53
	ErrNotdir         Errno = 54
	ErrNotempty       Errno = 55
	ErrNotrecoverable Errno = 56
	ErrNotsock        Errno = 57
	ErrNotsup         Errno = 58
	ErrNotty          Errno = 59
	ErrNxio           Errno = 60
	ErrOverflow       Errno = 61
	ErrOwnerdead      Errno = 62
	ErrPerm           Errno = 63
	ErrPipe           Errno = 64
	ErrProto          Errno = 65
	ErrProtonosupport Errno = 66
	ErrPrototype      Errno = 67
	ErrRange          Errno = 68
	ErrRofs           Errno = 69
	ErrSpipe          Errno = 70
	ErrSrch           Errno = 71
	ErrStale          Errno = 72
	ErrTimedout       Errno = 73
	ErrTxtbsy         Errno = 74
	ErrXdev           Errno = 75
	ErrNotcapable     Errno = 76
)

type Rights uint64

const (
	RightFdDatasync           Rights = 1 << 0
	RightFdRead               Rights = 1 << 1
	RightFdSeek               Rights = 1 << 2
	RightFdFdstatSetFlags     Rights = 1 << 3
	RightFdSync               Rights = 1 << 4
	RightFdTell               Rights = 1 << 5
	RightFdWrite              Rights = 1 << 6
	RightFdAdvise             Rights = 1 << 7
	RightFdAllocate           Rights = 1 << 8
	RightPathCreateDirectory  Rights = 1 << 9
	RightPathCreateFile       Rights = 1 << 10
	RightPathLinkSource       Rights = 1 << 11
	RightPathLinkTarget       Rights = 1 << 12
	RightPathOpen             Rights = 1 << 13
	RightFdReaddir            Rights = 1 << 14
	RightPathReadlink         Rights = 1 << 15
	RightPathRenameSource     Rights = 1 << 16
	RightPathRenameTarget     Rights = 1 << 17
	RightPathFilestatGet      Rights = 1 << 18
	RightPathFilestatSetSize  Rights = 1 << 19
	RightPathFilestatSetTimes Rights = 1 << 20
	RightFdFilestatGet        Rights = 1 << 21
	RightFdFilestatSetSize    Rights = 1 << 22
	RightFdFilestatSetTimes   Rights = 1 << 23
	RightPathSymlink          Rights = 1 << 24
	RightPathRemoveDirectory  Rights = 1 << 25
	RightPathUnlinkFile       Rights = 1 << 26
	RightPollFdReadwrite      Rights = 1 << 27
	RightSockShutdown         Rights = 1 << 28
	RightSockAccept           Rights = 1 << 29
)
