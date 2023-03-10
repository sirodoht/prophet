# prophet

> A Prophet stood alone with sight so clear,  
> Of things to come, a premonition dear.  
> With visions bold, he spoke of what was near,  
> Of troubles, pain, and hope to conquer fear.  

Long-form content nostr client. Pre-alpha.

## Development

With Nix and direnv, setup PostgreSQL:

```sh
# clone & cd prophet/
cd postgresql/
cp .envrc.example .envrc
make init
```

Then, back to the repo root:

```sh
cd ..
make serve
```

## Tools

There is key encoding tool (hex -> nsec/npub) in the codebase:

```sh
go run cmd/encodesecretkey/main.go put-hex-secret-key-here
go run cmd/encodepublickey/main.go put-hex-public-key-here
```

## License

MIT
