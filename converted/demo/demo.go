package main

HWND g_hWnd;
HDC g_hDC;
HGLRC g_hRC;
GLuint g_font_texture;
typedef char Path[MAX_PATH];
typedef unsigned int u32;
ofbx::IScene* g_scene = nullptr;

int getPropertyCount(ofbx::IElementProperty* prop)
{
	return prop ? getPropertyCount(prop->getNext()) + 1 : 0;
}


template <int N>
void catProperty(char(&out)[N], const ofbx::IElementProperty& prop)
{
	char tmp[128];
	switch (prop.getType())
	{
		case ofbx::IElementProperty::DOUBLE: sprintf_s(tmp, "%f", prop.getValue().toDouble()); break;
		case ofbx::IElementProperty::LONG: sprintf_s(tmp, "%" PRId64, prop.getValue().toU64()); break;
		case ofbx::IElementProperty::INTEGER: sprintf_s(tmp, "%d", prop.getValue().toInt()); break;
		case ofbx::IElementProperty::STRING: prop.getValue().toString(tmp); break;
		default: sprintf_s(tmp, "Type: %c", (char)prop.getType()); break;
	}
	strcat_s(out, tmp);
}


void showElement(const ofbx::IElement& parent)
{
	for (const ofbx::IElement* element = parent.getFirstChild(); element; element = element->getSibling())
	{
		auto id = element->getID();
		char label[128];
		id.toString(label);
		strcat_s(label, " (");
		ofbx::IElementProperty* prop = element->getFirstProperty();
		bool first = true;
		while (prop)
		{
			if (!first)
				strcat_s(label, ", ");
			first = false;
			catProperty(label, *prop);
			prop = prop->getNext();
		}
		strcat_s(label, ")");

		ImGui::PushID((const void*)id.begin);
		ImGuiTreeElementFlags flags = g_selected_element == element ? ImGuiTreeElementFlags_Selected : 0;
		if (!element->getFirstChild()) flags |= ImGuiTreeElementFlags_Leaf;
		if (ImGui::TreeElementEx(label, flags))
		{
			if (ImGui::IsItemHovered() && ImGui::IsMouseClicked(0)) g_selected_element = element;
			if (element->getFirstChild()) showGUI(*element);
			ImGui::TreePop();
		}
		else
		{
			if (ImGui::IsItemHovered() && ImGui::IsMouseClicked(0)) g_selected_element = element;
		}
		ImGui::PopID();
	}
}


template <typename T>
void showArray(const char* label, const char* format, ofbx::IElementProperty& prop)
{
	if (!ImGui::CollapsingHeader(label)) return;

	int count = prop.getCount();
	ImGui::Text("Count: %d", count);
	std::vector<T> tmp;
	tmp.resize(count);
	prop.getValues(&tmp[0], int(sizeof(tmp[0]) * tmp.size()));
	for (T v : tmp)
	{
		ImGui::Text(format, v);
	}
}


void showProp(ofbx::IElementProperty& prop)
{
	ImGui::PushID((void*)&prop);
	char tmp[256];
	switch (prop.getType())
	{
		case ofbx::IElementProperty::LONG: ImGui::Text("Long: %" PRId64, prop.getValue().toU64()); break;
		case ofbx::IElementProperty::FLOAT: ImGui::Text("Float: %f", prop.getValue().toFloat()); break;
		case ofbx::IElementProperty::DOUBLE: ImGui::Text("Double: %f", prop.getValue().toDouble()); break;
		case ofbx::IElementProperty::INTEGER: ImGui::Text("Integer: %d", prop.getValue().toInt()); break;
		case ofbx::IElementProperty::ARRAY_FLOAT: showArray<float>("float array", "%f", prop); break;
		case ofbx::IElementProperty::ARRAY_DOUBLE: showArray<double>("double array", "%f", prop); break;
		case ofbx::IElementProperty::ARRAY_INT: showArray<int>("int array", "%d", prop); break;
		case ofbx::IElementProperty::ARRAY_LONG: showArray<ofbx::u64>("long array", "%" PRId64, prop); break;
		case ofbx::IElementProperty::STRING:
			toString(prop.getValue(), tmp);
			ImGui::Text("String: %s", tmp);
			break;
		default:
			ImGui::Text("Other: %c", (char)prop.getType());
			break;
	}

	ImGui::PopID();
	if (prop.getNext()) showGUI(*prop.getNext());
}

type GUI struct {
	*render.Sprite
	scene *ofbx.Scene
}

func (g *GUI) Draw(buff draw.Image) {
	g.DrawOffset(buff, 0, 0)
}

func (g *GUI) DrawOffset(buff draw.Image, xOff, yOff float64) {

	var selected *ofbx.Element

	root := g_scene.GetRootElement()
	if root != nil && root.GetFirstChild() != nil {
		selected = g.showElement(root, buff)
	}

	if selected != nil {
		prop := selected.GetFirstProperty()
		if prop != nil {
			g.showProp(prop, buff)
		}
	}

	g.showObjectsGUI(buff)
	return g.Sprite.DrawOffset(buff, xOff, yOff)
}

func (g *GUI) showObjectsGUI(buff draw.Image) {
	root := scene.GetRoot()
	if root != nil {
		g.showObjectGUI(root, buff)
	}
	
	for _, anim := range scene.GetAnimationStacks() {
		g.showObjectGUI(anim, buff)
	}
}

func (g *GUI) showObjectGUI(object ofbx.Obj, buff draw.Image) {
	label := object.GetType().String()

	
	ImGuiTreeElementFlags flags = g_selected_object == &object ? ImGuiTreeElementFlags_Selected : 0;
	char tmp[128];
	sprintf_s(tmp, "%" PRId64 " %s (%s)", object.id, object.name, label);
	if (ImGui::TreeElementEx(tmp, flags))
	{
		if (ImGui::IsItemHovered() && ImGui::IsMouseClicked(0)) g_selected_object = &object;
		int i = 0;
		while (ofbx::Object* child = object.resolveObjectLink(i))
		{
			showObjectGUI(*child);
			++i;
		}
		ImGui::TreePop();
	}
	else
	{
		if (ImGui::IsItemHovered() && ImGui::IsMouseClicked(0)) g_selected_object = &object;
	}
}