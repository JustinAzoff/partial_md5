# partial_md5
Figure out if it's possible to truncate a large file so that it has a particular md5.

## Quickstart

So, say we have a 32MB file (contents not really relevant)

```
justin@mbp partial_md5 % dd if=/dev/random of=test_file bs=1m count=32
32+0 records in
32+0 records out
33554432 bytes transferred in 0.021748 secs (1542874379 bytes/sec)
justin@mbp partial_md5 % du -hs test_file
 32M	test_file
```

The md5sum of this random file is `ab14b102968f20181dcc34c61c2b7509`:

```
justin@mbp partial_md5 % md5sum test_file
ab14b102968f20181dcc34c61c2b7509  test_file
justin@mbp partial_md5 %
```

However, let's say someone reports that they have that file, but it has a different md5sum of `74f66c2026e7f0cda04dbff8288b9d97`.
Can we determine if this is a valid md5sum for some subset of the file, in other words, that the file was incomplete.
This would be easy to determine if we knew how big the other file was.
Since we don't, we need to determine if there is any possible subset of `test_file` that results in an md5 of `74f66c2026e7f0cda04dbff8288b9d97`

`partial_md5` does exactly this:

```
justin@mbp partial_md5 % ./partial_md5 test_file 74f66c2026e7f0cda04dbff8288b9d97
2023/01/24 11:34:46 Hashing test_file until 74f66c2026e7f0cda04dbff8288b9d97 is found
2023/01/24 11:34:46 The file is 33554432 bytes long
2023/01/24 11:34:46 Using a chunksize of 4194304 for 8 CPUs
2023/01/24 11:34:47 Found hash 74f66c2026e7f0cda04dbff8288b9d97 after 29360128 bytes
```

And we can confirm that that is the case:

```
justin@mbp partial_md5 % dd if=test_file bs=29360128 count=1|md5sum
1+0 records in
1+0 records out
29360128 bytes transferred in 0.090910 secs (322958178 bytes/sec)
74f66c2026e7f0cda04dbff8288b9d97  -
```
