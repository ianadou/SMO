package ports

// PasswordHasher abstracts the password hashing scheme used by the
// authentication flow. The concrete implementation (bcrypt today) lives
// in infrastructure/auth/bcrypt.
//
// Hash takes a plain-text password and returns its opaque hash. Compare
// returns nil if the plain password matches the stored hash, an error
// otherwise. Comparison is constant-time inside the bcrypt
// implementation.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}
