# decifrare
Transparent decryption gateway for encrypted objects on S3

## Install

```sh

go get github.com/harshavardhana/decifer

```

## Testing

### Upload an encrypted file

```sh

go run upload-encrypt.go <my-largefile> <my-largefile-encrypted>

```

```sh

go run upload-encrypt.go /boot/initrd.img-4.4.0-62-generic /tmp/initrd.img-4.4.0-62-generic-encrypted

```

Now upload to Minio server using `mc cp` as `initrd.img-4.4.0-62-generic`

```sh

mc cp /tmp/initrd.img-4.4.0-62-generic-encrypted myminio/testbucket/initrd.img-4.4.0-62-generic

```

### Configure `mc`

Now generate the download URL pointing to your encryption gateway server assuming its configured with `mc` as shown below.

```sh

mc config host add my-gateway http://localhost:8000 USWUXHGYZQYFYFFIT3RE  MOJRH0mkL1IPauahWITSVvyDrQbEEIwljvmxdq03  S3v4

```

Now generate presigned URL to download the encrypted object

```sh

mc share download my-gateway/testbucket/initrd.img-4.4.0-62-generic

URL: http://localhost:8000/testbucket/initrd.img-4.4.0-62-generic
Expire: 7 days 0 hours 0 minutes 0 seconds
Share: http://localhost:8000/testbucket/initrd.img-4.4.0-62-generic?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=USWUXHGYZQYFYFFIT3RE%2F20170213%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20170213T233704Z&X-Amz-Expires=604800&X-Amz-SignedHeaders=host&X-Amz-Signature=5372b12b3c11a2cd5fc65692ee6bd214a78c620b7206aac7b25525c5804f38ed

```

### Validate

Now download the object using `curl` and calculate `md5sum`

```sh

curl http://localhost:8000/testbucket/initrd.img-4.4.0-62-generic?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=USWUXHGYZQYFYFFIT3RE%2F20170213%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20170213T233704Z&X-Amz-Expires=604800&X-Amz-SignedHeaders=host&X-Amz-Signature=5372b12b3c11a2cd5fc65692ee6bd214a78c620b7206aac7b25525c5804f38ed | md5sum -
1cf8f421b1a0f9bbcd7825ba0016b13b -

md5sum /boot/initrd.img-4.4.0-62-generic
1cf8f421b1a0f9bbcd7825ba0016b13b /boot/initrd.img-4.4.0-62-generic

```

`md5sum` should match then we have successfully decoded the encrypted object upon a presigned download operation.

