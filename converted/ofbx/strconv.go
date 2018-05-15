package ofbx


string intFromString(str, end string, int* val) string {
	*val = atoi(str)
	iter := str
	for (iter < end && *iter != ',') ++iter
	if (iter < end) ++iter // skip ','
	return (string)iter
}

func uint64FromString(str, end string, uint64* val) string {
	*val = strtoull(str, nullptr, 10)
	iter := str
	for (iter < end && *iter != ',') ++iter
	if (iter < end) ++iter // skip ','
	return (string)iter
}

func int64FromString(str, end string, int64* val) string {
	*val = atoll(str)
	iter := str
	for (iter < end && *iter != ',') ++iter
	if (iter < end) ++iter // skip ','
	return (string)iter
}

func doubleFromString(str, end string, val *double) string {
	*val = atof(str)
	iter := str
	for (iter < end && *iter != ',') ++iter
	if (iter < end) ++iter // skip ','
	return (string)iter
}

func floatFromString(str, end string, float* val) string {
	*val = (float)atof(str)
	iter := str
	for (iter < end && *iter != ',') ++iter
	if (iter < end) ++iter // skip ','
	return (string)iter
}

func fromString(str, end string, double* val, int count) string {
	iter := str
	for i := 0; i < count; i++ {
		*val = atof(iter)
		++val
		for (iter < end && *iter != ',') ++iter
		if (iter < end) ++iter // skip ','

		if (iter == end) return iter
	}
	return (string)iter
}

func vec2FromString(string str, string end, Vec2* val) string {
	return fromString(str, end, &val.x, 2)
}

func vec3FromString(string str, string end, Vec3* val) string {
	return fromString(str, end, &val.x, 3)
}

func vec4FromString(string str, string end, Vec4* val) string {
	return fromString(str, end, &val.x, 4)
}

func matrixFromString(string str, string end, Matrix* val) string {
	return fromString(str, end, &val.m[0], 16)
}