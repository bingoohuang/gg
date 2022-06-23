package fp

import "log"

func Must[T any](t T, err error) T {
	if err != nil {
		log.Fatalf("failed: %v", err)
	}

	return t
}

func Must2[T, T2 any](t T, t2 T2, err error) (T, T2) {
	if err != nil {
		log.Fatalf("failed: %v", err)
	}

	return t, t2
}

func Must3[T, T2, T3 any](t T, t2 T2, t3 T3, err error) (T, T2, T3) {
	if err != nil {
		log.Fatalf("failed: %v", err)
	}

	return t, t2, t3
}

func Must4[T, T2, T3, T4 any](t T, t2 T2, t3 T3, t4 T4, err error) (T, T2, T3, T4) {
	if err != nil {
		log.Fatalf("failed: %v", err)
	}

	return t, t2, t3, t4
}

func Must5[T, T2, T3, T4, T5 any](t T, t2 T2, t3 T3, t4 T4, t5 T5, err error) (T, T2, T3, T4, T5) {
	if err != nil {
		log.Fatalf("failed: %v", err)
	}

	return t, t2, t3, t4, t5
}
