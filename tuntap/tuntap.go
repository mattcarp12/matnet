package tuntap

import (
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// Interface is a TUN/TAP interface.
//
// MultiQueue(Linux kernel > 3.8): With MultiQueue enabled, user should hold multiple
// interfaces to send/receive packet in parallel.
// Kernel document about MultiQueue: https://www.kernel.org/doc/Documentation/networking/tuntap.txt
type Interface struct {
	DeviceType DeviceType
	io.ReadWriteCloser
	name string
}

// DeviceType is the type for specifying device types.
type DeviceType int

// TUN and TAP device types.
const (
	_ = iota
	TUN
	TAP
)

// Config defines parameters required to create a TUN/TAP interface. It's only
// used when the device is initialized. A zero-value Config is a valid
// configuration.
type Config struct {
	// DeviceType specifies whether the device is a TUN or TAP interface. A
	// zero-value is treated as TUN.
	DeviceType DeviceType

	// PlatformSpecificParams defines parameters that differ on different
	// platforms. See comments for the type for more details.
	PlatformSpecificParams
}

// New creates a new TUN/TAP interface using config.
func New(config Config) (*Interface, error) {
	var zeroConfig Config
	if config == zeroConfig {
		config = Config{
			DeviceType:             TUN,
			PlatformSpecificParams: PlatformSpecificParams{},
		}
	}

	switch config.DeviceType {
	case TUN, TAP:
		return openDev(config)
	default:
		return nil, errors.New("unknown device type")
	}
}

// Name returns the interface name of ifce, e.g. tun0, tap1, tun0, etc..
func (iface *Interface) Name() string {
	return iface.name
}

// DevicePermissions determines the owner and group owner for the newly created
// interface.
type DevicePermissions struct {
	// Owner is the ID of the user which will be granted ownership of the
	// device.  If set to a negative value, the owner value will not be
	// changed.  By default, Linux sets the owner to -1, which allows any user.
	Owner uint

	// Group is the ID of the group which will be granted access to the device.
	// If set to a negative value, the group value will not be changed.  By
	// default, Linux sets the group to -1, which allows any group.
	Group uint
}

// PlatformSpecificParams defines parameters in Config that are specific to
// Linux. A zero-value of such type is valid, yielding an interface
// with OS defined name.
type PlatformSpecificParams struct {
	// Persist specifies whether persistence mode for the interface device
	// should be enabled or disabled.
	Persist bool

	// MultiQueue specifies whether the multiqueue flag should be set on the
	// interface.  From version 3.8, Linux supports multiqueue tuntap which can
	// uses multiple file descriptors (queues) to parallelize packets sending
	// or receiving.
	MultiQueue bool

	// Permissions, if non-nil, specifies the owner and group owner for the
	// interface.  A zero-value of this field, i.e. nil, indicates that no
	// changes to owner or group will be made.
	Permissions *DevicePermissions

	// Name is the name to be set for the interface to be created. This overrides
	// the default name assigned by OS such as tap0 or tun0. A zero-value of this
	// field, i.e. an empty string, indicates that the default name should be
	// used.
	Name string
}

func openDev(config Config) (*Interface, error) {
	fdInt, err := syscall.Open("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	name, err := setupFd(config, uintptr(fdInt))
	if err != nil {
		return nil, err
	}

	return &Interface{
		DeviceType:      config.DeviceType,
		ReadWriteCloser: os.NewFile(uintptr(fdInt), "tun"),
		name:            name,
	}, nil
}

const (
	cIFFTUN        = 0x0001
	cIFFTAP        = 0x0002
	cIFFNOPI       = 0x1000
	cIFFMULTIQUEUE = 0x0100
)

// See <linux/if.h>
type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func ioctl(fd uintptr, request uintptr, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}

	return nil
}

func setupFd(config Config, fd uintptr) (string, error) {
	var flags uint16 = cIFFNOPI
	if config.DeviceType == TUN {
		flags |= cIFFTUN
	} else {
		flags |= cIFFTAP
	}

	if config.PlatformSpecificParams.MultiQueue {
		flags |= cIFFMULTIQUEUE
	}

	name, err := createInterface(fd, config.Name, flags)
	if err != nil {
		return "", err
	}

	if err = setDeviceOptions(fd, config); err != nil {
		return "", err
	}

	return name, nil
}

func createInterface(fd uintptr, ifName string, flags uint16) (string, error) {
	var req ifReq
	req.Flags = flags
	copy(req.Name[:], ifName)

	if err := ioctl(fd, syscall.TUNSETIFF, uintptr(unsafe.Pointer(&req))); err != nil {
		return "", err
	}

	createdIFName := strings.Trim(string(req.Name[:]), "\x00")

	return createdIFName, nil
}

func setDeviceOptions(fd uintptr, config Config) error {
	if config.Permissions != nil {
		if err := ioctl(fd, syscall.TUNSETOWNER, uintptr(config.Permissions.Owner)); err != nil {
			return err
		}

		if err := ioctl(fd, syscall.TUNSETGROUP, uintptr(config.Permissions.Group)); err != nil {
			return err
		}
	}

	// set clear the persist flag
	value := 0
	if config.Persist {
		value = 1
	}

	return ioctl(fd, syscall.TUNSETPERSIST, uintptr(value))
}
