package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCodeOwnerFingerprints(t *testing.T) {
	result := getCodeOwnerFingerprints("test_fixtures/CODEOWNERS_FINGERPRINTS")

	assert.NotNil(t, result)

	email := "georg.pfuetzenreuter@suse.com"
	fingerprint := "9cf35828cec50de0294e04a1c645433b1e5e7a65"

	assert.Contains(t, result, email)
	value, ok := result[email]
	assert.True(t, ok)
	assert.Equal(t, value, fingerprint)
}

func TestGetCodeOwnerUsernames(t *testing.T) {
	result := getCodeOwnerUsernames("test_fixtures/CODEOWNERS_USERNAMES")

	assert.NotNil(t, result)

	email := "georg.pfuetzenreuter@suse.com"
	username := "crameleon"

	assert.Contains(t, result, email)
	value, ok := result[email]
	assert.True(t, ok)
	assert.Equal(t, value, username)
}

func TestGetExclusions(t *testing.T) {
	result := getExclusions("test_fixtures/UNTRUSTED_COMMITS")

	assert.NotNil(t, result)

	assert.Contains(t, result, "5ea8af38993138b5164451434b453ad9fd3993bd")
}

func TestBuildKeyring(t *testing.T) {
	result := buildKeyring("test_fixtures/CODEOWNERS_USERNAMES", "https://gitlab.suse.de")

	assert.NotNil(t, result)

	assert.Contains(t, result, "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: GopenPGP 2.8.0\nComment: https://gopenpgp.org\n\nxjMEYjjrdhYJKwYBBAHaRw8BAQdA2a6xdQeqXgqaZQqecJRAaFzzW/einDHKZPv1\n7A+ZncjNLkdlb3JnIFBmdWV0emVucmV1dGVyIDxncGZ1ZXR6ZW5yZXV0ZXJAc3Vz\nZS5kZT7CkwQTFgoAOxYhBJzzWCjOxQ3gKU4EocZFQzseXnplBQJi5DKZAhsBBQsJ\nCAcCAiICBhUKCQgLAgQWAgMBAh4HAheAAAoJEMZFQzseXnplUGABAJ7z2ivz0euF\n8DilirmJ7Y9cptieOyndFUtx2DfV4dAXAQDEuQiK/KJO/NFqZLQUZGrvvgi5EUp0\nafFkM5Ek7FkSDc0pR2VvcmcgUGZ1ZXR6ZW5yZXV0ZXIgPGdlb3JnQGx5c2VyZ2lj\nLmRldj7ClAQTFgoAPBYhBJzzWCjOxQ3gKU4EocZFQzseXnplBQJiOOu/AhsBBQsJ\nCAcCAyICAQYVCgkICwIEFgIDAQIeBwIXgAAKCRDGRUM7Hl56ZexhAQDQumpSnljD\nBq1Y7MkabWxvEb9olsPVjhGTj+gLX3osyQD/ZrhA/fnsGzuK95qp6FuVUf0cAwcH\noSbKc4iMggqG8wPNNEdlb3JnIFBmdWV0emVucmV1dGVyIDxtYWlsQGdlb3JnLXBm\ndWV0emVucmV1dGVyLm5ldD7ClwQTFgoAPwIbAQULCQgHAgMiAgEGFQoJCAsCBBYC\nAwECHgcCF4AWIQSc81gozsUN4ClOBKHGRUM7Hl56ZQUCYjj25QIZAQAKCRDGRUM7\nHl56ZTqxAQD/IDCu44FAfKXD/tvtfzMR0mlqlx+Omqe1Y2J40UQtegEA1JZ67OhE\n4ozopQu3BZO4KXNQOL0WvAYsIBCmN8AGVAzNNEdlb3JnIFBmdWV0emVucmV1dGVy\nIDxnZW9yZy5wZnVldHplbnJldXRlckBzdXNlLmNvbT7CkwQTFgoAOxYhBJzzWCjO\nxQ3gKU4EocZFQzseXnplBQJigg7qAhsBBQsJCAcCAiICBhUKCQgLAgQWAgMBAh4H\nAheAAAoJEMZFQzseXnplHQwA/28VZVfsz+d0BgbLymDB/RiOdrBtlyFufEzpt1TK\n5M72APwIi2GzxEuehr/cuQ8Wz3ZJ9nbqEGxMa2p08/FYzwsjB84zBGI46/8WCSsG\nAQQB2kcPAQEHQL/wP1XqT2dpZNIRHUcKIGpg56zP6/puAQcs0mnHjVOvwsA1BBgW\nCgAmAhsCFiEEnPNYKM7FDeApTgShxkVDOx5eemUFAmX8VfcFCQWknXgAgXYgBBkW\nCgAdFiEEVeDyauW9ojw970nRHtLxOOfm/1cFAmI46/8ACgkQHtLxOOfm/1elQAD+\nIraPrE+YRwJoXNnv/6KPRowU5zc99CTDYlTtscDVvxcA/1f9d1znC1g1paXbWaSw\nSexHy5aFL/iSVJu/RP1S0/ULCRDGRUM7Hl56ZZ++AQCY6kAEcb35XXFt0kPDQKiM\nWDZkGvZ0h7xWmpPehDE53QD/XRa9BS1vDTApOHgdfp9rhEjLSgDBIKhA8CqEmUuo\nZQrOOARiOOwaEgorBgEEAZdVAQUBAQdAOl/vqDvtNgpp9KNZHcLYU2pOKLKMzevF\nrLS7pQyL2DgDAQgHwn4EGBYKACYCGwwWIQSc81gozsUN4ClOBKHGRUM7Hl56ZQUC\nZfxV9wUJBaSdXQAKCRDGRUM7Hl56ZSWcAQC9B+ubejiLSAd2Rgg30OppINVaC82x\nrNeoQmkTQbPIBwEApGQldbaxA3wDDtc9cXdE7vzpPHVl+ZZnTn3owT8oSArOMwRi\nOO3OFgkrBgEEAdpHDwEBB0D/bUnXqnyPWAptJh5ZdpV14H/5L0ECVDIO0FSDE2f+\nz8J+BBgWCgAmAhsgFiEEnPNYKM7FDeApTgShxkVDOx5eemUFAmX8VfgFCQWkm6kA\nCgkQxkVDOx5eemVzDAD9FswnHkAL1HEXGb/loCw/cvdJttpCNo8vXHNcESML4aoB\nANDbcfRZNywFvp2iFyGJpianOiUj15bfvg8NZ0sSJ/IJzjMEYoTDcxYJKwYBBAHa\nRw8BAQdAzweiSFkBlBG8FC68duBg5a3uCdVJJOHPmEZWZm22TW7CfgQYFgoAJgIb\nIBYhBJzzWCjOxQ3gKU4EocZFQzseXnplBQJkXRqABQkFmr4NAAoJEMZFQzseXnpl\nmtYA/jHalyTyt98IzzbI/64rNA54ub6uVv5W4MfwAoRVh9BLAQCt3T7XIb9Ee41e\nSPm7ZaQO2JN/XMbnrHu4+fRDZwiiCA==")
}
