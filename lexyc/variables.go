package lexyc

import (
	"strings"

	"go-custom-compiler/helpers"
	"go-custom-compiler/models"
)

//^[a-zA-Z]+[a-zA-Z0-9]*(\[[a-zA-Z0-9]+[a-zA-Z0-9]*\])*(\s*,\s*[a-zA-Z]+[a-zA-Z0-9]*(\[[a-zA-Z0-9]+[a-zA-Z0-9]*\])*)*:[a-zA-Z]+;$

//NextVariable ...
func (l *LexicalAnalyzer) NextVariable(currentLine string, lineIndex int64, debug bool) {
	// var moduleName string = "[variables.go][NextVariable()]"
	funcName := "[NextVariable()] "

	if l.CurrentBlockType == models.VARIABLEBLOCK {
		if l.R.RegexVariable.StartsWithVariableNoCheck(currentLine) {
			data := strings.Split(currentLine, " ")
			currentLine = ""
			for _, str := range data[1:] {
				currentLine += str + " "
			}
		}
		currentLine = strings.TrimSpace(currentLine)

		if l.R.RegexVariableAlfabetico.MatchVariableAlfabetico(currentLine) {
			_, variableData := getVariablesFromString(currentLine)
			for index, name := range variableData {
				groups := helpers.GetGroupMatches(name, helpers.ARRAYREGEXP)
				symbol := models.Token{Type: models.ALFABETICO, Key: groups[0]}
				l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{groups[0], helpers.IDENTIFICADOR}))
				for _, group := range groups[1:] {
					dim := l.FindSymbol(currentLine, lineIndex, group)
					if dim != nil {
						if dim.Type != models.ENTERO {
							l.LogError(lineIndex, "N/A", "UNEXPECTED", "Unexpected "+string(dim.Type)+" for dimension", currentLine)
						} else {
							symbol.Dimensions = append(symbol.Dimensions, dim.Value.(int))
						}
					}
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"[", helpers.DELIMITADOR}))
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{group, helpers.IDENTIFICADOR}))
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"]", helpers.DELIMITADOR}))
				}
				if index != len(variableData)-1 {
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{",", helpers.DELIMITADOR}))
				}

				if l.Context == "Global" {
					/* CHECK Verificar si variable local ya existe de manera global y/o local. (Mandar Error)*/
					if test := l.DoesTheTokenExistsInGlobalVariables(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of variable", currentLine)
					}
					if test := l.DoesTheTokenExistsInGlobalConstants(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of constant", currentLine)
					}
					if test := l.DoesTheTokenExistsInFunctionsStorage(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of function as variable", currentLine)
					}

					/* CHECK END */
					l.VariableStorage = append(l.VariableStorage, &symbol)
				} else {
					function := l.FindFunction(currentLine, lineIndex, l.Context)
					if function != nil {
						/* CHECK Verificar si variable local ya existe de manera global y/o local. (Mandar Error)*/
						if test := l.DoesTheTokenExistsInGlobalVariables(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of variable", currentLine)
						}
						if test := l.DoesTheTokenExistsInGlobalConstants(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of constant", currentLine)
						}
						if test := l.DoesTheTokenExistsInLocalVariables(&symbol, function); test {
							// //log.Printf("[ERR] Found redeclaration of local variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of local variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of local variable", currentLine)
						}
						if test := l.DoesTheTokenExistsInLocalParams(&symbol, function); test {
							// //log.Printf("[ERR] Found redeclaration of local function parameter at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of local function parameter at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of local function parameter", currentLine)
						}
						if test := l.DoesTheTokenExistsInFunctionsStorage(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of function as variable", currentLine)
						}
						/* CHECK END */
						function.Vars = append(function.Vars, &symbol)
					}
				}
			}
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{":", helpers.DELIMITADOR}))
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"alfabetico", helpers.PALABRARESERVADA}))
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{";", helpers.DELIMITADOR}))

			l.GL.Printf("%+v[VARIABLE] Alfabetico Found > %+v", funcName, currentLine)
			if debug {
				// //log.Printf("[VARIABLE] Alfabetico Found > %+v", currentLine)
			}
			return
		}

		if l.R.RegexVariableEntero.MatchVariableEntero(currentLine) {
			_, variableData := getVariablesFromString(currentLine)
			for index, name := range variableData {
				groups := helpers.GetGroupMatches(name, helpers.ARRAYREGEXP)
				symbol := models.Token{Type: models.ENTERO, Key: groups[0]}
				l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{groups[0], helpers.IDENTIFICADOR}))
				for _, group := range groups[1:] {
					dim := l.FindSymbol(currentLine, lineIndex, group)
					if dim != nil {
						if dim.Type != models.ENTERO {
							l.LogError(lineIndex, "N/A", "UNEXPECTED", "Unexpected "+string(dim.Type)+" for dimension", currentLine)
						} else {
							symbol.Dimensions = append(symbol.Dimensions, dim.Value.(int))
						}
					}
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"[", helpers.DELIMITADOR}))
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{group, helpers.IDENTIFICADOR}))
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"]", helpers.DELIMITADOR}))
				}
				if index != len(variableData)-1 {
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{",", helpers.DELIMITADOR}))
				}

				if l.Context == "Global" {

					/* CHECK Verificar si variable local ya existe de manera global y/o local. (Mandar Error)*/
					if test := l.DoesTheTokenExistsInGlobalVariables(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of variable", currentLine)
					}
					if test := l.DoesTheTokenExistsInGlobalConstants(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of constant", currentLine)
					}
					if test := l.DoesTheTokenExistsInFunctionsStorage(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of function as variable", currentLine)
					}
					/* CHECK END */

					l.VariableStorage = append(l.VariableStorage, &symbol)
				} else {
					function := l.FindFunction(currentLine, lineIndex, l.Context)
					if function != nil {

						/* CHECK Verificar si variable local ya existe de manera global y/o local. (Mandar Error)*/
						if test := l.DoesTheTokenExistsInGlobalVariables(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of variable", currentLine)
						}
						if test := l.DoesTheTokenExistsInGlobalConstants(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of constant", currentLine)
						}
						if test := l.DoesTheTokenExistsInLocalVariables(&symbol, function); test {
							// //log.Printf("[ERR] Found redeclaration of local variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of local variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of local variable", currentLine)
						}
						if test := l.DoesTheTokenExistsInLocalParams(&symbol, function); test {
							// //log.Printf("[ERR] Found redeclaration of local function parameter at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of local function parameter at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of local function parameter", currentLine)
						}
						if test := l.DoesTheTokenExistsInFunctionsStorage(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of function as variable", currentLine)
						}
						/* CHECK END */

						function.Vars = append(function.Vars, &symbol)
					}
				}
			}
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{":", helpers.DELIMITADOR}))
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"entero", helpers.PALABRARESERVADA}))
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{";", helpers.DELIMITADOR}))

			l.GL.Printf("%+v[VARIABLE] Entero Found > %+v", funcName, currentLine)
			if debug {
				// //log.Printf("[VARIABLE] Entero Found > %+v", currentLine)
			}
			return
		}

		if l.R.RegexVariableLogico.MatchVariableLogico(currentLine) {
			_, variableData := getVariablesFromString(currentLine)
			for index, name := range variableData {
				groups := helpers.GetGroupMatches(name, helpers.ARRAYREGEXP)
				symbol := models.Token{Type: models.LOGICO, Key: groups[0]}
				l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{groups[0], helpers.IDENTIFICADOR}))
				for _, group := range groups[1:] {
					dim := l.FindSymbol(currentLine, lineIndex, group)
					if dim != nil {
						if dim.Type != models.ENTERO {
							l.LogError(lineIndex, "N/A", "UNEXPECTED", "Unexpected "+string(dim.Type)+" for dimension", currentLine)
						} else {
							symbol.Dimensions = append(symbol.Dimensions, dim.Value.(int))
						}
					}
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"[", helpers.DELIMITADOR}))
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{group, helpers.IDENTIFICADOR}))
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"]", helpers.DELIMITADOR}))
				}
				if index != len(variableData)-1 {
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{",", helpers.DELIMITADOR}))
				}

				if l.Context == "Global" {
					/* CHECK Verificar si variable local ya existe de manera global y/o local. (Mandar Error)*/
					if test := l.DoesTheTokenExistsInGlobalVariables(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of variable", currentLine)
					}
					if test := l.DoesTheTokenExistsInGlobalConstants(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of constant", currentLine)
					}
					if test := l.DoesTheTokenExistsInFunctionsStorage(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of function as variable", currentLine)
					}
					/* CHECK END */
					l.VariableStorage = append(l.VariableStorage, &symbol)
				} else {
					function := l.FindFunction(currentLine, lineIndex, l.Context)
					if function != nil {
						/* CHECK Verificar si variable local ya existe de manera global y/o local. (Mandar Error)*/
						if test := l.DoesTheTokenExistsInGlobalVariables(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of variable", currentLine)
						}
						if test := l.DoesTheTokenExistsInGlobalConstants(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of constant", currentLine)
						}
						if test := l.DoesTheTokenExistsInLocalVariables(&symbol, function); test {
							// //log.Printf("[ERR] Found redeclaration of local variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of local variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of local variable", currentLine)
						}
						if test := l.DoesTheTokenExistsInLocalParams(&symbol, function); test {
							// //log.Printf("[ERR] Found redeclaration of local function parameter at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of local function parameter at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of local function parameter", currentLine)
						}
						if test := l.DoesTheTokenExistsInFunctionsStorage(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of function as variable", currentLine)
						}
						/* CHECK END */
						function.Vars = append(function.Vars, &symbol)
					}
				}
			}
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{":", helpers.DELIMITADOR}))
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"logico", helpers.PALABRARESERVADA}))
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{";", helpers.DELIMITADOR}))

			l.GL.Printf("%+v[VARIABLE] Logico Found > %+v", funcName, currentLine)
			if debug {
				// //log.Printf("[VARIABLE] Logico Found > %+v", currentLine)
			}
			return
		}

		if l.R.RegexVariableReal.MatchVariableReal(currentLine) {
			_, variableData := getVariablesFromString(currentLine)
			for index, name := range variableData {
				groups := helpers.GetGroupMatches(name, helpers.ARRAYREGEXP)
				symbol := models.Token{Type: models.REAL, Key: groups[0]}
				l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{groups[0], helpers.IDENTIFICADOR}))
				for _, group := range groups[1:] {
					dim := l.FindSymbol(currentLine, lineIndex, group)
					if dim != nil {
						if dim.Type != models.ENTERO {
							l.LogError(lineIndex, "N/A", "UNEXPECTED", "Unexpected "+string(dim.Type)+" for dimension", currentLine)
						} else {
							symbol.Dimensions = append(symbol.Dimensions, dim.Value.(int))
						}
					}
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"[", helpers.DELIMITADOR}))
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{group, helpers.IDENTIFICADOR}))
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"]", helpers.DELIMITADOR}))
				}
				if index != len(variableData)-1 {
					l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{",", helpers.DELIMITADOR}))
				}

				if l.Context == "Global" {
					/* CHECK Verificar si variable local ya existe de manera global y/o local. (Mandar Error)*/
					if test := l.DoesTheTokenExistsInGlobalVariables(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of variable", currentLine)
					}
					if test := l.DoesTheTokenExistsInGlobalConstants(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of constant", currentLine)
					}
					if test := l.DoesTheTokenExistsInFunctionsStorage(&symbol); test {
						// //log.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
						l.GL.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
						//"# Linea | # Columna | Error | Descripcion | Linea del Error"
						l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of function as variable", currentLine)
					}
					/* CHECK END */
					l.VariableStorage = append(l.VariableStorage, &symbol)
				} else {
					function := l.FindFunction(currentLine, lineIndex, l.Context)
					if function != nil {
						/* CHECK Verificar si variable local ya existe de manera global y/o local. (Mandar Error)*/
						if test := l.DoesTheTokenExistsInGlobalVariables(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of variable", currentLine)
						}
						if test := l.DoesTheTokenExistsInGlobalConstants(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of constant at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of constant", currentLine)
						}
						if test := l.DoesTheTokenExistsInLocalVariables(&symbol, function); test {
							// //log.Printf("[ERR] Found redeclaration of local variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of local variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of local variable", currentLine)
						}
						if test := l.DoesTheTokenExistsInLocalParams(&symbol, function); test {
							// //log.Printf("[ERR] Found redeclaration of local function parameter at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of local function parameter at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of local function parameter", currentLine)
						}
						if test := l.DoesTheTokenExistsInFunctionsStorage(&symbol); test {
							// //log.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
							l.GL.Printf("[ERR] Found redeclaration of function as variable at [%+v][Line: %+v]", 0, lineIndex)
							//"# Linea | # Columna | Error | Descripcion | Linea del Error"
							l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, 0, "REDECLARE", "Found redeclaration of function as variable", currentLine)
						}
						/* CHECK END */
						function.Vars = append(function.Vars, &symbol)
					}
				}
			}
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{":", helpers.DELIMITADOR}))
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{"real", helpers.PALABRARESERVADA}))
			l.LL.Println(helpers.IndentString(helpers.LEXINDENT, []string{";", helpers.DELIMITADOR}))

			l.GL.Printf("%+v[VARIABLE] Real Found > %+v", funcName, currentLine)
			if debug {
				// //log.Printf("[VARIABLE] Real Found > %+v", currentLine)
			}
			return
		}

		if l.R.RegexVariableDefault.MatchVariableDefault(currentLine) {
			// //log.Printf("BACKEND WARNING > ENTERED DEFAULT VARIABLE CASE")
			typeOfData, variableData := getVariablesFromString(currentLine)

			if l.R.RegexVariableAlfabetico.MatchVariableAlfabeticoCaseless(typeOfData) {
				for _, name := range variableData {
					l.VariableStorage = append(l.VariableStorage, &models.Token{Type: models.ALFABETICO, Key: name})
				}

				l.GL.Printf("%+v[VARIABLE] Alfabetico Found > %+v", funcName, currentLine)
				if debug {
					// //log.Printf("[VARIABLE] Alfabetico Found > %+v", currentLine)
				}

				foundTypo := false
				keyData := strings.Split(l.R.RegexVariableAlfabetico.Keyword, "")
				for i, char := range typeOfData {
					if i < len(keyData)-1 {
						if !foundTypo {
							if string(char) != keyData[i] {
								foundTypo = true
								// //log.Printf("[ERR] Found typo in '%+v' declaration at [%+v][Line: %+v]. Correct syntax should be '%+v'", typeOfData, i, lineIndex, l.R.RegexVariableAlfabetico.Keyword)
								l.GL.Printf("[ERR] Found typo in '%+v' declaration at [%+v][Line: %+v]. Correct syntax should be '%+v'", typeOfData, i, lineIndex, l.R.RegexVariableAlfabetico.Keyword)
								//"# Linea | # Columna | Error | Descripcion | Linea del Error"
								l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, i, typeOfData, l.R.RegexVariableAlfabetico.Keyword, currentLine)
							}
						}
					}
				}
			}

			if l.R.RegexVariableEntero.MatchVariableEnteroCaseless(typeOfData) {
				for _, name := range variableData {
					l.VariableStorage = append(l.VariableStorage, &models.Token{Type: models.ENTERO, Key: name})
				}

				l.GL.Printf("%+v[VARIABLE] Entero Found > %+v", funcName, currentLine)
				if debug {
					// //log.Printf("[VARIABLE] Entero Found > %+v", currentLine)
				}

				foundTypo := false
				keyData := strings.Split(l.R.RegexVariableEntero.Keyword, "")
				for i, char := range typeOfData {
					if i < len(keyData)-1 {
						if !foundTypo {
							if string(char) != keyData[i] {
								foundTypo = true
								// //log.Printf("[ERR] Found typo in '%+v' declaration at [%+v][Line: %+v]. Correct syntax should be '%+v'", typeOfData, i, lineIndex, l.R.RegexVariableEntero.Keyword)
								l.GL.Printf("[ERR] Found typo in '%+v' declaration at [%+v][Line: %+v]. Correct syntax should be '%+v'", typeOfData, i, lineIndex, l.R.RegexVariableEntero.Keyword)
								//"# Linea | # Columna | Error | Descripcion | Linea del Error"
								l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, i, typeOfData, l.R.RegexVariableEntero.Keyword, currentLine)
							}
						}
					}
				}
			}

			if l.R.RegexVariableLogico.MatchVariableLogicoCaseless(typeOfData) {
				for _, name := range variableData {
					l.VariableStorage = append(l.VariableStorage, &models.Token{Type: models.LOGICO, Key: name})
				}

				l.GL.Printf("%+v[VARIABLE] Logico Found > %+v", funcName, currentLine)
				if debug {
					// //log.Printf("[VARIABLE] Logico Found > %+v", currentLine)
				}

				foundTypo := false
				keyData := strings.Split(l.R.RegexVariableLogico.Keyword, "")
				for i, char := range typeOfData {
					if i < len(keyData)-1 {
						if !foundTypo {
							if string(char) != keyData[i] {
								foundTypo = true
								// //log.Printf("[ERR] Found typo in '%+v' declaration at [%+v][Line: %+v]. Correct syntax should be '%+v'", typeOfData, i, lineIndex, l.R.RegexVariableLogico.Keyword)
								l.GL.Printf("[ERR] Found typo in '%+v' declaration at [%+v][Line: %+v]. Correct syntax should be '%+v'", typeOfData, i, lineIndex, l.R.RegexVariableLogico.Keyword)
								//"# Linea | # Columna | Error | Descripcion | Linea del Error"
								l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, i, typeOfData, l.R.RegexVariableLogico.Keyword, currentLine)
							}
						}
					}
				}
			}

			if l.R.RegexVariableReal.MatchVariableRealCaseless(typeOfData) {
				for _, name := range variableData {
					l.VariableStorage = append(l.VariableStorage, &models.Token{Type: models.REAL, Key: name})
				}

				l.GL.Printf("%+v[VARIABLE] Real Found > %+v", funcName, currentLine)
				if debug {
					// //log.Printf("[VARIABLE] Real Found > %+v", currentLine)
				}

				foundTypo := false
				keyData := strings.Split(l.R.RegexVariableReal.Keyword, "")
				for i, char := range typeOfData {
					if i < len(keyData)-1 {
						if !foundTypo {
							if string(char) != keyData[i] {
								foundTypo = true
								// //log.Printf("[ERR] Found typo in '%+v' declaration at [%+v][Line: %+v]. Correct syntax should be '%+v'", typeOfData, i, lineIndex, l.R.RegexVariableReal.Keyword)
								l.GL.Printf("[ERR] Found typo in '%+v' declaration at [%+v][Line: %+v]. Correct syntax should be '%+v'", typeOfData, i, lineIndex, l.R.RegexVariableReal.Keyword)
								//"# Linea | # Columna | Error | Descripcion | Linea del Error"
								l.EL.Printf("%+v\t|\t%+v\t|\t%+v\t|\t%+v\t|\t%+v", lineIndex, i, typeOfData, l.R.RegexVariableReal.Keyword, currentLine)
							}
						}
					}
				}
			}

			return
		}

		l.GL.Printf("%+v Did not found any type of match on Line[%+v]! ", funcName, lineIndex)

	}

}

func getVariablesFromString(currentLine string) (string, []string) {
	currentLine = strings.TrimSuffix(currentLine, ";")
	currentLine = strings.TrimSuffix(currentLine, " ")
	lineData := splitAtCharRespectingQuotes(currentLine, ':')
	varType := lineData[1]
	variables := lineData[0]
	// variableData := strings.Split(variables, ",")
	variableData := splitAtCharRespectingQuotes(variables, ',')
	for i := range variableData {
		variableData[i] = strings.TrimSpace(variableData[i])
	}
	return varType, variableData
}
