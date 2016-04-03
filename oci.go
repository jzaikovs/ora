package ora

const (
	OCI_DEFAULT = 0
)

// http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci02bas.htm#g466063
const (
	OCI_HTYPE_ENV    = 1
	OCI_HTYPE_ERROR  = 2
	OCI_HTYPE_SVCCTX = 3
	OCI_HTYPE_STMT   = 4
)

const (
	OCI_BATCH_MODE        = 0x1
	OCI_COMMIT_ON_SUCCESS = 0x20
	OCI_BATCH_ERRORS      = 0x80
)

const (
	OCI_NTV_SYNTAX = 1
)

const (
	OCI_ATTR_NUM_DML_ERRORS = 73
)

const (
	OCI_DTYPE_LOB   = 50
	OCI_DTYPE_PARAM = 53
	OCI_DTYPE_ROWID = 54
)

const (
	OCI_ATTR_CHAR_USED = 285
	OCI_ATTR_CHAR_SIZE = 286
)

const (
	OCI_ATTR_DATA_SIZE   = 1
	OCI_ATTR_DATA_TYPE   = 2
	OCI_ATTR_DISP_SIZE   = 3
	OCI_ATTR_NAME        = 4
	OCI_ATTR_ROWID       = 19
	OCI_ATTR_FETCH_ROWID = 448
)

// internal data types: http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci03typ.htm#CEGGBDFC
const (
	OCI_TYP_VARCHAR = 1 // 40000byte, ora12 -> 32k
	OCI_TYP_NUMBER  = 2 // [21]byte, can be casted, float64, int64, string
	OCI_TYP_DATE    = 12
	OCI_TYP_ROWID   = 104 // this is strange, docs say it is 69, but in practice it is 104

	// working progress
	OCI_TYP_LONG = 8
	OCI_TYP_CHAR = 96

	// unsupported by this driver
	OCI_TYP_RAW           = 23
	OCI_TYP_LONG_RAW      = 24
	OCI_TYP_BINARY_FLOAT  = 100
	OCI_TYP_BINARY_DOUBLE = 100
	OCI_TYP_CLOB          = 112
	OCI_TYP_BLOB          = 113
	OCI_TYP_BFILE         = 114
	OCI_TYP_TIMESTAMP     = 180
)

// external data types http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci03typ.htm#LNOCI16271
const (
	SQLT_CHR  = 1  // [n]byte
	SQLT_NUM  = 2  // float64?
	SQLT_INT  = 3  // int64
	SQLT_FLT  = 4  // float64
	SQLT_STR  = 5  // [n+1]byte
	SQLT_LNG  = 8  // [n]char
	SQLT_DAT  = 12 // [7]byte
	SQLT_AFC  = 96 // [n]char
	SQLT_CLOB = 112
	SQLT_RDD  = 104
)

// http://docs.oracle.com/cd/E11882_01/appdev.112/e10646/oci02bas.htm#LNOCI16220
const (
	OCI_SUCCESS           = 0
	OCI_SUCCESS_WITH_INFO = 1
	OCI_NO_DATA           = 100
	OCI_ERROR             = -1
	OCI_INVALID_HANDLE    = -2
	OCI_NEED_DATA         = 99
	OCI_STILL_EXECUTING   = -3123
	OCI_CONTINUE          = -24200
	OCI_ROWCBK_DONE       = -24201
)

const (
	OCI_ONE_PIECE   = 0
	OCI_FIRST_PIECE = 1
	OCI_NEXT_PIECE  = 2
	OCI_LAST_PIECE  = 3
)

const (
	SQLCS_IMPLICIT = 1
	SQLCS_NCHAR    = 2
	SQLCS_EXPLICIT = 3
	SQLCS_FLEXIBLE = 4
	SQLCS_LIT_NULL = 5
)

const (
	OCI_LOB_READONLY      = 1
	OCI_LOB_READWRITE     = 2
	OCI_LOB_WRITEONLY     = 3
	OCI_LOB_APPENDONLY    = 4
	OCI_LOB_FULLOVERWRITE = 5
	OCI_LOB_FULLREAD      = 6
)

var (
	oci_OCIAttrGet         = ociLibrary.NewProc("OCIAttrGet")
	oci_OCIAttrSet         = ociLibrary.NewProc("OCIAttrSet")
	oci_OCIBindByPos       = ociLibrary.NewProc("OCIBindByPos")
	oci_OCIDefineByPos     = ociLibrary.NewProc("OCIDefineByPos")
	oci_OCIDescriptorAlloc = ociLibrary.NewProc("OCIDescriptorAlloc")
	oci_OCIDescriptorFree  = ociLibrary.NewProc("OCIDescriptorFree")
	oci_OCIEnvCreate       = ociLibrary.NewProc("OCIEnvCreate")
	oci_OCIErrorGet        = ociLibrary.NewProc("OCIErrorGet")
	oci_OCIHandleAlloc     = ociLibrary.NewProc("OCIHandleAlloc")
	oci_OCIHandleFree      = ociLibrary.NewProc("OCIHandleFree")
	oci_OCIInitialize      = ociLibrary.NewProc("OCIInitialize")
	oci_OCILogoff          = ociLibrary.NewProc("OCILogoff")
	oci_OCILogon           = ociLibrary.NewProc("OCILogon")
	oci_OCIParamGet        = ociLibrary.NewProc("OCIParamGet")
	oci_OCIRowidToChar     = ociLibrary.NewProc("OCIRowidToChar")
	oci_OCIStmtExecute     = ociLibrary.NewProc("OCIStmtExecute")
	oci_OCIStmtFetch2      = ociLibrary.NewProc("OCIStmtFetch2")
	oci_OCIStmtPrepare2    = ociLibrary.NewProc("OCIStmtPrepare2") // this allows statement caching
	oci_OCIStmtRelease     = ociLibrary.NewProc("OCIStmtRelease")
	oci_OCITransCommit     = ociLibrary.NewProc("OCITransCommit")
	oci_OCITransRollback   = ociLibrary.NewProc("OCITransRollback")
	oci_OCILobRead         = ociLibrary.NewProc("OCILobRead")
	oci_OCILobRead2        = ociLibrary.NewProc("OCILobRead2")
	oci_OCILobOpen         = ociLibrary.NewProc("OCILobOpen")
	oci_OCILobClose        = ociLibrary.NewProc("OCILobClose")
)
