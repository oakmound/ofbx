package ofbx

import (
	"errors"
	"fmt"

	"github.com/oakmound/oak/alg/floatgeom"
)

// VertexDataMapping dictates how the vertex is mapped
type VertexDataMapping int

// VertexDataMapping Options
const (
	ByPolygonVertex VertexDataMapping = iota
	ByPolygon       VertexDataMapping = iota
	ByVertex        VertexDataMapping = iota
)

var vtxDataMapFromStrs = map[string]VertexDataMapping{
	"ByPolygonVertex": ByPolygonVertex,
	"ByPolygon":       ByPolygon,
	"ByVertex":        ByVertex,
	"ByVertice":       ByVertex,
}

// MaxUvs is the highest number of UVs allowed
const MaxUvs = 4

//Geometry is the base geometric shape objec that is implemented in forms such as meshes that dictate control point deformations
type Geometry struct {
	Object
	Skin *Skin

	Vertices, Normals, Tangents []floatgeom.Point3

	UVs                 [MaxUvs][]floatgeom.Point2
	Colors              []floatgeom.Point4
	Materials, oldVerts []int
	newVerts            []Vertex
	Faces               [][]int
}

func (g *Geometry) String() string {
	return g.stringPrefix("")
}

func (g *Geometry) stringPrefix(prefix string) string {
	s := prefix + "Geometry:" + fmt.Sprintf("%v", g.ID()) + "\n"
	if len(g.Vertices) != 0 {
		s += prefix + "Verts:"
		for i, v := range g.Vertices {
			if i != 0 {
				if i > 100 {
					s += "..."
					break
				}
				s += ","
			}
			s += fmt.Sprintf("%+v", v)
		}
		s += "\n"
	}
	if len(g.Normals) != 0 {
		s += prefix + "Norms:"
		for i, v := range g.Normals {
			if i != 0 {
				if i > 100 {
					s += "..."
					break
				}
				s += ","
			}
			s += fmt.Sprintf("%+v", v)
		}
		s += "\n"
	}
	if len(g.Tangents) != 0 {
		s += prefix + "Tangents:"
		for i, v := range g.Tangents {
			if i != 0 {
				if i > 100 {
					s += "..."
					break
				}
				s += ","
			}
			s += fmt.Sprintf("%+v", v)
		}
		s += "\n"
	}
	if len(g.Materials) != 0 {
		s += prefix + "Materials:"
		for i, v := range g.Materials {
			if i != 0 {
				if i > 100 {
					s += "..."
					break
				}
				s += ","
			}
			s += fmt.Sprintf("%v", v)
		}
		s += "\n"
	}
	if len(g.Colors) != 0 {
		s += prefix + "Colors:"
		for i, v := range g.Colors {
			if i != 0 {
				if i > 100 {
					s += "..."
					break
				}
				s += ","
			}
			s += fmt.Sprintf("%+v", v)
		}
		s += "\n"
	}

	if len(g.Faces) != 0 {
		s += prefix + "Faces:"
		for i, v := range g.Faces {
			if i != 0 {
				if i > 100 {
					s += "..."
					break
				}
				s += ", "
			}
			s += fmt.Sprintf("%v", v)
		}
		s += "\n"
	}

	s += prefix + "UVs:"
	for _, v := range g.UVs {
		if len(v) == 0 {
			continue
		}
		for i, v2 := range v {
			if i != 0 {
				if i > 100 {
					s += "..."
					break
				}
				s += ","
			}
			s += fmt.Sprintf("%+v", v2)
		}
		s += "\n"
	}
	s += prefix + "Skin:\n"
	s += g.Skin.stringPrefix(prefix + "\t")
	return s
}

// Vertex hey wit its a linked list of indices!.....
type Vertex struct {
	index int //should start as -1
	next  *Vertex
}

func (nv *Vertex) add(index int) {
	if nv.index == -1 {
		//TODO: change this cuz we aint implementing it this way. Really its checking if the newvertex exists...
		nv.index = index
	} else if nv.next != nil {
		nv.next.add(index)
	} else {
		nv.next = &Vertex{-1, nil}
		nv.next.index = index
	}
}

// NewGeometry makes a stub Geometry
func NewGeometry(scene *Scene, element *Element) *Geometry {
	g := &Geometry{}
	g.Object = *NewObject(scene, element)
	g.oldVerts = make([]int, 0)
	return g
}

// Type returns GEOMETRY
func (g *Geometry) Type() Type {
	return GEOMETRY
}

func (g *Geometry) triangulate(indices []int) []int {
	old := make([]int, 0)
	polyIdx := 0
	for i := 0; i < len(indices); i++ {
		idx := indices[i]
		if idx < 0 {
			idx = (-idx) - 1
		}

		if polyIdx <= 2 {
			g.oldVerts = append(g.oldVerts, idx)
			old = append(old, i)
		} else {
			g.oldVerts = append(g.oldVerts, indices[i-polyIdx])
			old = append(old, i-polyIdx)
			g.oldVerts = append(g.oldVerts, indices[i-1])
			old = append(old, i-1)
			g.oldVerts = append(g.oldVerts, idx)
			old = append(old, i)
		}
		polyIdx++
		if indices[i] < 0 {
			polyIdx = 0
		}
	}
	return old
}

func parseGeometry(scene *Scene, element *Element) (*Geometry, error) {
	if element.Properties == nil {
		return nil, errors.New("Geometry invalid")
	}
	geom := NewGeometry(scene, element)

	verticesProp := findChildProperty(element, "Vertices")
	if len(verticesProp) == 0 {
		return nil, errors.New("Geometry Vertices Missing")
	}

	polysProp := findChildProperty(element, "PolygonVertexIndex")
	if len(polysProp) == 0 {
		return nil, errors.New("Geometry Indicies missing")
	}
	vertices, err := parseDoubleVecDataVec3(verticesProp[0])
	if err != nil {
		return nil, err
	}
	origIndices, err := parseBinaryArrayInt(polysProp[0])
	if err != nil {
		return nil, err
	}

	geom.Faces = make([][]int, 0)
	curFace := []int{}
	//Parse out the polygons. List of vertex references with a negative value indicating the last vertex of a face.
	for _, v := range origIndices {
		if v < 0 {
			curFace = append(curFace, (v*-1)-1)
			geom.Faces = append(geom.Faces, curFace)
			curFace = []int{}
		} else {
			curFace = append(curFace, v)
		}
	}

	toOldIndices := geom.triangulate(origIndices)
	geom.Vertices = make([]floatgeom.Point3, len(geom.oldVerts))

	for i, vIdx := range geom.oldVerts {
		v := vertices[vIdx]
		geom.Vertices[i] = v
	}

	geom.newVerts = make([]Vertex, len(geom.Vertices))

	for i := 0; i < len(geom.oldVerts); i++ {
		old := geom.oldVerts[i]
		geom.newVerts[old].add(i)
	}

	layerMaterialElements := findChildren(element, "LayerElementMaterial")
	if len(layerMaterialElements) > 0 {
		mappingProp := findChildProperty(layerMaterialElements[0], "MappingInformationType")
		referenceProp := findChildProperty(layerMaterialElements[0], "ReferenceInformationType")
		if len(mappingProp) == 0 || len(referenceProp) == 0 {
			return nil, errors.New("Invalid LayerElementMaterial")
		}
		var err error
		tmp := make([]int, 0)

		if mappingProp[0].value.String() == "ByPolygon" &&
			referenceProp[0].value.String() == "IndexToDirect" {
			geom.Materials = make([]int, len(geom.Vertices)/3)
			for i := 0; i < len(geom.Vertices)/3; i++ {
				geom.Materials[i] = -1
			}

			indiciesProp := findChildProperty(layerMaterialElements[0], "Materials")
			if indiciesProp == nil {
				return nil, errors.New("Invalid LayerElementMaterial")
			}

			tmp, err = parseBinaryArrayInt(indiciesProp[0])
			if err != nil {
				return nil, err
			}

			tmpI := 0
			triCt := 0
			insertIdx := 0
			for poly := 0; poly < len(tmp); {
				poly++
				triCt, tmpI = getTriCountFromPoly(origIndices, tmpI)
				for i := 0; i < triCt; i++ {
					geom.Materials[insertIdx] = tmp[poly]
					insertIdx++
				}
			}
		} else {
			if mappingProp[0].value.String() != "AllSame" {
				return nil, errors.New("Mapping not supported")
			}
		}
	}

	for _, elem := range element.Children {
		if elem.ID.String() != "LayerElementUV" {
			continue
		}
		uvIdx := 0
		if len(elem.Properties) > 0 {
			uvIdx = int(elem.Properties[0].value.toInt32())
		}
		if uvIdx >= 0 && uvIdx < MaxUvs {
			tmp, tmpIndices, mapping, err := parseVertexDataVec2(elem, "UV", "UVIndex")
			if err != nil {
				return nil, err
			}
			if tmp != nil && len(tmp) > 0 {
				//uvs = [4]floatgeom.Point2{} //resize(tmpIndices.empty() ? tmp.size() : tmpIndices.size());
				geom.UVs[uvIdx] = splatVec2(mapping, tmp, tmpIndices, origIndices)
				remapVec2(&geom.UVs[uvIdx], toOldIndices)
			}
		}

	}

	layerTangentElems := findChildren(element, "LayerElementTangents")
	if len(layerTangentElems) > 0 {
		tans := findChildren(layerTangentElems[0], "Tangents")
		var tmp []floatgeom.Point3
		var tmpIndices []int
		var mapping VertexDataMapping
		var err error
		if len(tans) > 0 {
			tmp, tmpIndices, mapping, err = parseVertexDataVec3(layerTangentElems[0], "Tangents", "TangentsIndex")
		} else {
			tmp, tmpIndices, mapping, err = parseVertexDataVec3(layerTangentElems[0], "Tangent", "TangentIndex")
		}
		if err != nil {
			return nil, err
		}
		if tmp != nil && len(tmp) > 0 {
			geom.Tangents = splatVec3(mapping, tmp, tmpIndices, origIndices)
			remapVec3(&geom.Tangents, toOldIndices)
		}
	}

	layerColorElems := findChildren(element, "LayerElementColor")
	if len(layerColorElems) > 0 {
		tmp, tmpIndices, mapping, err := parseVertexDataVec4(layerColorElems[0], "Colors", "ColorIndex")
		if err != nil {
			return nil, err
		}
		if len(tmp) > 0 {
			geom.Colors = splatVec4(mapping, tmp, tmpIndices, origIndices)
			remapVec4(&geom.Colors, toOldIndices)
		}
	}

	layerNormalElems := findChildren(element, "LayerElementNormal")
	if len(layerNormalElems) > 0 {
		tmp, tmpIndices, mapping, err := parseVertexDataVec3(layerNormalElems[0], "Normals", "NormalsIndex")
		if err != nil {
			return nil, err
		}
		if len(tmp) > 0 {
			geom.Normals = splatVec3(mapping, tmp, tmpIndices, origIndices)
			remapVec3(&geom.Normals, toOldIndices)
		}
	}

	// Todo: undo / redo some work above to not require redoing vertices

	geom.Vertices = vertices

	return geom, nil
}
