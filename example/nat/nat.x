type event {
  a i64
  b u64
  c f64[10]
}

@native
func add(i64, i64) i64

@native
func sub(i64, i64) i64

@native(16)
func mean(i64) i64
