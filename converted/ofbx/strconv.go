package ofbx


template <> const char* fromString<int>(const char* str, const char* end, int* val) {
	*val = atoi(str);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}

template <> const char* fromString<uint64>(const char* str, const char* end, uint64* val) {
	*val = strtoull(str, nullptr, 10);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}

template <> const char* fromString<int64>(const char* str, const char* end, int64* val) {
	*val = atoll(str);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}

template <> const char* fromString<double>(const char* str, const char* end, double* val) {
	*val = atof(str);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}

template <> const char* fromString<float>(const char* str, const char* end, float* val) {
	*val = (float)atof(str);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}

const char* fromString(const char* str, const char* end, double* val, int count) {
	const char* iter = str;
	for (int i = 0; i < count; ++i) {
		*val = atof(iter);
		++val;
		while (iter < end && *iter != ',') ++iter;
		if (iter < end) ++iter; // skip ','

		if (iter == end) return iter;

	}
	return (const char*)iter;
}

template <> const char* fromString<Vec2>(const char* str, const char* end, Vec2* val) {
	return fromString(str, end, &val.x, 2);
}

template <> const char* fromString<Vec3>(const char* str, const char* end, Vec3* val) {
	return fromString(str, end, &val.x, 3);
}

template <> const char* fromString<Vec4>(const char* str, const char* end, Vec4* val) {
	return fromString(str, end, &val.x, 4);
}

template <> const char* fromString<Matrix>(const char* str, const char* end, Matrix* val) {
	return fromString(str, end, &val.m[0], 16);
}