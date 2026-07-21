// Package tuple provides fixed-size heterogeneous product types from 2 to 16 elements.
package tuple

// Tuple2 is a pair of values with types T1 and T2.
type Tuple2[T1 any, T2 any] struct {
	First  T1
	Second T2
}

// NewTuple2 creates a new Tuple2 instance with the given values
func NewTuple2[T1 any, T2 any](first T1, second T2) Tuple2[T1, T2] {
	return Tuple2[T1, T2]{
		First:  first,
		Second: second,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple2[T1, T2]) Values() (T1, T2) {
	return this.First, this.Second
}

// Tuple3 represents a 3-element tuple with generic types T1, T2, and T3
type Tuple3[T1 any, T2 any, T3 any] struct {
	First  T1
	Second T2
	Third  T3
}

// NewTuple3 creates a new Tuple3 instance with the given values
func NewTuple3[T1 any, T2 any, T3 any](first T1, second T2, third T3) Tuple3[T1, T2, T3] {
	return Tuple3[T1, T2, T3]{
		First:  first,
		Second: second,
		Third:  third,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple3[T1, T2, T3]) Values() (T1, T2, T3) {
	return this.First, this.Second, this.Third
}

// Tuple4 represents a 4-element tuple with generic types T1, T2, T3, and T4
type Tuple4[T1 any, T2 any, T3 any, T4 any] struct {
	First  T1
	Second T2
	Third  T3
	Fourth T4
}

// NewTuple4 creates a new Tuple4 instance with the given values
func NewTuple4[T1 any, T2 any, T3 any, T4 any](first T1, second T2, third T3, fourth T4) Tuple4[T1, T2, T3, T4] {
	return Tuple4[T1, T2, T3, T4]{
		First:  first,
		Second: second,
		Third:  third,
		Fourth: fourth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple4[T1, T2, T3, T4]) Values() (T1, T2, T3, T4) {
	return this.First, this.Second, this.Third, this.Fourth
}

// Tuple5 represents a 5-element tuple with generic types T1, T2, T3, T4, and T5
type Tuple5[T1 any, T2 any, T3 any, T4 any, T5 any] struct {
	First  T1
	Second T2
	Third  T3
	Fourth T4
	Fifth  T5
}

// NewTuple5 creates a new Tuple5 instance with the given values
func NewTuple5[T1 any, T2 any, T3 any, T4 any, T5 any](first T1, second T2, third T3, fourth T4, fifth T5) Tuple5[T1, T2, T3, T4, T5] {
	return Tuple5[T1, T2, T3, T4, T5]{
		First:  first,
		Second: second,
		Third:  third,
		Fourth: fourth,
		Fifth:  fifth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple5[T1, T2, T3, T4, T5]) Values() (T1, T2, T3, T4, T5) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth
}

// Tuple6 represents a 6-element tuple with generic types T1, T2, T3, T4, T5, and T6
type Tuple6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any] struct {
	First  T1
	Second T2
	Third  T3
	Fourth T4
	Fifth  T5
	Sixth  T6
}

// NewTuple6 creates a new Tuple6 instance with the given values
func NewTuple6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6) Tuple6[T1, T2, T3, T4, T5, T6] {
	return Tuple6[T1, T2, T3, T4, T5, T6]{
		First:  first,
		Second: second,
		Third:  third,
		Fourth: fourth,
		Fifth:  fifth,
		Sixth:  sixth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple6[T1, T2, T3, T4, T5, T6]) Values() (T1, T2, T3, T4, T5, T6) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth
}

// Tuple7 represents a 7-element tuple with generic types T1, T2, T3, T4, T5, T6, and T7
type Tuple7[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any] struct {
	First   T1
	Second  T2
	Third   T3
	Fourth  T4
	Fifth   T5
	Sixth   T6
	Seventh T7
}

// NewTuple7 creates a new Tuple7 instance with the given values
func NewTuple7[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7) Tuple7[T1, T2, T3, T4, T5, T6, T7] {
	return Tuple7[T1, T2, T3, T4, T5, T6, T7]{
		First:   first,
		Second:  second,
		Third:   third,
		Fourth:  fourth,
		Fifth:   fifth,
		Sixth:   sixth,
		Seventh: seventh,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple7[T1, T2, T3, T4, T5, T6, T7]) Values() (T1, T2, T3, T4, T5, T6, T7) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh
}

// Tuple8 represents an 8-element tuple with generic types T1, T2, T3, T4, T5, T6, T7, and T8
type Tuple8[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any] struct {
	First   T1
	Second  T2
	Third   T3
	Fourth  T4
	Fifth   T5
	Sixth   T6
	Seventh T7
	Eighth  T8
}

// NewTuple8 creates a new Tuple8 instance with the given values
func NewTuple8[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8) Tuple8[T1, T2, T3, T4, T5, T6, T7, T8] {
	return Tuple8[T1, T2, T3, T4, T5, T6, T7, T8]{
		First:   first,
		Second:  second,
		Third:   third,
		Fourth:  fourth,
		Fifth:   fifth,
		Sixth:   sixth,
		Seventh: seventh,
		Eighth:  eighth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple8[T1, T2, T3, T4, T5, T6, T7, T8]) Values() (T1, T2, T3, T4, T5, T6, T7, T8) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth
}

// Tuple9 represents a 9-element tuple with generic types T1, T2, T3, T4, T5, T6, T7, T8, and T9
type Tuple9[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any] struct {
	First   T1
	Second  T2
	Third   T3
	Fourth  T4
	Fifth   T5
	Sixth   T6
	Seventh T7
	Eighth  T8
	Ninth   T9
}

// NewTuple9 creates a new Tuple9 instance with the given values
func NewTuple9[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8, ninth T9) Tuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9] {
	return Tuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]{
		First:   first,
		Second:  second,
		Third:   third,
		Fourth:  fourth,
		Fifth:   fifth,
		Sixth:   sixth,
		Seventh: seventh,
		Eighth:  eighth,
		Ninth:   ninth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple9[T1, T2, T3, T4, T5, T6, T7, T8, T9]) Values() (T1, T2, T3, T4, T5, T6, T7, T8, T9) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth, this.Ninth
}

// Tuple10 represents a 10-element tuple with generic types T1, T2, T3, T4, T5, T6, T7, T8, T9, and T10
type Tuple10[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any] struct {
	First   T1
	Second  T2
	Third   T3
	Fourth  T4
	Fifth   T5
	Sixth   T6
	Seventh T7
	Eighth  T8
	Ninth   T9
	Tenth   T10
}

// NewTuple10 creates a new Tuple10 instance with the given values
func NewTuple10[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8, ninth T9, tenth T10) Tuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10] {
	return Tuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]{
		First:   first,
		Second:  second,
		Third:   third,
		Fourth:  fourth,
		Fifth:   fifth,
		Sixth:   sixth,
		Seventh: seventh,
		Eighth:  eighth,
		Ninth:   ninth,
		Tenth:   tenth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple10[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10]) Values() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth, this.Ninth, this.Tenth
}

// Tuple11 represents an 11-element tuple with generic types T1 through T11
type Tuple11[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any] struct {
	First    T1
	Second   T2
	Third    T3
	Fourth   T4
	Fifth    T5
	Sixth    T6
	Seventh  T7
	Eighth   T8
	Ninth    T9
	Tenth    T10
	Eleventh T11
}

// NewTuple11 creates a new Tuple11 instance with the given values
func NewTuple11[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8, ninth T9, tenth T10, eleventh T11) Tuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11] {
	return Tuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]{
		First:    first,
		Second:   second,
		Third:    third,
		Fourth:   fourth,
		Fifth:    fifth,
		Sixth:    sixth,
		Seventh:  seventh,
		Eighth:   eighth,
		Ninth:    ninth,
		Tenth:    tenth,
		Eleventh: eleventh,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple11[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11]) Values() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth, this.Ninth, this.Tenth, this.Eleventh
}

// Tuple12 represents a 12-element tuple with generic types T1 through T12
type Tuple12[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any] struct {
	First    T1
	Second   T2
	Third    T3
	Fourth   T4
	Fifth    T5
	Sixth    T6
	Seventh  T7
	Eighth   T8
	Ninth    T9
	Tenth    T10
	Eleventh T11
	Twelfth  T12
}

// NewTuple12 creates a new Tuple12 instance with the given values
func NewTuple12[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8, ninth T9, tenth T10, eleventh T11, twelfth T12) Tuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12] {
	return Tuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]{
		First:    first,
		Second:   second,
		Third:    third,
		Fourth:   fourth,
		Fifth:    fifth,
		Sixth:    sixth,
		Seventh:  seventh,
		Eighth:   eighth,
		Ninth:    ninth,
		Tenth:    tenth,
		Eleventh: eleventh,
		Twelfth:  twelfth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple12[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12]) Values() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth, this.Ninth, this.Tenth, this.Eleventh, this.Twelfth
}

// Tuple13 represents a 13-element tuple with generic types T1 through T13
type Tuple13[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any] struct {
	First      T1
	Second     T2
	Third      T3
	Fourth     T4
	Fifth      T5
	Sixth      T6
	Seventh    T7
	Eighth     T8
	Ninth      T9
	Tenth      T10
	Eleventh   T11
	Twelfth    T12
	Thirteenth T13
}

// NewTuple13 creates a new Tuple13 instance with the given values
func NewTuple13[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8, ninth T9, tenth T10, eleventh T11, twelfth T12, thirteenth T13) Tuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13] {
	return Tuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]{
		First:      first,
		Second:     second,
		Third:      third,
		Fourth:     fourth,
		Fifth:      fifth,
		Sixth:      sixth,
		Seventh:    seventh,
		Eighth:     eighth,
		Ninth:      ninth,
		Tenth:      tenth,
		Eleventh:   eleventh,
		Twelfth:    twelfth,
		Thirteenth: thirteenth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple13[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13]) Values() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth, this.Ninth, this.Tenth, this.Eleventh, this.Twelfth, this.Thirteenth
}

// Tuple14 represents a 14-element tuple with generic types T1 through T14
type Tuple14[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any] struct {
	First      T1
	Second     T2
	Third      T3
	Fourth     T4
	Fifth      T5
	Sixth      T6
	Seventh    T7
	Eighth     T8
	Ninth      T9
	Tenth      T10
	Eleventh   T11
	Twelfth    T12
	Thirteenth T13
	Fourteenth T14
}

// NewTuple14 creates a new Tuple14 instance with the given values
func NewTuple14[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8, ninth T9, tenth T10, eleventh T11, twelfth T12, thirteenth T13, fourteenth T14) Tuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14] {
	return Tuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]{
		First:      first,
		Second:     second,
		Third:      third,
		Fourth:     fourth,
		Fifth:      fifth,
		Sixth:      sixth,
		Seventh:    seventh,
		Eighth:     eighth,
		Ninth:      ninth,
		Tenth:      tenth,
		Eleventh:   eleventh,
		Twelfth:    twelfth,
		Thirteenth: thirteenth,
		Fourteenth: fourteenth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple14[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14]) Values() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth, this.Ninth, this.Tenth, this.Eleventh, this.Twelfth, this.Thirteenth, this.Fourteenth
}

// Tuple15 represents a 15-element tuple with generic types T1 through T15
type Tuple15[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any] struct {
	First      T1
	Second     T2
	Third      T3
	Fourth     T4
	Fifth      T5
	Sixth      T6
	Seventh    T7
	Eighth     T8
	Ninth      T9
	Tenth      T10
	Eleventh   T11
	Twelfth    T12
	Thirteenth T13
	Fourteenth T14
	Fifteenth  T15
}

// NewTuple15 creates a new Tuple15 instance with the given values
func NewTuple15[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8, ninth T9, tenth T10, eleventh T11, twelfth T12, thirteenth T13, fourteenth T14, fifteenth T15) Tuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15] {
	return Tuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]{
		First:      first,
		Second:     second,
		Third:      third,
		Fourth:     fourth,
		Fifth:      fifth,
		Sixth:      sixth,
		Seventh:    seventh,
		Eighth:     eighth,
		Ninth:      ninth,
		Tenth:      tenth,
		Eleventh:   eleventh,
		Twelfth:    twelfth,
		Thirteenth: thirteenth,
		Fourteenth: fourteenth,
		Fifteenth:  fifteenth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple15[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15]) Values() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth, this.Ninth, this.Tenth, this.Eleventh, this.Twelfth, this.Thirteenth, this.Fourteenth, this.Fifteenth
}

// Tuple16 represents a 16-element tuple with generic types T1 through T16
type Tuple16[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any] struct {
	First      T1
	Second     T2
	Third      T3
	Fourth     T4
	Fifth      T5
	Sixth      T6
	Seventh    T7
	Eighth     T8
	Ninth      T9
	Tenth      T10
	Eleventh   T11
	Twelfth    T12
	Thirteenth T13
	Fourteenth T14
	Fifteenth  T15
	Sixteenth  T16
}

// NewTuple16 creates a new Tuple16 instance with the given values
func NewTuple16[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any](first T1, second T2, third T3, fourth T4, fifth T5, sixth T6, seventh T7, eighth T8, ninth T9, tenth T10, eleventh T11, twelfth T12, thirteenth T13, fourteenth T14, fifteenth T15, sixteenth T16) Tuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16] {
	return Tuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]{
		First:      first,
		Second:     second,
		Third:      third,
		Fourth:     fourth,
		Fifth:      fifth,
		Sixth:      sixth,
		Seventh:    seventh,
		Eighth:     eighth,
		Ninth:      ninth,
		Tenth:      tenth,
		Eleventh:   eleventh,
		Twelfth:    twelfth,
		Thirteenth: thirteenth,
		Fourteenth: fourteenth,
		Fifteenth:  fifteenth,
		Sixteenth:  sixteenth,
	}
}

// Values returns the tuple elements as individual values
func (this Tuple16[T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16]) Values() (T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) {
	return this.First, this.Second, this.Third, this.Fourth, this.Fifth, this.Sixth, this.Seventh, this.Eighth, this.Ninth, this.Tenth, this.Eleventh, this.Twelfth, this.Thirteenth, this.Fourteenth, this.Fifteenth, this.Sixteenth
}
