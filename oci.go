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
	OCI_NTV_SYNTAX = 1
)

const (
	OCI_ATTR_NUM_DML_ERRORS = 73
)

const (
	OCI_NO_DATA = 100
)

const (
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
	OCI_TYP_CHAR = 96

	// unsportod by this driver
	OCI_TYP_LONG          = 8
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
	SQLT_CHR = 1  // [n]byte
	SQLT_NUM = 2  // float64?
	SQLT_INT = 3  // int64
	SQLT_FLT = 4  // float64
	SQLT_STR = 5  // [n+1]byte
	SQLT_DAT = 12 // [7]byte
	SQLT_AFC = 96 // [n]char
	SQLT_RDD = 104
)
