package util

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test certificate created with an OpenSSL command in the following form:
// # Create "root1" certification and "cert1" certification signed by "root1" certification.
// openssl req -x509 -nodes -days 365000 -newkey rsa:4096 -keyout root1.key -out root1.pem -subj "/CN=oauth2-proxy test root1 ca"
// openssl req -nodes -days 365000 -newkey rsa:4096 -keyout cert1.key -out cert1-req.pem -subj "/CN=oauth2-proxy test cert1 ca"
// openssl x509 -req -in cert1-req.pem -days 365000 -CA root1.pem -CAkey root1.key -set_serial 01 -out cert1.pem
// openssl verify -CAfile ./root1.pem cert1.pem
// # Create "root2" certification and "cert2" certification signed by "root2" certification.
// openssl req -x509 -nodes -days 365000 -newkey rsa:4096 -keyout root2.key -out root2.pem -subj "/CN=oauth2-proxy test root2 ca"
// openssl req -nodes -days 365000 -newkey rsa:4096 -keyout cert2.key -out cert2-req.pem -subj "/CN=oauth2-proxy test cert2 ca"
// openssl x509 -req -in cert2-req.pem -days 365000 -CA root2.pem -CAkey root2.key -set_serial 01 -out cert2.pem
// openssl verify -CAfile ./root2.pem cert2.pem
// # Create "root3" certification and "cert3" certification signed by "root3" certification.
// openssl req -x509 -nodes -days 365000 -newkey rsa:4096 -keyout root3.key -out root3.pem -subj "/CN=oauth2-proxy test root3 ca"
// openssl req -nodes -days 365000 -newkey rsa:4096 -keyout cert3.key -out cert3-req.pem -subj "/CN=oauth2-proxy test cert3 ca"
// openssl x509 -req -in cert3-req.pem -days 365000 -CA root3.pem -CAkey root3.key -set_serial 01 -out cert3.pem
// openssl verify -CAfile ./root3.pem cert3.pem

var (
	root1Cert = `-----BEGIN CERTIFICATE-----
MIIFLTCCAxWgAwIBAgIUGeHG3/izV2LRV/FrJ4LfhdfB+E0wDQYJKoZIhvcNAQEL
BQAwJTEjMCEGA1UEAwwab2F1dGgyLXByb3h5IHRlc3Qgcm9vdDEgY2EwIBcNMjIx
MDAyMTU1NjU2WhgPMzAyMjAyMDIxNTU2NTZaMCUxIzAhBgNVBAMMGm9hdXRoMi1w
cm94eSB0ZXN0IHJvb3QxIGNhMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKC
AgEAxVrw1wE9ePEF8WjuTXbBj8P2IiPGBUIDj3iBFqSRy5bvgHZ6RhWIDS81E0V5
MkGW9jnVnY3jh1MnAWaszIIzIWUZzhb/mP0jqlrE5Yh0294owM/sIQ/ucqSWfoug
LTzUun3iLNWfD85wXKRzbDoyzo6l2LsHVh6cOzOgVJaBO9Ulb2QGqVEyQ1mXLAaU
Dqk7MiFpAO7SW+5gcosn6rte79glc9KcEoRERG0y7cDjnKm0WDVO+35K4+VDuWev
4c0b1IM83d9n1sm3MgrKeod5LPEVmIH8iSm/36vBxrUS6ISEQwQpoX8dYtiv7D2M
b7IaewLr19wYSnwlHlYVtlU2tgejX9VavlOq76qM1s3XQ+EWrTrtnuUddhVdvvJy
ch4Rx9A35rVS7q3suEYJcPlsjA/ZzQBuhSBtUUhuithlSrFYE0t5ljNDqrzTkC3J
SP9t3/UjjbyseT+LdzmKiAMwybQBvCZxXhRo+ZJ5i0GxQw0sB4zh7T+5Vu2amdbc
eBY8vAJu75LSMzY7zo8zzQh6kxpkhq1QoYwg5JmeSGB9jQM6fF3I7y8Yogm/dMMP
Q0eLxhyzkdKffQze6Fh9uH8AtST8x1AvPM/XugC6efE6mRIFMS23ccBypPIXjeEY
FKorYbne6FsMRffTHJjaCIG5LNUB4E0gPR9CeEJi+eNNj1kCAwEAAaNTMFEwHQYD
VR0OBBYEFEUFhqI23av2StktalBkvZXUw07fMB8GA1UdIwQYMBaAFEUFhqI23av2
StktalBkvZXUw07fMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggIB
AEYa0ZOaepXi4Q2ZTHqxYLz6oWf8lPW/bqsCNHum1463qduzBhkpndMKXNUU5/Kn
i2HMpON49z+gnYDmzXhZe7AlqFPc2hjY21eglRA439HDJWSIhWjAm3c4Zz1Ov7IW
b0xYyIbIxVWBSP1tTGkCPzHxGlfHH8qzzrYu0ygyMCF6S8/npCP39nssXbw/YxhT
5HK0i9J8aQ+Kr5fbFjXkaMC/pF8/cjuNIoAj8fCAXE+WM986G4aBvgMfDf+VWfct
TrzwT/71i52k3Sb8To9Gf4Ms/dj1NdjnOBejsjH/6QYCLe5fTVf4TYjO1xRsDRRM
tcBE5UqZ2EsxNnMN0mvsqRXFKi7HSa4+OpEZXuVECr4bOhfm2lyXdpHT0NnXecV4
CPdtuTq+ARDENFG6wOt1qcDGPCFeawIAFKx2DrS72MupdUltE68nxQEncg1ElPx0
Q8euF6SPrl46vySWmkhLlWVGtjDYAOH+HiDZMUowf55SsjWYRMzVA7TfVZAcMMg2
1uPeGXIT8pvcSynP+8aZ43dYAujTjLRHLmnvY/7qmYkQsXrwtNtKpzhyeMRKQ6ym
3EX8PQMuU5f8IwyNSG3/OMGwJ4YaYxTFGBwFpkEEwalVsOutn4StKSEmk+gqXlY1
gs4g3ejPns3UdgANy7Z3974hcMNl0KNgzs49AH2xZl5w
-----END CERTIFICATE-----
`
	root2Cert = `-----BEGIN CERTIFICATE-----
MIIFLTCCAxWgAwIBAgIUCAc9oeOnhzVN3SW3oI0QctsOZswwDQYJKoZIhvcNAQEL
BQAwJTEjMCEGA1UEAwwab2F1dGgyLXByb3h5IHRlc3Qgcm9vdDIgY2EwIBcNMjIx
MDAyMTU1NjU3WhgPMzAyMjAyMDIxNTU2NTdaMCUxIzAhBgNVBAMMGm9hdXRoMi1w
cm94eSB0ZXN0IHJvb3QyIGNhMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKC
AgEA4jC4EF2XFcqwZxqMYQmPFVOLDCkiuRkqXwT2ps6t21UpjIYS6Bs1Kg/mMHog
Zt4dY3u/acIVws23I+wbcrPCQJiySMXdrU16qhP1aGBvVwVmkUQTrUi1VRR18DGf
TzUm9KQ6Jpf/wJoU/2F7sbH3qSllTQ4PM7Sg6FLO7Gqzti6Xet90HOnwzDnXnOBR
A09Scheqyy7UJcfBDK+g9FAxvr08Gn5Olbd//Os4kuoyCpfWlpz353U0fHRWMpkt
S42o+KfolTYNdaYsImc3QOUeZD4l1ME5bj26bJG1tkmVpe3j3E4QxT9ZP30Bl3Fh
mMKPxsZ51d/iXHkQMBZSxrXgoADyxdFiYUGNmeUkm+uTsfbtJQp2dGHTD0YCvfKg
dmcIkMTb8ivyAYxqlA5BKxgHF3YGBI0oEMb1QRFmQGLezw5V1MD+a3zuP6XJDHJb
NZA05ZWdCMWec9W8KJiYVeD9MrZzQCoFlrhAycJx91V5MLyXxi1sjQPEQ+a/cqp4
dprZ4oOaeOU/tYTOxVAu88rTi1cVBpkUvZXjKjS6PLEGwTYyN+h7knrIGQX24/pf
juo5zhS9xhi6WdRxIzDBHIB6FnruGPE7uzh9wHHW0x4iz3W2/7dJGahHPj1P71sH
VjjNYJSS0bc6Lg/B2r44pMEK2QOyzvDQzkHwUgtBMpIDGCkCAwEAAaNTMFEwHQYD
VR0OBBYEFJKec+oSsW7LCA4ffv22lj7jPCbyMB8GA1UdIwQYMBaAFJKec+oSsW7L
CA4ffv22lj7jPCbyMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggIB
AJD1KUIFb86Ukx3ecuBFevcjESfwpKym23XLWxdrH6s3c+DA55WqpyH/SQXKddxL
qtmfiS9XH+Nf/mOHTTelhJin8pLFHd0V7UFlRcFnb/4uBHNLqEA53f5ORQ+YR6zY
o6kh3aXPNx/oj8JZXupPltlwrSgLLGYMB3wPtE7EfUQarEHhcOChDUkzh5UJdo3p
a4X2lg9t/kWCDyjWApnC5Uo5prBmDefZcp46LiBFZyJz82y1BsPN3aNOEUgGhLSU
7tQFQbaW5TSKnLyr+ntSmAPUDgA63OMJ6j4seixjPgY3KG8KzQ1uVK5aCJxczsaH
P22QuMFSbKC5s+7kmmEbkrnI+4XSpOO2kcBQE2djM4pRFG+q7KMtTGyTdxLE/0UV
qmx0ZysvUbrDUyQknGK9OIzA3THbVTt4ZG6XNlMIC/PZLor3if/f82Jg1x7y5q23
j6wgN7208tLOr+KfSgZzNhy/pd1oREgBDYiFEEZnKbQPrByecZYGovUVYtu/GjAE
vdoTORNVp76NPxtjACPTJ5XALKAlLwPLJke4XrssNAOoG3am4S8uS9Bvi7bJEGLV
2OThwvTOjenJpPOmAMMmdeXYWdRJEGy7d4dvY5WchjusIyEJS7JiCTPKNPHEM9IT
gODlyyBPhMZ55ZUAV8SmqrZYPxOt6YycO3gMBFp7kQt6
-----END CERTIFICATE-----
`
	cert1CertSubj = "CN=oauth2-proxy test cert1 ca"
	cert1Cert     = `-----BEGIN CERTIFICATE-----
MIIEwDCCAqgCAQEwDQYJKoZIhvcNAQELBQAwJTEjMCEGA1UEAwwab2F1dGgyLXBy
b3h5IHRlc3Qgcm9vdDEgY2EwIBcNMjIxMDAyMTU1NjU3WhgPMzAyMjAyMDIxNTU2
NTdaMCUxIzAhBgNVBAMMGm9hdXRoMi1wcm94eSB0ZXN0IGNlcnQxIGNhMIICIjAN
BgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA4fO9HqGLxVNlFGWUPO1t5bXQomd5
0Mjb8gQihfgDvAMzPnvsSGTZXv7zhPAXcv4IkfP8J4SkhpUwHYyyyHlcRJfYhrdk
XnqA9+3OnOgc6dKgk5UAojqQU0Ud4LdSSzelpf7JwMNiX67ogXbQOnYn+XZCJODR
dAOTpyytXc7SBWXqkU8Oqyjbjo3B63cWNF2UFjXuWD8CKFhQ5MAVyPaMYiR2PZOv
fjPUxMH0REpB5A2UJPS/cH0MX6HqSWMAMA7huCbkw0Lnzrpl3sk9Y2SVID3sl6di
cx8puxH3SzuMkJXB7zyiJe24zlBkPLWl8ax1eKm8VyyzxTbocEnshpGTtEGoC1sT
tKNreuJH589KbceqE/XHVcEWA5bGghJRVL/siImdKh4tZ+lN5QG+hlZz4Zi77AY6
f+FDXsBcUnHX3Jmx+ToXiOPYdTmpir1pBMUzkPaVERUJtqbFzdOHFFdk6blxFeTK
ZnYynxIDVTTEZTwp/aKeXHQ+Cv01R2FRDlA1mDYmUA8Oxzhq3/DNnaz97hBn2jn8
7AaclwGDdgT3AXyFOm6zV5vGJDRpK/timefguDPw6KFMznjq+9/azLosem7grMen
Tu6S8qdb62eXiOq6fXHESeL5TMGB3NKoRq75/y1akGbZbnM3zKiq6xs0i3zv4m4B
Y4qg704mQbJ4lsUCAwEAATANBgkqhkiG9w0BAQsFAAOCAgEAslHL7NxcCqxGYPut
ee+aJUBYsIknKWgwSdH/uNQAp0f2vzdaJvK/Fq2RtEH8q6ASK5yr8T8DXO/RvhCB
P8DnvMV5e9OUiyBkb/GJvKnuPGbtKpsvzOnEEx8PGvMCH2TkgBf13UXLO8/QKVNY
9iVLXZWZBWWwa0mk7OgEtxpTiuK7lqYKs4SpxhLHgOoFfgaJYzcvEUokHcySOSBh
GrtZgzqENqTz3Sh/hnxkIuYVPLmXYUnTHSt7P1vNe8XDQBQOb4HJkBP2uIVt/4gJ
Dx/XNuzyk7OHZFm1z8uxfkz/c4DQ1JhcnrrEjO0rWr0lxBTCZwmgQo+xrzfIyDTa
lTTxp4K5277JIVg/8kNZvfsap9xjsmM3o7zZHfTbYZq/Jp9vm5DSY3Nzmmc3Alvt
feOCuxvQLh/TLOgeFBYMRIGm63/NZTMaPOrSxCjBQVneMA/wHiujL+0qPvV71fv9
TKT2pRDqByco73D50FjyE4MWPok/cG6MHMAF28/3t5Ibgbhlod4JQrdReE2ripAb
CblIzq3iawqFqtGCJ0HvzaJu7v+kEh5xNb+GaUrnC/muecNSl2HFdfuYsMmxVwGW
lDrTYyOeO3A7KR5b9f5XrCYNWQACbln3oLKbS1DZzMnbOwvHZZ6Wvf+FUepLDh7M
EThrpRLeQ3ka/CBkdRcXGEh+Odw=
-----END CERTIFICATE-----
`
	cert1Key = `-----BEGIN PRIVATE KEY-----
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
8rjOfdqojUTBoHbTvd5IU1EX8zJgdHo=
-----END PRIVATE KEY-----
`
	cert2CertSubj = "CN=oauth2-proxy test cert2 ca"
	cert2Cert     = `-----BEGIN CERTIFICATE-----
MIIEwDCCAqgCAQEwDQYJKoZIhvcNAQELBQAwJTEjMCEGA1UEAwwab2F1dGgyLXBy
b3h5IHRlc3Qgcm9vdDIgY2EwIBcNMjIxMDAyMTU1NjU4WhgPMzAyMjAyMDIxNTU2
NThaMCUxIzAhBgNVBAMMGm9hdXRoMi1wcm94eSB0ZXN0IGNlcnQyIGNhMIICIjAN
BgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA9bVSpGQIs4cmuaTGwGcaN1rjzGuj
NbnO2qBpld7U1WeWpE0fifMDxBqhEUR4gks0VfCV6ecA8x3bVSKi7tl5H6+3oxL1
gbcae3d9MnOFqDzy1gywn2SQeSz3ktR1DaF823P2ibGFtY1Ab0W/9DVHzj/o+vAJ
ZHCipjw59l1D4T4YTP0E+kZX4596hKT9Hwi3jq8A5GR5LZzLUXwb1X+xEEgimc3X
0or8qIK8tzDkHk564ActJNLF1RpWAKj5L8f97OyWGh2JQILAosprLkOy9XIOin7i
tBuAEqaLMcWPwz9lh45bl99sE4bRJC5unYbMn80XLOpAgqvH34us4mPWjnrzp6p9
wN2kRukhaqVBgYhxHaLXnaYLzbbSdDj3EkJRSxxX10a4lDFSDOzmJx0VB6pxbn8T
zL/ou6Kqas42ia/nTmnmJ4kvmzxRQmfhoNhp270TjaQ0ZxWEZcd6bVKCjPih/ikZ
oYl2nxpVgK/eAEcJDYlwSK4gEvT0GThCe6oXr0XwTxY7GZQE9lKiPUTFsqsQXyXQ
Xlf0ADxgipkam3z5BP7mCerkOHSX/MMzYRCbcZFntM94nIA/8Rx+/YSSHOQs1xSS
LkjPFu0Mio6MrJMrwvLoXwAHtGM898TA4ClJJ9w4nuHiPCoHGQTljuw0f92yviFR
QkN4sPUcIRnuvLsCAwEAATANBgkqhkiG9w0BAQsFAAOCAgEAtF+Mmdryv98LGNcA
IiNchCGTtggZQ/0XGblLualBvHAMOvBL+6gGjJ+VgoQjadDK5aqzWi+101J/Cqh6
5zJ7Fy+NZ7P/uoGW/QHPdbXkYcfwI3gFr3RsgAN05M+DkbVoGW8jSjCTZuoUr1KK
heQqpGuZEnOPeJNcWFbZRG4TsHEfFUTrD0Po4W4lqS8BsvNT8djmt9eYwaaJl4AP
zJLEIgLlhIhz1rjryQSawfcT7NYt8Ku2KwXGKdP5P6OpUgd39gMp/oszOA0M20TH
be8QcIw1VL7p3G+Fk1qLfztGCASOmh9TNFaJg7NW0BCuTwk+6epbMLfDN+w1MZ+T
hd94SgTc6HOx4EHrdJdlQ999NS4B/hGVBRPDgsv+4vG+JsygH3uB00EhGDT8IKxD
KGAKWCKeinoZWeVKNThKPBb5B7ogeL1uG8MXlaj86Y9yHRZt9g9bvNtMOTE2qtrO
Yikzo+Ug+H4s3VX4fi/540q0F4nfWaWolhsLM+7LVSgArweDUjDdYfqpRizP7S2M
Fi7RlwmTPNmkx4UeOsOjwmhsIRagJHsi1jP9BL7qeUGqIL2lknITEhtAcNzk3Y09
soml63C8enibw/jQyh9+ON+o+uj55yc1EPPp4ZdZPWtQJkYpu5kfvr2NgWt4D1Di
L+TQPQQX7L5UUFjXQvzMaZnhDIk=
-----END CERTIFICATE-----
`
	cert2Key = `-----BEGIN PRIVATE KEY-----
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
XRoBYXCF19tcBwJWGC5kOfP9E8hl94La
-----END PRIVATE KEY-----
`
	cert3CertSubj = "CN=oauth2-proxy test cert3 ca"
	cert3Cert     = `-----BEGIN CERTIFICATE-----
MIIEwDCCAqgCAQEwDQYJKoZIhvcNAQELBQAwJTEjMCEGA1UEAwwab2F1dGgyLXBy
b3h5IHRlc3Qgcm9vdDMgY2EwIBcNMjIxMDAyMTU1NzAxWhgPMzAyMjAyMDIxNTU3
MDFaMCUxIzAhBgNVBAMMGm9hdXRoMi1wcm94eSB0ZXN0IGNlcnQzIGNhMIICIjAN
BgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA6DibDDVCzbqaZgvxrD4gUsHAbM4Y
YbZps7FOlXCKgbpyN17aCy6enmqhlzjJDWbbRNRfGpEFQpSqlUaHeAxCvpXBZGNL
7zmF3X3mOuve1uO+5rmenySd38aiG9noOJO0tnLxogBf6+IJlTBKmXoS3JVMG851
5kR1D6Q5zWd2/reeAN9eIZkPtTIymIXngKlIj9Y0X6k98WIf8waElZ1srga1RefJ
yg481ZQrPx6/gDmijWXwM3lbBxlg2HiXkoj7WQA6+Nf+HsyD8KrqO78u3TSlKhPh
MF9x73almoy+9mlyidf1T7lqflv5eGOtG5uQIgNDQMXpTqToJw1nmO0Vpr+RkkUR
Xr0Mq+ju1bHghufRuiJGu9/uaLz9n8zfBfINX7F6nLj2T0Wu5AhtwQbdU+WNEEaW
OWhzczHyxOQTrAk4I+ldWezqITGD+aqtdp2SyPMMqZc8/046KFDMFrvjzsi8FlIs
7SePaFIV7aJ5izk7112z4eNKw4PxN+2hMG99n+UbUkxVNFZzhdPf8B1PZnjO1s+r
Ei4ZULMuO6dthW4OLkQa0E5OuzZxH36OsC2e0Vna115FVzdievKMFutyAnkCm6pX
YslPUrxtxUJ3ksTjlRMeNbytCKf8z17bpiKWoAn+HmtMeFrWeiu+Fubn3nr3zcUW
z0f9h2QSk83NAukCAwEAATANBgkqhkiG9w0BAQsFAAOCAgEAO4gfydvaulLsf+J3
Vl2qO4gqbDLCl3gr+kkt40dUsrMxVMnzG0eVLavdhOf5LoTzVlbfwpr4LW6dhDnG
KU0bqn/wz78AYtbDv+bsr8wZqNC+kDyibtQqIa3Rk+DAfc6sz3WfjHns6nJ8fjvT
HujaWCRKYl5PrIUepZRVm71KRIZi1ufEsCrWj6HDiP0rBnxpaCvTuqHlb6v4yax2
pysA5j32Dsa9pndYpSAQRLheGwopIFevsXScrzSgbCaiZfEV8pFO3LpdBG7qaj4A
I3/5YRqrkbtn4C23NZY6oOg6gGYmklIopP3HoRz3VikXtD8JzRJI/8Lz+LRWSayd
tDj/cByrgnvkFFZPyDh3xNnPn/QT3IVQJOPeAJW9gundC2gc0Sjo73MnKetKbyG9
TdsW5+GXBINQUCHTiA0qLZlw0ejuT5gVjsQRXvQP466gQnA1r4e6L/ZFBsIRp960
d40E3zNpUqeCOgucZlZycU60aC+GFaxbNJHLb47VKZLXSvs0rV6kW7EozKb9NdeZ
sXyUVOp8cF6a9G/MSlYAN0WY4N41lvvPXShStq9kTQMJkaOa6xZc1Y06Og4Q36wf
ru1Bby42wGHXSD7YB1IjHjDF1vA1prTGDVIdbUq2PjOOdFcuX82em79EUzv2hP7X
pwrBB0KdhQb2fk9ZT+Qtw9H+LNw=
-----END CERTIFICATE-----
`
	cert3Key = `-----BEGIN PRIVATE KEY-----
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
1QCB3z4iKoxwYQaPBFOXJuyzO5Bi3V8=
-----END PRIVATE KEY-----
`
)

func makeTestCertFile(t *testing.T, pem, dir string) *os.File {
	file, err := os.CreateTemp(dir, "test-certfile")
	assert.NoError(t, err)
	_, err = file.Write([]byte(pem))
	assert.NoError(t, err)
	return file
}

func TestGetCertPool_NoRoots(t *testing.T) {
	_, err := GetCertPool([]string(nil))
	assert.Error(t, err, "invalid empty list of Root CAs file paths")
}

func TestGetCertPool(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "certtest")
	assert.NoError(t, err)
	defer func(path string) {
		rerr := os.RemoveAll(path)
		if rerr != nil {
			panic(rerr)
		}
	}(tempDir)

	certFile1 := makeTestCertFile(t, root1Cert, tempDir)
	certFile2 := makeTestCertFile(t, root2Cert, tempDir)

	certPool, err := GetCertPool([]string{certFile1.Name(), certFile2.Name()})
	assert.NoError(t, err)

	cert1Block, _ := pem.Decode([]byte(cert1Cert))
	cert1, _ := x509.ParseCertificate(cert1Block.Bytes)
	assert.Equal(t, cert1.Subject.String(), cert1CertSubj)

	cert2Block, _ := pem.Decode([]byte(cert2Cert))
	cert2, _ := x509.ParseCertificate(cert2Block.Bytes)
	assert.Equal(t, cert2.Subject.String(), cert2CertSubj)

	cert3Block, _ := pem.Decode([]byte(cert3Cert))
	cert3, _ := x509.ParseCertificate(cert3Block.Bytes)
	assert.Equal(t, cert3.Subject.String(), cert3CertSubj)

	opts := x509.VerifyOptions{
		Roots: certPool,
	}

	// "cert1" and "cert2" should be valid because "root1" and "root2" are in the certPool
	// "cert3" should not be valid because "root3" is not in the certPool
	_, err1 := cert1.Verify(opts)
	assert.NoError(t, err1)
	_, err2 := cert2.Verify(opts)
	assert.NoError(t, err2)
	_, err3 := cert3.Verify(opts)
	assert.Error(t, err3)
}

func TestGetClientCertificates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "certtest")
	assert.NoError(t, err)
	defer func(path string) {
		rerr := os.RemoveAll(path)
		if rerr != nil {
			panic(rerr)
		}
	}(tempDir)

	certFile1 := makeTestCertFile(t, cert1Cert, tempDir)
	keyFile1 := makeTestCertFile(t, cert1Key, tempDir)
	certFile2 := makeTestCertFile(t, cert2Cert, tempDir)
	keyFile2 := makeTestCertFile(t, cert2Key, tempDir)
	certFile3 := makeTestCertFile(t, cert3Cert, tempDir)
	keyFile3 := makeTestCertFile(t, cert3Key, tempDir)

	certs := []string{
		certFile1.Name(),
		certFile2.Name(),
		certFile3.Name(),
	}

	keys := []string{
		keyFile1.Name(),
		keyFile2.Name(),
		keyFile3.Name(),
	}

	for i := range certs {
		tlsCerts, err := getClientCertificates(certs[i], keys[i])
		assert.NoError(t, err)
		assert.Equal(t, len(certs), len(tlsCerts))
	}
}
