package secretz

import "github.com/awnumar/memguard"

func FromEnclave(enclave *memguard.Enclave) *memguard.LockedBuffer {
	var err error

	if enclave != nil {
		var locked *memguard.LockedBuffer

		if locked, err = enclave.Open(); err == nil {
			return locked
		}

		panic(err)
	}

	return memguard.NewBuffer(0)
}
