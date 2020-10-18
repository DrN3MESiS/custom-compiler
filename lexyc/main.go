package lexyc

import (
	"bufio"
	"fmt"
	"go-custom-compiler/helpers"
	"go-custom-compiler/models"
	"go-custom-compiler/regex"
	"log"
	"regexp"
	"strings"

	"github.com/DrN3MESiS/pprnt"
)

//LexicalAnalyzer ...
type LexicalAnalyzer struct {
	File *bufio.Scanner     //File
	R    *regex.CustomRegex //Regex Handler
	EL   *log.Logger        //Error Logger
	LL   *log.Logger        //Lex Logger
	GL   *log.Logger        //General Logger

	//TEST
	CurrentBlockType models.BlockType
	ParentBlockType  models.BlockType
	BlockQueue       []models.BlockType
	OpQueue          []models.TokenComp
	ConstantStorage  []models.Token
	VariableStorage  []models.Token
}

//NewLexicalAnalyzer ...
func NewLexicalAnalyzer(file *bufio.Scanner, ErrorLogger, LexLogger, GeneralLogger *log.Logger) (*LexicalAnalyzer, error) {
	var moduleName string = "[Lexyc][NewLexicalAnalyzer()]"

	if file == nil {
		GeneralLogger.Printf("[ERR]%+v file is not present", moduleName)
		return nil, fmt.Errorf("[ERR]%+v file is not present", moduleName)
	}
	if ErrorLogger == nil || LexLogger == nil || GeneralLogger == nil {
		GeneralLogger.Printf("[ERR]%+v Loggers are not present", moduleName)
		return nil, fmt.Errorf("[ERR]%+v Loggers are not present", moduleName)
	}
	R, err := regex.NewRegex(ErrorLogger, LexLogger, GeneralLogger)
	if err != nil {
		GeneralLogger.Printf("[ERR]%+v %+v", moduleName, err.Error())
		return nil, fmt.Errorf("[ERR]%+v %+v", moduleName, err.Error())
	}

	ErrorLogger.Printf("=============================================================")
	ErrorLogger.Printf("# Linea | # Columna | Error | Descripcion | Linea del Error")
	ErrorLogger.Printf("=============================================================")

	return &LexicalAnalyzer{
		File: file,
		R:    R,
		EL:   ErrorLogger,
		LL:   LexLogger,
		GL:   GeneralLogger,

		ParentBlockType:  models.NULLBLOCK,
		BlockQueue:       []models.BlockType{},
		CurrentBlockType: models.NULLBLOCK,
		OpQueue:          []models.TokenComp{},
		ConstantStorage:  []models.Token{},
		VariableStorage:  []models.Token{},
	}, nil
}

//Analyze ...
func (l *LexicalAnalyzer) Analyze(debug bool) error {
	funcName := "[Analyze()]"
	var lineIndex int64 = 1
	for l.File.Scan() {
		currentLine := l.File.Text()
		// l.GL.Printf("%+v Analyzing line: %+v", funcName, lineIndex)

		if len(currentLine) == 0 {
			l.GL.Printf("%+v Skipped [Line: %+v]; Reason: Empty", funcName, lineIndex)
			lineIndex++

			continue
		}
		var LastBlockState models.BlockType
		LastBlockState = l.CurrentBlockType
		/* Type Validation */
		isComment, err := l.R.StartsWith("//", currentLine)
		if err != nil {
			l.GL.Printf("%+v[APP_ERR] %+v", funcName, err.Error())
			return fmt.Errorf("%+v[APP_ERR] %+v", funcName, err.Error())
		}

		if isComment {
			l.GL.Printf("%+vSkipping Comment at line %+v", funcName, lineIndex)
			if debug {
				log.Printf("Skipping Comment at line %+v", lineIndex)
			}
			lineIndex++

			continue
		}

		currentLine = strings.TrimSpace(currentLine)

		log.Printf("> %+v", l.BlockQueue)

		/* StartsWith */

		if l.R.RegexConstante.StartsWithConstante(currentLine, lineIndex) {
			l.CurrentBlockType = models.CONSTANTBLOCK
		}

		if l.R.RegexVariable.StartsWithVariable(currentLine, lineIndex) {
			l.CurrentBlockType = models.VARIABLEBLOCK
		}

		if l.R.RegexFuncionProto.StartsWithFuncionProto(currentLine, lineIndex) && l.ParentBlockType == models.NULLBLOCK {
			l.CurrentBlockType = models.FUNCTIONPROTOBLOCK
		}

		if l.R.RegexProcedureProto.StartsWithProcedureProto(currentLine, lineIndex) && l.ParentBlockType == models.NULLBLOCK {
			l.CurrentBlockType = models.PROCEDUREPROTOBLOCK
		}

		if l.R.RegexProcedure.StartsWithProcedure(currentLine, lineIndex) {
			if len(l.BlockQueue) > 0 {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to create new procedure without finalizing the last Function or Procedure", currentLine)
				l.BlockQueue = []models.BlockType{}
			}
			l.BlockQueue = append(l.BlockQueue, models.PROCEDUREBLOCK)
		}

		if l.R.RegexFunction.StartsWithFunction(currentLine, lineIndex) {
			if len(l.BlockQueue) > 0 {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to create new function without finalizing the last Function or Procedure", currentLine)
				l.BlockQueue = []models.BlockType{}
			}
			l.BlockQueue = append(l.BlockQueue, models.FUNCTIONBLOCK)
		}

		if l.R.RegexInicio.StartsWithInicio(currentLine, lineIndex) {
			if len(l.BlockQueue) == 0 {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to initialize something outside of a Block", currentLine)
			}

			switch l.BlockQueue[len(l.BlockQueue)-1] {
			case models.INITBLOCK:
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to initialize something when already initialized", currentLine)
				break
			case models.PROCEDUREBLOCK:
				l.BlockQueue = append(l.BlockQueue, models.INITBLOCK)
				l.GL.Printf("%+v Initialized a PROCEDUREBLOCK [Line: %+v]", funcName, lineIndex)
				break
			case models.FUNCTIONBLOCK:
				l.BlockQueue = append(l.BlockQueue, models.INITBLOCK)
				l.GL.Printf("%+v Initialized a FUNCTIONBLOCK [Line: %+v]", funcName, lineIndex)
				break

			default:
				break
			}
		}

		if l.R.RegexFinFunction.StartsWithFinDeFuncion(currentLine, lineIndex) {
			if len(l.BlockQueue) == 0 {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a FUNCTIONBLOCK outside of a FUNCTIONBLOCK", currentLine)
			}

			if l.BlockQueue[len(l.BlockQueue)-1] != models.INITBLOCK {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a FUNCTIONBLOCK that wasn't initialized", currentLine)
			}

			newArr, ok := helpers.RemoveFromQueue(l.BlockQueue, models.INITBLOCK)
			if ok {
				l.BlockQueue = newArr
			} else {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a FUNCTIONBLOCK that wasn't initialized", currentLine)
			}

			newArr, ok = helpers.RemoveFromQueue(l.BlockQueue, models.FUNCTIONBLOCK)
			if ok {
				l.BlockQueue = newArr
			} else {
				if helpers.QueueContainsBlock(l.BlockQueue, models.PROCEDUREBLOCK) {
					l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a FUNCTIONBLOCK:Inicio with a PROCEDUREBLOCK as parent", currentLine)
				} else {
					l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a FUNCTIONBLOCK outside of a FUNCTIONBLOCK", currentLine)
				}

			}
		}

		if l.R.RegexFinProcedure.StartsWithFinDeProcedimiento(currentLine, lineIndex) {
			if len(l.BlockQueue) == 0 {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a PROCEDUREBLOCK outside of a PROCEDUREBLOCK", currentLine)
			}

			newArr, ok := helpers.RemoveFromQueue(l.BlockQueue, models.INITBLOCK)
			if ok {
				l.BlockQueue = newArr
			} else {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a PROCEDUREBLOCK that wasn't initialized", currentLine)
			}

			newArr, ok = helpers.RemoveFromQueue(l.BlockQueue, models.PROCEDUREBLOCK)
			if ok {
				l.BlockQueue = newArr
			} else {
				if helpers.QueueContainsBlock(l.BlockQueue, models.FUNCTIONBLOCK) {
					l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a PROCEDUREBLOCK:Inicio with a FUNCTIONBLOCK as parent", currentLine)
				} else {
					l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a PROCEDUREBLOCK outside of a PROCEDUREBLOCK", currentLine)
				}
			}
		}

		if l.R.RegexLoopRepetir.StartsWithRepetir(currentLine, lineIndex) {
			if len(l.BlockQueue) == 0 {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to create a REPEATBLOCK outside of a BLOCK", currentLine)
			}

			l.BlockQueue = append(l.BlockQueue, models.REPEATBLOCK)
			l.GL.Printf("%+v Initialized a REPEATBLOCK [Line: %+v]", funcName, lineIndex)
		}

		if l.R.RegexLoopHastaQue.StartsWithHastaQue(currentLine, lineIndex) {
			if len(l.BlockQueue) == 0 {
				l.LogError(lineIndex, "N/A", "N/A", "Attempted to end a REPEATBLOCK outside of a BLOCK", currentLine)
			}

			/* Analyze Params */

			data := strings.Split(currentLine, " ")
			currentLine = ""
			for _, str := range data[2:] {
				currentLine += str + " "
			}

			l.OpQueue = []models.TokenComp{}

			stage1 := regexp.MustCompile(`^(\s*)\((.*)\);(\s*)$`)
			stage1v2 := regexp.MustCompile(`^(\s*)\((.*)\)(\s*)$`)
			if stage1.MatchString(currentLine) || stage1v2.MatchString(currentLine) {
				currentLine = strings.TrimPrefix(currentLine, "(")
				currentLine = strings.TrimSuffix(currentLine, ";")
				currentLine = strings.TrimSuffix(currentLine, ")")

				lineData := strings.Split(currentLine, " ")
				switch len(lineData) {
				case 0:
					l.LogError(lineIndex, "N/A", "N/A", "Instruction 'Hasta que' doesn't have params", currentLine)
					break
				case 2:
					l.LogError(lineIndex, "N/A", "N/A", "Instruction 'Hasta que' only has 2 params", currentLine)
					break
				default:
					for _, dat := range lineData {
						if l.R.RegexCustom.MatchCteLog(dat, lineIndex) {
							l.OpQueue = append(l.OpQueue, models.CTELOG)
							continue
						}
						if l.R.RegexCustom.MatchCteEnt(dat) {
							l.OpQueue = append(l.OpQueue, models.CTEENT)
							continue
						}
						if l.R.RegexCustom.MatchCteAlfa(dat) {
							l.OpQueue = append(l.OpQueue, models.CTEALFA)
							continue
						}
						if l.R.RegexCustom.MatchCteReal(dat) {
							l.OpQueue = append(l.OpQueue, models.CTEREAL)
							continue
						}
						if l.R.RegexCustom.MatchOpArit(dat) {
							l.OpQueue = append(l.OpQueue, models.OPARIT)
							continue
						}
						if l.R.RegexCustom.MatchOpLog(dat) {
							l.OpQueue = append(l.OpQueue, models.OPLOG)
							continue
						}
						if l.R.RegexCustom.MatchOpRel(dat) {
							l.OpQueue = append(l.OpQueue, models.OPREL)
							continue
						}
						if l.R.RegexCustom.MatchIdent(dat) {
							l.OpQueue = append(l.OpQueue, models.ID)
							continue
						}
					}
					break
				}

				//TODO: Create Func to eval OPQueue

				// if len(l.OpQueue) > 0 {
				// 	log.Printf("OP Q > %+v", l.OpQueue)
				// }
			} else {
				l.LogError(lineIndex, "N/A", "N/A", "Instruction 'Hasta que' doesn't have params", currentLine)
			}

			/* End Analyze Params*/

			if l.BlockQueue[len(l.BlockQueue)-1] == models.REPEATBLOCK {
				newArr, ok := helpers.RemoveFromQueue(l.BlockQueue, models.REPEATBLOCK)
				if ok {
					l.BlockQueue = newArr
				} else {
					l.LogErrorGeneral(lineIndex, "N/A", "N/A", "I tried to delete something that was inside the slice that I saw before trying to delete", currentLine)
				}
			} else {
				l.LogError(lineIndex, "N/A", "N/A", fmt.Sprintf("Attempted to end a REPEATBLOCK before finalizing a %+v", l.BlockQueue[len(l.BlockQueue)-1]), currentLine)
			}
		}

		if l.R.RegexIO.MatchImprimenl(currentLine, lineIndex) {
			if !l.R.RegexIO.MatchPC(currentLine, lineIndex) {
				l.LogError(lineIndex, len(currentLine)-1, ";", "Missing ';'", currentLine)
			}
		} else if l.R.RegexIO.MatchImprime(currentLine, lineIndex) {
			if !l.R.RegexIO.MatchPC(currentLine, lineIndex) {
				l.LogError(lineIndex, len(currentLine)-1, ";", "Missing ';'", currentLine)
			}
		}

		if l.R.RegexIO.MatchLee(currentLine, lineIndex) {
			if !l.R.RegexIO.MatchPC(currentLine, lineIndex) {
				l.LogError(lineIndex, len(currentLine)-1, ";", "Missing ';'", currentLine)
			}
		}

		//Logger
		l.RegisterBlockChange(LastBlockState, debug, funcName, lineIndex)

		/* Data Segregator */
		if l.CurrentBlockType == models.CONSTANTBLOCK {
			l.NextConstant(currentLine, lineIndex, debug)
		}

		if l.CurrentBlockType == models.VARIABLEBLOCK {
			l.NextVariable(currentLine, lineIndex, debug)
		}

		if l.CurrentBlockType == models.FUNCTIONPROTOBLOCK {
			l.NextFuncionProto(currentLine, lineIndex, debug)
		}

		if l.CurrentBlockType == models.PROCEDUREPROTOBLOCK {
			l.NextProcedureProto(currentLine, lineIndex, debug)
		}

		lineIndex++
	}

	return nil
}

//LogError ...
//"# Linea | # Columna | Error | Descripcion | Linea del Error"
func (l *LexicalAnalyzer) LogError(lineIndex int64, columnIndex interface{}, err string, description string, currentLine string) {
	log.Printf("[ERR] %+v [Line: %+v]", description, lineIndex)
	l.GL.Printf("[ERR] %+v [Line: %+v]", description, lineIndex)
	l.EL.Printf("%+v|%+v|%+v|%+v|%+v", lineIndex, columnIndex, err, description, currentLine)
}

//LogErrorGeneral ...
//"# Linea | # Columna | Error | Descripcion | Linea del Error"
func (l *LexicalAnalyzer) LogErrorGeneral(lineIndex int64, columnIndex interface{}, err string, description string, currentLine string) {
	log.Printf("[ERR] %+v [Line: %+v]", description, lineIndex)
	l.GL.Printf("[ERR] %+v [Line: %+v]", description, lineIndex)
}

//RegisterBlockChange ...
func (l *LexicalAnalyzer) RegisterBlockChange(LastBlockState models.BlockType, debug bool, funcName string, lineIndex int64) {
	if LastBlockState != l.CurrentBlockType {
		l.GL.Printf("%+vSwitched to %+v [%+v]", funcName, l.CurrentBlockType, lineIndex)
		if debug {
			log.Printf("Switched to %+v [%+v]", l.CurrentBlockType, lineIndex)
		}
	}
}

//Print ...
func (l *LexicalAnalyzer) Print() {
	log.SetFlags(0)
	if len(l.ConstantStorage) > 0 {
		log.Print("Constants: ")
		pprnt.Print(l.ConstantStorage)
		log.Print("\n")
	} else {
		log.Println("Constants: 0")
	}

	if len(l.VariableStorage) > 0 {
		log.Print("Variables: ")
		pprnt.Print(l.VariableStorage)
		log.Print("\n")
	} else {
		log.Println("Variables: 0")
	}

	log.SetFlags(log.LstdFlags)
}
