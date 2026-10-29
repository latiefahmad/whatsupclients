//go:build !windows && !darwin

package sound

// No player here: Play does nothing and the toast's own sound is all there
// is.
