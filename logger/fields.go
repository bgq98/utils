package logger

/**
   @author：biguanqun
   @since： 2023/10/14
   @desc：
**/

func String(key, val string) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}

func Any(key string, val any) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Int64(key string, val int64) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Int32(key string, val int32) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Bool(key string, b bool) Field {
	return Field{
		Key:   key,
		Value: b,
	}
}
