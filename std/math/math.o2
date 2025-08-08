func add(a i64, b i64) i64 {
  return a + b;
}

type event {
  dev u64
  msg str
}

type compound {
  event event
}

func test(e event, a i64, b i64) i64 {
  return 0;
}

func getEvent() event {
  let c =
    compound {
      event: 
        event {
          dev: 0,
          msg: ""
        }
    };

  var a = 
    event {
      dev: 0,
      msg: ""
    };

  test(a, 10, 11);

  test(
    event {
      dev: 0,
      msg: ""
    },
    10, 11
  );

  let b = 
    test(
      event {
        dev: 0,
        msg: ""
      },
      10, 11
    )
  );

  return 
    event {
      dev: 0,
      msg: ""
    };
}
