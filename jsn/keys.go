package jsn

func Keys(b []byte) [][]byte {
	res := make([][]byte, 0, 20)

	s, e, d := 0, 0, 0

	var k []byte
	state := expectValue

	st := NewStack()
	ae := 0
	instr := false

	for i := 0; i < len(b); i++ {
		if state == expectObjClose || state == expectListClose {
			if b[i-1] != '\\' && b[i] == '"' {
				instr = !instr
			}
			if !instr {
				switch b[i] {
				case '{', '[':
					d++
				case '}', ']':
					d--
				}
			}
		}

		si := st.Peek()

		switch {
		case state == expectKey && si != nil && i >= si.ss:
			i = si.se + 1
			st.Pop()

		case state == expectKey && b[i] == '{':
			state = expectObjClose
			s = i
			d++

		case state == expectObjClose && d == 0 && b[i] == '}':
			state = expectKey
			if ae != 0 {
				st.Push(skipInfo{i, ae})
				ae = 0
			}
			e = i
			i = s

		case state == expectKey && b[i] == '"':
			state = expectKeyClose
			s = i

		case state == expectKeyClose && (b[i-1] != '\\' && b[i] == '"'):
			state = expectColon
			k = b[(s + 1):i]

		case state == expectColon && b[i] == ':':
			state = expectValue

		case state == expectValue && b[i] == '"':
			state = expectString
			s = i

		case state == expectString && (b[i-1] != '\\' && b[i] == '"'):
			e = i

		case state == expectValue && b[i] == '{':
			state = expectObjClose
			s = i
			d++

		case state == expectObjClose && d == 0 && b[i] == '}':
			state = expectKey
			e = i
			i = s

		case state == expectValue && b[i] == '[':
			state = expectListClose
			s = i
			d++

		case state == expectListClose && d == 0 && b[i] == ']':
			state = expectKey
			ae = i
			e = i
			i = s

		case state == expectValue && (b[i] >= '0' && b[i] <= '9'):
			state = expectNumClose
			s = i

		case state == expectNumClose &&
			((b[i] < '0' || b[i] > '9') &&
				(b[i] != '.' && b[i] != 'e' && b[i] != 'E' && b[i] != '+' && b[i] != '-')):
			e = i - 1

		case state == expectValue &&
			(b[i] == 'f' || b[i] == 'F' || b[i] == 't' || b[i] == 'T'):
			state = expectBoolClose
			s = i

		case state == expectBoolClose && (b[i] == 'e' || b[i] == 'E'):
			e = i

		case state == expectValue && b[i] == 'n':
			state = expectNull
			s = i

		case state == expectNull && (b[i-1] == 'l' && b[i] == 'l'):
			e = i
		}

		if e != 0 {
			if k != nil {
				res = append(res, k)
			}

			if state == expectListClose {
			loop:
				for j := i + 1; j < len(b); j++ {
					switch b[j] {
					case ' ', '\t', '\n':
						continue
					case '{':
						break loop
					}
					i = e
					break loop
				}
			}
			state = expectKey
			k = nil
			e = 0
		}

	}

	return res
}
