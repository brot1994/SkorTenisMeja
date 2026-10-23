"""Minimal APK Signature Scheme v2 signer (RSA PKCS#1 v1.5 + SHA-256).

Usage: python3 v2sign.py in.apk out.apk keystore.p12 password alias
The input APK may already carry a v1 (JAR) signature; it is kept.
"""
import struct, sys, hashlib
from cryptography.hazmat.primitives.serialization import pkcs12, Encoding, PublicFormat
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import padding

def lp(b): return struct.pack('<I', len(b)) + b

def find_eocd(data):
    i = data.rfind(b'PK\x05\x06')
    if i < 0: raise SystemExit('EOCD not found')
    return i

def chunk_digest(sections):
    chunks = []
    for sec in sections:
        for o in range(0, len(sec), 1 << 20):
            c = sec[o:o + (1 << 20)]
            chunks.append(hashlib.sha256(b'\xa5' + struct.pack('<I', len(c)) + c).digest())
    return hashlib.sha256(b'\x5a' + struct.pack('<I', len(chunks)) + b''.join(chunks)).digest()

def main(inp, out, ks, pw, alias):
    data = open(inp, 'rb').read()
    e = find_eocd(data)
    cd_size, cd_off = struct.unpack('<II', data[e + 12:e + 20])
    contents, cd, eocd = data[:cd_off], data[cd_off:cd_off + cd_size], data[e:]
    key, cert, _ = pkcs12.load_key_and_certificates(open(ks, 'rb').read(), pw.encode())
    digest = chunk_digest([contents, cd, eocd])
    ALG = 0x0103  # RSASSA-PKCS1-v1_5 with SHA2-256
    digests = lp(lp(struct.pack('<I', ALG) + lp(digest)))
    certs = lp(lp(cert.public_bytes(Encoding.DER)))
    signed_data = digests + certs + lp(b'')
    sig = key.sign(signed_data, padding.PKCS1v15(), hashes.SHA256())
    signatures = lp(lp(struct.pack('<I', ALG) + lp(sig)))
    pub = key.public_key().public_bytes(Encoding.DER, PublicFormat.SubjectPublicKeyInfo)
    signer = lp(signed_data) + signatures + lp(pub)
    v2_value = lp(lp(signer))
    pair = struct.pack('<Q', 4 + len(v2_value)) + struct.pack('<I', 0x7109871a) + v2_value
    block_size = len(pair) + 8 + 16
    block = struct.pack('<Q', block_size) + pair + struct.pack('<Q', block_size) + b'APK Sig Block 42'
    new_eocd = bytearray(eocd)
    struct.pack_into('<I', new_eocd, 16, cd_off + len(block))
    open(out, 'wb').write(contents + block + cd + bytes(new_eocd))

if __name__ == '__main__':
    main(*sys.argv[1:6])
