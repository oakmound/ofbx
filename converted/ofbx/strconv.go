package ofbx

func intFromString(str, end string, val *int) string {
	*val = atoi(str)
	iter := 0
	for iter < end && s[iter] != ',' {
		iter++
	}
	if iter < end {
		iter++
	}
	return str[iter:]
}

func uint64FromString(str, end string, val *uint64) string {
	*val = strtoull(str, nullptr, 10)
	iter := 0
	for iter < end && s[iter] != ',' {
		iter++
	}
	if iter < end {
		iter++
	}
	return str[iter:]
}

func int64FromString(str, end string, val *int64) string {
	*val = atoll(str)
	iter := 0
	for iter < end && s[iter] != ',' {
		iter++
	}
	if iter < end {
		iter++
	}
	return str[iter:]
}

func doubleFromString(str, end string, val *float64) string {
	*val = atof(str)
	iter := 0
	for iter < end && s[iter] != ',' {
		iter++
	}
	if iter < end {
		iter++
	}
	return str[iter:]
}

func floatFromString(str, end string, val *float32) string {
	*val = float32(atof(str))
	iter := 0
	for iter < end && s[iter] != ',' {
		iter++
	}
	if iter < end {
		iter++
	}
	return str[iter:]
}

func fromString(str, end string, val *float64, count int) string {
	iter := 0
	for i := 0; i < count; i++ {
		*val = atof(iter)
		iter := 0
		for iter < end && s[iter] != ',' {
			iter++
		}
		if iter < end {
			iter++
		}

		if iter == end {
			return str[iter:]
		}
	}
	return str[iter:]
}

func vec2FromString(string str, string end, Vec2 *val) string {
	return fromString(str, end, &val.x, 2)
}

func vec3FromString(string str, string end, Vec3 *val) string {
	return fromString(str, end, &val.x, 3)
}

func vec4FromString(string str, string end, Vec4 *val) string {
	return fromString(str, end, &val.x, 4)
}

func matrixFromString(string str, string end, Matrix *val) string {
	return fromString(str, end, &val.m[0], 16)
}
