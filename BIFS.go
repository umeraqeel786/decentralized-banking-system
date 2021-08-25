package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//=====================
//bank : bank interface
//=====================
type Bank struct {
	identity cid.ClientIdentity
}

//==============================================================
// AccountWithNumber will store all the account with their Name
//so that we can check account of receiver very easily
//==============================================================
//var AccountWithNumber map[string]string

//=============================================
// Kyc : this Struct store kyc data of customer
//=============================================
type Kyc struct {
	ObjectType        string             `json:"doc_type"` //docType is used to distinguish the various types of objects in state database
	CustID            string             `json:"cust_id"`
	PersonalHash      string             `json:"personal_hash"`
	PersonalKycStatus KycStatus          `json:"kyc_status"`
	IsBlackList       bool               `json:"is_black_list"`
	AgentID           string             `json:"agent_id"`
	MSPID             string             `json:"msp_id"`
	Accounts          map[string]Account `json:"account_number_with_name"`
}

//=====================================================================
//BranchAccount: this struct will store Account number and branch code
//=====================================================================
type Account struct {
	ObjectType       string      `json:"doc_type"` //docType is used to distinguish the various types of objects in state database
	AccountNumber    string      `json:"account_number"`
	BranchCode       string      `json:"branch"`
	OwnerName        string      `json: "owner_name"`
	AccountType      AccountType `json:"account_type"`
	AccountKycStatus KycStatus   `json:"kyc_status"`
	BusinessHash     string      `json:"business_hash"`
}

//================================================
//certStruct:
//================================================
type CertDetail struct {
	ClientID    string
	ClientMSPID string
}

//=============================================
//Branch :
//============================================
type Branch struct {
	BranchCode    string `json:"branch_code"`
	BranchName    string `json:"branch_name"`
	BranchAddress string `json:"branch_address"`
	MSPID         string `json:"msp_id"`
}

//=============================================================
//Transaction : this strut store data related to IR transaction
//=============================================================
type Transaction struct {
	ObjectType         string            `json:"doc_type"` //docType is used to distinguish the various types of objects in state database
	TransactionID      string            `json:'transaction_id'`
	SenderAccount      string            `json:'sender_account'`
	SenderName         string            `json:'sender_id'`
	SenderBranchName   string            `json:'sender_branch_name'`
	ReceiverAccount    string            `json:'receiver_account'`
	ReceiverName       string            `json:'receiver_name'`
	ReceiverBranchName string            `json:'receiver_branch_name'`
	Created            string            `json:'created'`
	Amount             float64           `json :'amount'`
	Reference          string            `json:"reference"`
	Purpose            string            `json:'purpose'`
	TransactionStatus  TransactionStatus `json:'transaction_status'`
	Comment            string            `json:'comment'`
	Flagged            bool              `json:'flagged'`
	DocHash            string            `json:'doc_hash'`
	Riskrating         Riskrating        `json:'risk_rating'`
	Crimerelated       string            `json:'crime_related'`
	RecieverEDD        string            `json:'reciever_edd'`
}

//=====================================================================
// AccountTypes: This Struct will store Account type with their Limites
//=====================================================================
type AccountType struct {
	AccountTypeName string  `json: "account_type_name"`
	Limit           float64 `json:"limit"`
}

//=====================================================================================================================================================
// setCustomerNameWithAccountNo : this function will add customer Account No amd map it with customer Name for verification when sending IR Transaction
//=====================================================================================================================================================
func (k *Kyc) setCustomerNameWithAccountNo(stub shim.ChaincodeStubInterface, accNo string, accName string, branchCode string, accType string, buisenessHash string) (pb.Response, string, bool) {
	if k.Accounts == nil {
		k.Accounts = make(map[string]Account) //args[1], args[2], args[3], args[4], args[5]
	}
	// if AccountWithNumber == nil {
	// 	AccountWithNumber = make(map[string]string)
	// }
	accountType := AccountType{}
	accountTypeAsBytes, err := stub.GetState(accType)
	if err != nil {
		fmt.Println("Account Type Not Found")
		return shim.Error(`{"status": 500 , "message": "Account Type Not Found ` + string(err.Error()) + `"}`), "Account Not Found", false
	}
	json.Unmarshal(accountTypeAsBytes, &accountType)
	if accountType.AccountTypeName == "" {
		return shim.Error(`{"status": 500 , "message": "Account Not Found "}`), "Account  Type Not Found", false
	}
	mspID, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error(""), "can not find mspID", false
	}
	branchAsBytes, _ := stub.GetState(mspID + "_" + branchCode)
	branch := Branch{}
	json.Unmarshal(branchAsBytes, &branch)
	if branch.BranchCode == "" {
		return shim.Error(`{"Status : 500 , "message": "Please Provide correct Branch Code :  ` + branchCode + `}`), "branch Not Found", false
	}

	account := Account{
		ObjectType:       "account",
		AccountNumber:    accNo,
		OwnerName:        accName,
		BranchCode:       branch.BranchCode,
		AccountType:      accountType,
		BusinessHash:     buisenessHash,
		AccountKycStatus: PENDING,
	}
	k.Accounts[accNo] = account
	fmt.Println(k.Accounts)
	var AccountWithNumberFromState = make(map[string]string)
	asBytes, _ := stub.GetState("accNo")
	json.Unmarshal(asBytes, &AccountWithNumberFromState)
	//if AccountWithNumber[accNo] == "" {
	AccountWithNumberFromState[accNo] = accName + "," + k.CustID
	// for k, v := range AccountWithNumberFromState {
	// 	AccountWithNumber[k] = v
	// }
	ValAsBytes, _ := json.Marshal(AccountWithNumberFromState)
	//mspID , _ := stub.GetMSPID()

	err = stub.PutState("accNo", ValAsBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`), "", false
	}
	//}
	fmt.Printf(AccountWithNumberFromState[accNo])
	return shim.Success(nil), "", true

}

//======================================================================================
//getAccountList :	getAccountList is utility func which will return list of all account
//=======================================================================================
// func (b *Bank) getAccountDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	if len(args) != 1 {
// 		return shim.Error(`{"status":500 , "message":"Expecting 1 Arguments , Please Provide 1 Argument"}`)
// 	}
// 	valAsbytes, err := stub.GetState("accNo") //get the Customer from chaincode state
// 	if err != nil {
// 		return shim.Error(`{"status": 500, "message": "` + err.Error() + `"}`)
// 	}
// 	json.Unmarshal()
// 	return shim.Success(valAsbytes)

// }

//===============================================================
//setBusinessAccount : this func will add business account with kyc
//===============================================================
func (k Kyc) setBuisenessAccount(stub shim.ChaincodeStubInterface, accNo string, accName string, branchCode string, accType string, buisenessHash string, businessAcc map[string]interface{}) (pb.Response, string, bool) {
	businesAccountList := businessAcc["Customers"].([]interface{})
	for _, value := range businesAccountList {
		customer := value.(map[string]interface{})
		valueAsBytes, _ := stub.GetState(customer["CustID"].(string))
		kyc := Kyc{}
		json.Unmarshal(valueAsBytes, &kyc)
		if kyc.CustID == "" {
			str := `{"CustID Does not have Personal KYC     ` + customer["CustID"].(string) + `"}`
			return shim.Error(str), "CustID Does Not Found   " + customer["CustID"].(string), false

		}
		_, err1, isTrue := kyc.setCustomerNameWithAccountNo(stub, accNo, customer["AccountName"].(string), branchCode, accType, buisenessHash)
		if isTrue != true {
			return shim.Error(`{"` + err1 + `"}`), err1, false
		}
		writeKycToLedger(stub, kyc)

	}
	// k.setCustomerNameWithAccountNo(stub, accNo, accName, branchCode, "Business", buisenessHash, branchName)
	// writeKycToLedger(stub, k)

	return shim.Success(nil), "", true
}

// Checking whether the account exists in the ledger or not bassing account number

func (b *Bank) IsAccountExists(stub shim.ChaincodeStubInterface, args string) (bool, string) {
	var AccountFromState = make(map[string]string)
	accNoAsBytes, err := stub.GetState("accNo")
	if err != nil {
		return false, ""
	}
	json.Unmarshal(accNoAsBytes, &AccountFromState)
	if value, found := AccountFromState[args]; found {
		fmt.Println("Value::", value)
		custID := strings.Split(value, ",")
		return true, custID[1]
	}
	return false, ""
}

//============================================================================
//addUpdateBankAccount: this func takes custID and add or update Bank Account
//============================================================================
func (b *Bank) addUpdateBankAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		return shim.Error(`{"status": 500 , "message":"Expecting 5 Arguments , Please Provide 5 arguments"}`)
	}
	kyc := Kyc{}
	asBytes, _ := stub.GetState(args[0])
	json.Unmarshal(asBytes, &kyc)
	if kyc.CustID == "" {
		return shim.Error(`{"status":500 , "message" :"Sorry Customer Not Found , Please Provide Correct Customer ID"}`)
	}
	err, value := b.IsAccountExists(stub, args[1])
	if err == true {
		return shim.Error(`{"status": 404 , "message" : "Account Already Found and register with Account Holder ` + value + `"}`)
	}
	_, err1, isTrue := kyc.setCustomerNameWithAccountNo(stub, args[1], args[2], args[3], args[4], args[5])
	if isTrue != true {
		return shim.Error(`{"status": 500 , "message": "` + err1 + `"}`)

	}
	writeKycToLedger(stub, kyc)

	return shim.Success([]byte(`{"status":200 , "message: "Data Updated"}`))
}

//=========================================================================================================
//updateAccountLimit: this function will takes custid and account number then update limit of that account
//=========================================================================================================
func (b *Bank) updateAccountLimit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error(`{"status": 500 , "message" :"Expecting three (3) arguments"}`)
	}
	value, _ := stub.GetCreator()
	fmt.Println(string(value))
	customerKyc := Kyc{}
	customerKycAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(`{"status": 500 , "message" : "` + string(err.Error()) + `"}`)
	}
	// accountTypeStruct := AccountType{}
	json.Unmarshal(customerKycAsBytes, &customerKyc)
	account := Account{}
	account = customerKyc.Accounts[args[1]]
	if account.AccountNumber == "" {
		return shim.Error(`{"status": 404 , "message" : "Account Not Found"}`)

	}
	limit, err1 := strconv.ParseFloat(args[2], 64)
	if err1 != nil {
		return shim.Error(`{"status": 500 , "message": "Can Not Convert Limit to Number"}`)
	}

	account.AccountType.Limit = limit
	// accountTypeStruct = accountType
	customerKyc.Accounts[args[1]] = account
	valueAsBytes, _ := json.Marshal(customerKyc)
	err = stub.PutState(customerKyc.CustID, valueAsBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message" : "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
}

//======================================================================
//checkAccount: this function will retrun account name
//======================================================================
func (b *Bank) checkAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(`{"status" : 500 , "message" :"Please Provide 1 Argument"}`)
	}
	err, value := b.IsAccountExists(stub, args[0])
	if err == false {
		return shim.Error(`{"status": 404 , "message" : "Account Not Found"}`)
	} else {
		return shim.Success([]byte(`{"status": 200 , "data": "` + string(value) + `"}`))
	}

}

//======================================================================================================================
//args[0]:custID string,
//args[1] : hashofDoc string,
//args[2]: accountType string,
//args[3]:isBlackList bool,
// args[4]: accountNumber string
//args[5]: customerName string
//args[6] : Bank MSPID
//======================================================================================================================
func (b *Bank) addKyc(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) == 3 {
		value, _ := stub.GetState(args[0])

		if value != nil {
			return shim.Error(`{ "status":403 , "message":"Customer with same ID Exist  ` + args[0] + `  Please Provide Unique ID"}`)
		}
		isBlack, _ := strconv.ParseBool(args[2])
		mspid, err := cid.GetMSPID(stub)
		id, err := cid.GetID(stub)
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}
		newKyc := Kyc{
			CustID:            args[0],
			PersonalHash:      args[1],
			PersonalKycStatus: PENDING,
			IsBlackList:       isBlack,
			MSPID:             mspid + "_" + id,
		}
		writeKycToLedger(stub, newKyc)
	} else if len(args) == 9 {
		value, _ := stub.GetState(args[0])

		if value != nil {
			return shim.Error(`{ "status":403 , "message":"Customer with same ID Exist  ` + args[0] + `  Please Provide Unique ID"}`)
		}
		//to be checked here
		isExist, value1 := b.IsAccountExists(stub, args[4])
		if isExist == true {
			return shim.Error(`{"status": 404 , "message" :"Account Already Found and register with Account Holder ` + value1 + `"}`)
		}
		isBlack, _ := strconv.ParseBool(args[2])
		mspid, err := cid.GetMSPID(stub)
		id, err := cid.GetID(stub)
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}
		accountType := AccountType{}
		accountTypeAsBytes, _ := stub.GetState(args[3])
		json.Unmarshal(accountTypeAsBytes, &accountType)

		newKyc := Kyc{
			CustID:            args[0],
			PersonalHash:      args[1],
			PersonalKycStatus: PENDING,
			IsBlackList:       isBlack,
			MSPID:             mspid + "_" + id,
		}
		if args[3] == "Business" || args[3] == "Joint" || args[3] == "business" || args[3] == "joint" {

			fmt.Println(args[8])
			var businessAccounts map[string]interface{}
			if len(args[8]) != 0 {
				if err := json.Unmarshal([]byte(args[8]), &businessAccounts); err != nil {
					fmt.Println(businessAccounts)
					panic(err)
				}
				_, err1, isTrue := newKyc.setBuisenessAccount(stub, args[4], args[5], args[6], args[3], args[7], businessAccounts)
				if isTrue != true {
					return shim.Error(`{"status": 500 , "message": "` + err1 + `"}`)
				}
				_, err1, isTrue = newKyc.setCustomerNameWithAccountNo(stub, args[4], args[5], args[6], args[3], args[7])
				if isTrue != true {
					return shim.Error(`{"status": 500 , "message": "` + err1 + `"}`)
				}
				writeKycToLedger(stub, newKyc)

			} else {
				//to be checked here
				isExist, value1 := b.IsAccountExists(stub, args[4])
				if isExist == true {
					return shim.Error(`{"status": 404 , "message" :"Account Already Found and register with Account Holder ` + value1 + `"}`)
				}
				_, err1, isTrue := newKyc.setCustomerNameWithAccountNo(stub, args[4], args[5], args[6], args[3], args[7])
				if isTrue != true {
					return shim.Error(`{"status": 500 , "message": "` + err1 + `"}`)
				}
				writeKycToLedger(stub, newKyc)

			}

			//return shim.Error(`{ "status":500 , "message":"Please Provide Correct Name of account type"}`)

		} else if args[3] == "Individual" || args[3] == "individual" {

			_, err1, isTrue := newKyc.setCustomerNameWithAccountNo(stub, args[4], args[5], args[6], args[3], args[7])
			if isTrue == false {
				return shim.Error(`{"status": 500 , "message": "` + err1 + `"}`)

			}
			writeKycToLedger(stub, newKyc)
		} else {
			return shim.Error(`{"status": 500 , "message": "Please Provide Correct Account Type"}`)
		}

	} else {
		return shim.Error(`{ "status":500 , "message":"Incorrect number of arguments. Expecting 3 or 7"}`)
	}

	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
}

//===================================================================
//addBranch: this function will add new branch to ledger
// args[0]: BranchCode    string `json:"branch_code"`
//	args[1]:BranchName    string `json:"branch_name"`
//	args[2]:BranchAddress string `json:"branch_address"`
//	args[3]:MSPID         string `json:"msp_id"`
//=====================================================================
func (b *Bank) addBranch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error(`{"status": 500 , "message" : "Please Provide 3 arguments"}`)
	}
	mspID, _ := cid.GetMSPID(stub)

	branchKey := mspID + "_" + args[0]
	asBytes, _ := stub.GetState(branchKey)
	branch := Branch{}
	json.Unmarshal(asBytes, &branch)
	fmt.Println(branch.BranchCode)
	if branch.BranchCode != "" {
		return shim.Error(`{"status": 404 , "message" : "Branch code already Exist` + args[0] + `"}`)

	}
	mspId, _ := cid.GetMSPID(stub)
	branch = Branch{
		BranchCode:    mspId + "_" + args[0],
		BranchName:    args[1],
		BranchAddress: args[2],
		MSPID:         mspId,
	}
	asBytes, _ = json.Marshal(branch)
	err := stub.PutState(branch.BranchCode, asBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , message: "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status": 200 , "message": "Data Updated"}`))
}

//===================================================================
//updateBranchAddress: this func will change address of branch
//==================================================================
func (b *Bank) updateBranchAddress(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error(`{"status": 500 , "message" : "Please Provide 2 argument"}`)
	}
	mspID, _ := cid.GetMSPID(stub)
	branchKey := mspID + "_" + args[0]
	fmt.Println(branchKey)

	asBytes, _ := stub.GetState(branchKey)

	branch := Branch{}
	err := json.Unmarshal(asBytes, &branch)
	if err != nil {
		return shim.Error(`{"Status: 500 , "message" : "Branch not found"}`)
	}
	if branch.BranchCode == "" {
		return shim.Error(`{"status": 404 , "message" : "Please Provide correct Branch code` + args[0] + `"}`)

	}
	branch.BranchAddress = args[1]
	asBytes, _ = json.Marshal(branch)
	err = stub.PutState(branch.BranchCode, asBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message" : "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status": 200 , "message": "Data Updated" }`))
}

//==============================================================================
//getBranchDetail :
//============================================================================
func (b *Bank) getBranchDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(`{"status": 500 , "message" : "Please Provide 1 Argument"}`)
	}
	mspID, _ := cid.GetMSPID(stub)
	branchKey := mspID + "_" + args[0]
	fmt.Println(branchKey)

	asBytes, _ := stub.GetState(branchKey)
	branch := Branch{}
	// if err := json.Unmarshal(asBytes, &branch); err != nil {
	//	return shim.Error(`{"status": 404 , "message" : "Branch Not Found"}`)
	// }
	err := json.Unmarshal(asBytes, &branch)
	if err != nil {
		return shim.Error(`{"status": 404 , "message" : "Branch Not Found"}`)
	}
	asBytes, _ = json.Marshal(&branch)
	// response, _ := json.Marshal([]byte())

	return shim.Success([]byte(`{"status": 200 , "data":` + string(asBytes) + `}`))
}

//===========================================================
// writeKycToLedger: this func will write kyc to ledger state
//============================================================
func writeKycToLedger(stub shim.ChaincodeStubInterface, kyc Kyc) pb.Response {
	fmt.Println(kyc)
	kyc.ObjectType = "kyc"
	asBytes, _ := json.Marshal(kyc)

	err := stub.PutState(kyc.CustID, asBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))

}

//==================================================================================
//updateKycDocumentHash : this function store hash of updated document in blockchain
//==================================================================================
func (b *Bank) updateKycDocumentHash(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		//return shim.Error("Incorrect number of arguments. Expecting 2")
		return shim.Error(`{"status": 500 , "message": incorrect number of argument , Expecting 2"}`)

	}

	customerAsBytes, _ := stub.GetState(args[0])
	kyc := Kyc{}
	// val, ok, err := cid.GetAttributeValue(stub, "agent")
	// if err != nil {
	// 	return shim.Error(`{"status": 500 , "message": ` + err.Error() + `"}`)
	// }
	// if !ok {
	// 	return shim.Error("Client has no attribute Agent")
	// 	return shim.Error(`{"status": 500 , "message": Client has no attribute Agemt"}`)

	// }
	_, ok, err := cid.GetAttributeValue(stub, "userType")
	if err != nil {
		return shim.Error(`{"status" : 500 , "message" : "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 403 , "message":"Client has no attribute UserType"}`)

	}
	json.Unmarshal(customerAsBytes, &kyc)

	if kyc.CustID != "" {
		// kyc.AgentID = val
		kyc.PersonalHash = args[1]

		customerAsBytes, _ = json.Marshal(kyc)
		err := stub.PutState(args[0], customerAsBytes)
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}
		return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
	} else {
		return shim.Error(`{"status": 500 , "message": "Please Provide Valid Customer ID"}`)

	}
}

//=========================================================
// addToBlackList : this function add customer to blacklist
//=========================================================
func (b *Bank) addToBlackList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(`{"status": 500 , "message": "Incorrect number of arguments , Expecting 1"}`)

	}
	customerAsBytes, _ := stub.GetState(args[0])
	kyc := Kyc{}
	json.Unmarshal(customerAsBytes, &kyc)
	kyc.IsBlackList = true
	_, ok, err := cid.GetAttributeValue(stub, "userType")
	if err != nil {
		return shim.Error(`{"status": 404 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 403 , "message":"Client has no attribute UserType"}`)

	}
	//kyc.AgentID = val
	customerAsBytes, _ = json.Marshal(kyc)
	err = stub.PutState(args[0], customerAsBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status":200 ,"message":"Data Updated"}`))
}

//=======================================================================
//removeFromBlackList : This function will Remove Customer From black list
//========================================================================
func (b *Bank) removeFromBlackList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(`{"status": 500 , "message": "Incorrect number of arguments , Expecting 1"}`)

	}
	customerAsBytes, _ := stub.GetState(args[0])
	kyc := Kyc{}
	_, ok, err := cid.GetAttributeValue(stub, "userType")
	if err != nil {
		return shim.Error(`{"status": 500 , "meesage": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 403 , "message":"Client has no attribute UserType"}`)

	}
	// kyc.AgentID = val
	err = json.Unmarshal(customerAsBytes, &kyc)
	if err != nil {
		return shim.Error(`{"status": 500 , "message":"Custid not Found"}`)
	}
	kyc.IsBlackList = false
	customerAsBytes, _ = json.Marshal(kyc)
	err = stub.PutState(args[0], customerAsBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
}

//======================================================================
//getPersonalKycStatus:
//=======================================================================
func (b *Bank) updatedPersonalKycStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error(`{"status": 500 , "message": "Incorrect number of arguments , Expecting id of customer and Kyc Status}`)

	}
	customerAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	} else if customerAsBytes == nil {
		return shim.Error(`{"status": 500 , "message": "Customer id Not Found"}`)

	}
	kyc := Kyc{}
	err = json.Unmarshal(customerAsBytes, &kyc)
	if err != nil {
		return shim.Error(`{"status" : 500 , "message" : "Kyc Not Found"}`)
	}
	if args[1] == "approved" || args[1] == "Approved" {
		kyc.PersonalKycStatus = APPROVED
	} else if args[1] == "disapproved" || args[1] == "Disapproved" {
		kyc.PersonalKycStatus = DISAPPROVED
	} else if args[1] == "pending" {
		kyc.PersonalKycStatus = PENDING
	} else if args[1] == "rejected" {
		kyc.PersonalKycStatus = REJECTED
	} else {
		return shim.Error(`{"status": 500 , "message": "please provide correct Status(pending , approved , disaproved , rejected )"}`)
	}
	customerAsBytes, _ = json.Marshal(kyc)
	err = stub.PutState(args[0], customerAsBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status": 200 , "data":"Data Updated"}`))
}

//==============================================================
//checkStatus : this function will return personal kyc status of customer
//==============================================================
func (b *Bank) checkPersonalKycStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(`{"status": 500 , "message": "Incorrect number of arguments , Expecting id of customer to query"}`)

	}
	customerAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	} else if customerAsBytes == nil {
		return shim.Error(`{"status": 500 , "message": "Customer id Not Found"}`)

	}
	kyc := Kyc{}
	json.Unmarshal(customerAsBytes, &kyc)
	return shim.Success([]byte(`{"status": 200 , "data":"` + kyc.PersonalKycStatus + `"}`))
}

//===============================================================================
//checkStatusOfAccount : this function will return kyc status of customer Account
//================================================================================
func (b *Bank) checkStatusOfAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error(`{"status": 500 , "message": "Incorrect number of arguments , Expecting id of customer and Account number"}`)

	}
	customerAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	} else if customerAsBytes == nil {
		return shim.Error(`{"status": 404 , "message": "Customer id not Found"}`)

	}
	kyc := Kyc{}
	err = json.Unmarshal(customerAsBytes, &kyc)
	if err != nil {
		return shim.Error(`{"status": 404 , "message": "Customer id not Found"}`)

	}
	if kyc.Accounts[args[1]].AccountNumber == "" {
		return shim.Error(`{"status": 404 , "message": "Account Number does not belong to this Customer"}`)

	}
	return shim.Success([]byte(`{"status" : 200 , "data":"` + kyc.Accounts[args[1]].AccountKycStatus + `"}`))
}

//==================================================
//updateKycStatus:  this func will update kyc status
//==================================================
func (b *Bank) updateAccountKycStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error(`{"status": 500 , "message":"Incorrect Number of argument , Expecting 3 argument"}`)

	}
	customerAsBytes, _ := stub.GetState(args[0])
	kyc := Kyc{}
	err := json.Unmarshal(customerAsBytes, &kyc)

	if err != nil {
		// return shim.Error("Customer Id Does Not Found")
		return shim.Error(`{"status": 404 , "message": "Customer Id  Not Found"}`)

	}
	account := Account(kyc.Accounts[args[1]])
	if account.AccountType.AccountTypeName == "" {
		return shim.Error(`{"status": 404 , "message": "Customer Does Not have this Account "}`)

	}
	// json.Unmarshal(kyc.Accounts[args[1]], &branchAccount)
	if args[2] == "approved" || args[2] == "Approved" {
		account.AccountKycStatus = APPROVED
	} else if args[2] == "disapproved" || args[2] == "Disapproved" {
		account.AccountKycStatus = DISAPPROVED
	} else if args[2] == "pending" {
		account.AccountKycStatus = PENDING
	} else if args[2] == "rejected" {
		account.AccountKycStatus = REJECTED
	} else {
		return shim.Error(`{"status": 403 , "message": "Please Provide Valid Kyc Status (pending , rejected , Approved , Disapproved)"}`)

	}
	_, ok, err := cid.GetAttributeValue(stub, "userType")
	if err != nil {
		return shim.Error(`{"status" : 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 403 , "message":"Client has no attribute UserType"}`)

	}
	fmt.Println(account.BranchCode)
	// branchCode := cid.AssertAttributeValue(stub, "branchCode", account.BranchCode)
	// regionalCheck := cid.AssertAttributeValue(stub, "userType", "regional")
	// if branchCode != nil || regionalCheck != nil {
	// 	return shim.Error(`{"status": 403 , "message": "` + branchCode.Error() + "" + regionalCheck.Error() + `"}`)
	// }
	kyc.AgentID = ""
	kyc.Accounts[args[1]] = account
	customerAsBytes, _ = json.Marshal(kyc)
	err = stub.PutState(args[0], customerAsBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status":200 , "data":"Data Updated"}`))
}

//===================================================
//UpdateKYcOfAccount :
//===================================================
// func (ba  *BranchAccount)
//================================================================================================
//Query Customer's Kyc History
//This method will return the KYC data of the customer for banks to perform customer Due diligence
//=================================================================================================
func (b *Bank) queryCustomerKycHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error(`{"status": 500 , "message": "Incorrect number of arguments , Expecting 1"}`)
	}

	customerID := args[0]

	fmt.Printf("- start getHistoryKYCForCustomer: %s\n", customerID)
	//Retreive All the Kyc data of
	resultsIterator, err := stub.GetHistoryForKey(customerID)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	defer resultsIterator.Close()
	// buffer is a JSON array containing historic values for the customerKyc
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(`{"status": 500 , "message" : "` + string(err.Error()) + `"}`)
		}
		fmt.Println(response)

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		buffer.WriteString(string(response.Value))

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForCustomer returning:\n%s\n", buffer.String())
	if len(buffer.Bytes()) == 0 {
		return shim.Success([]byte(`{"status": 200 , "data": "No KYC History"}`))
	}
	return shim.Success([]byte(`{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`))
}

//====================================================================
// getPendingCustomer: this function will return Customers kyc data of
// KycStatus Pending , new and updated
//====================================================================
func (b *Bank) searchPendingCustomer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(`{"status" : 500 , "message": "` + string(err.Error()) + `"}`)
	}
	defer resultsIterator.Close()

	// Create buffer of results
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		kyc := Kyc{}
		// Unmarshal kyc
		json.Unmarshal(queryResponse.Value, &kyc)

		// Select customers according to the search query
		if kyc.PersonalKycStatus == "new" || kyc.PersonalKycStatus == "updated" || kyc.PersonalKycStatus == "pending" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Value\":")
			buffer.WriteString(string(queryResponse.Value))

			buffer.WriteString("}")
		}
		// }
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return shim.Success([]byte(`{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`))
	}
	return shim.Success([]byte(`{"status": 200 , "data": "No Pending KYC"}`))
}

//=======================================================================================
// getCustomerKycData : this function return current data of kyc related to passed custid
//=======================================================================================
func (b *Bank) getCustomerKycData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error(`{"status": 500 , "message": "Incorrect number of arguments, Expecting id of customer to query"}`)

	}

	valAsbytes, err := stub.GetState(args[0]) //get the Customer from chaincode state
	kyc := Kyc{}
	err = json.Unmarshal(valAsbytes, &kyc)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "Can not find Customer KYC"}`)

	}
	if kyc.CustID == "" {
		return shim.Error(`{"status": 500 , "message": "No KYC found for Customer id:  ` + kyc.CustID + `"}`)

	} else {
		valuesASbytes, err := json.Marshal(&kyc)
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		return shim.Success([]byte(`{"status": 200 , "data": ` + string(valuesASbytes) + `}`))
	}

}

//====================================================================================
//verifyJointHash: this func will check and verify joint account hash and kyc
//====================================================================================
// func verifyJointHash(stub shim.ChaincodeStubInterface, accounts map[string]interface{}) (bool, []string) {
// 	accountList := accounts["Customers"].([]interface{})
// 	var found = true
// 	var custID []string
// 	for _, value := range accountList {
// 		customer := value.(map[string]interface{})

// 		valueAsBytes, _ := stub.GetState(customer["CustID"].(string))
// 		kyc := Kyc{}
// 		json.Unmarshal(valueAsBytes, &kyc)
// 		fmt.Println("HASH::", customer["Hash"])
// 		fmt.Println("Personal HASH::", kyc.PersonalHash)
// 		if customer["Hash"] != kyc.PersonalHash {
// 			found = false
// 			custID = append(custID, customer["CustID"].(string))
// 		}
// 	}
// 	if !found {
// 		return false, custID
// 	} else {
// 		return true, custID
// 	}
// }

//=====================================================================================
//verifyPersonalHash : this func will check and verify personal hash
//=====================================================================================
// func verifyPersonalHash(stub shim.ChaincodeStubInterface, accounts map[string]interface{}) (bool, string) {

// }

//=====================================================================================
//verifyHash: this function will check and verify business account hash and kyc
//======================================================================================
func verifyHash(stub shim.ChaincodeStubInterface, accounts map[string]interface{}) (bool, []string) {
	accountList := accounts["Customers"].([]interface{})
	var found = true
	var custID []string

	for _, value := range accountList {
		customer := value.(map[string]interface{})

		valueAsBytes, _ := stub.GetState(customer["CustID"].(string))
		kyc := Kyc{}
		json.Unmarshal(valueAsBytes, &kyc)
		if kyc.IsBlackList == true {
			found = false
			custID = append(custID, kyc.CustID+"  Is in Black List")
		}
		fmt.Println("HASH::", customer["Hash"])
		fmt.Println("Personal HASH::", kyc.PersonalHash)
		if customer["Hash"] != kyc.PersonalHash {
			found = false
			custID = append(custID, customer["CustID"].(string)+" Personal Hash not Match")
		}
		if accounts["AccountType"] == "Individual" {
			if kyc.PersonalKycStatus != "approved" {
				found = false
				custID = append(custID, kyc.CustID+"  KYC not Approved")

			}
		}
	}

	if !found {
		return false, custID
	} else {
		return true, custID
	}
}

//=======================================================================================
//transfer: this function will initiate transfer transaction
//=======================================================================================
func (b *Bank) transferInitiate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 11 {
		return shim.Error(`{"status": 500 , "message" :"Please Provide 11 Arguments"}`)
	}
	var isOk = true
	var customerArray []string
	_, ok, err := cid.GetAttributeValue(stub, "userType")
	if err != nil {
		return shim.Error(`{"status": 404 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 403 , "message":"Client has no attribute UserType"}`)

	}
	var accountsForVerification map[string]interface{}
	var limit float64
	// var isOk bool
	// var customerArray []string
	if len(args[4]) != 0 {
		if err := json.Unmarshal([]byte(args[4]), &accountsForVerification); err != nil {
			return shim.Error(`{"status" 500 , "message" : "Please check   ` + args[4] + `"}`)
		}
	}
	if accountsForVerification["Customers"] != nil {
		isOk, customerArray = verifyHash(stub, accountsForVerification)
		fmt.Println(isOk, customerArray)
	}
	senderIsExist, senderCustID := b.IsAccountExists(stub, args[0])
	senderKyc := Kyc{}
	if senderIsExist == false {
		return shim.Error(`{"status": 500 , "message": "Sender Account not found"}`)
	}
	senderAsBytes, _ := stub.GetState(senderCustID)
	json.Unmarshal(senderAsBytes, &senderKyc)
	amount, err1 := strconv.ParseFloat(args[2], 64)
	if err1 != nil {
		return shim.Error(`{"status": 500 , "message": "Can Not Convert Amount to Number"}`)
	}
	receiverIsExist, receiverCustID := b.IsAccountExists(stub, args[1])
	if receiverIsExist == false {
		return shim.Error(`{"status": 500 , "message": "Receiver Account not found"}`)
	}

	receiverAsBytes, _ := stub.GetState(receiverCustID)
	receiverKyc := Kyc{}
	json.Unmarshal(receiverAsBytes, &receiverKyc)
	limit = senderKyc.Accounts[args[0]].AccountType.Limit
	if receiverKyc.IsBlackList == true {
		return shim.Error(`{"status" :  500 , "message": "Receiver is BlackList"}`)
	}
	fmt.Println(limit, receiverKyc.CustID)
	if accountsForVerification["AccountType"] == "Business" {
		if senderKyc.Accounts[args[0]].BusinessHash == accountsForVerification["BusinessHash"] && senderKyc.Accounts[args[0]].AccountKycStatus == "approved" {
			if isOk == true {
				if amount <= limit {
					//senderAcc , senderName , senderBranch , amount , purpose , receiverAcc , receiverName , receiverBranch , reference , transactionStatusFlag
					writeTransactionToLedger(stub, senderKyc.Accounts[args[0]].AccountNumber, senderKyc.Accounts[args[0]].OwnerName, senderKyc.Accounts[args[0]].BranchCode, amount, args[3], receiverKyc.Accounts[args[1]].AccountNumber, receiverKyc.Accounts[args[1]].OwnerName, receiverKyc.Accounts[args[1]].BranchCode, args[5], "A", args[6], args[7], args[8], args[9], args[10])
				} else {
					writeTransactionToLedger(stub, senderKyc.Accounts[args[0]].AccountNumber, senderKyc.Accounts[args[0]].OwnerName, senderKyc.Accounts[args[0]].BranchCode, amount, args[3], receiverKyc.Accounts[args[1]].AccountNumber, receiverKyc.Accounts[args[1]].OwnerName, receiverKyc.Accounts[args[1]].BranchCode, args[5], "i", args[6], args[7], args[8], args[9], args[10])
				}
			} else {
				return shim.Error(`{"status": 500 , "message": ` + strings.Join(customerArray, ",") + `}`)
			}
		} else {
			return shim.Error(`{"status": 403 , "message": "Business Hash Not Match"}`)
		}
	} else if accountsForVerification["AccountType"] == "Joint" {
		if isOk == true && senderKyc.Accounts[args[0]].AccountKycStatus == "approved" {
			if amount <= limit {
				//senderAcc , senderName , senderBranch , amount , purpose , receiverAcc , receiverName , receiverBranch , reference , transactionStatusFlag
				writeTransactionToLedger(stub, senderKyc.Accounts[args[0]].AccountNumber, senderKyc.Accounts[args[0]].OwnerName, senderKyc.Accounts[args[0]].BranchCode, amount, args[3], receiverKyc.Accounts[args[1]].AccountNumber, receiverKyc.Accounts[args[1]].OwnerName, receiverKyc.Accounts[args[1]].BranchCode, args[5], "A", args[6], args[7], args[8], args[9], args[10])
			} else {
				writeTransactionToLedger(stub, senderKyc.Accounts[args[0]].AccountNumber, senderKyc.Accounts[args[0]].OwnerName, senderKyc.Accounts[args[0]].BranchCode, amount, args[3], receiverKyc.Accounts[args[1]].AccountNumber, receiverKyc.Accounts[args[1]].OwnerName, receiverKyc.Accounts[args[1]].BranchCode, args[5], "i", args[6], args[7], args[8], args[9], args[10])
			}
		} else {
			return shim.Error(`{"status":500 , "message":` + strings.Join(customerArray, ",") + `}`)
		}
	} else if accountsForVerification["AccountType"] == "Individual" {
		if isOk == true {
			if amount <= limit {
				//senderAcc , senderName , senderBranch , amount , purpose , receiverAcc , receiverName , receiverBranch , reference , transactionStatusFlag
				writeTransactionToLedger(stub, senderKyc.Accounts[args[0]].AccountNumber, senderKyc.Accounts[args[0]].OwnerName, senderKyc.Accounts[args[0]].BranchCode, amount, args[3], receiverKyc.Accounts[args[1]].AccountNumber, receiverKyc.Accounts[args[1]].OwnerName, receiverKyc.Accounts[args[1]].BranchCode, args[5], "A", args[6], args[7], args[8], args[9], args[10])
			} else {
				writeTransactionToLedger(stub, senderKyc.Accounts[args[0]].AccountNumber, senderKyc.Accounts[args[0]].OwnerName, senderKyc.Accounts[args[0]].BranchCode, amount, args[3], receiverKyc.Accounts[args[1]].AccountNumber, receiverKyc.Accounts[args[1]].OwnerName, receiverKyc.Accounts[args[1]].BranchCode, args[5], "i", args[6], args[7], args[8], args[9], args[10])
			}
		} else {
			fmt.Println("Return Status", isOk)
			fmt.Println("ACCC:", senderKyc.Accounts[args[0]].AccountKycStatus)
			return shim.Error(`{"status":500 , "message":"` + strings.Join(customerArray, ",") + `"}`)
		}
	} else {
		return shim.Error(`{"status": 500 , "message":"Please Provide Valid Account Type (Individual , Joint , Business)" `)
	}

	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
}

//============================================================================
//
//==============================================================================
// func
//===========================================================
// writeTransactionToLedger: this func will write Transaction to ledger state
//============================================================
func writeTransactionToLedger(stub shim.ChaincodeStubInterface, senderAcc string, senderName string, senderBranch string, amount float64, purpose string, receiverAcc string, receiverName string, receiverBranch string, reference string, transactionStatusFlag string, dochash string, timecreated string, riskrating string, crimerelated string, recieveredd string) pb.Response {
	txID := stub.GetTxID()
	fmt.Println(senderAcc, senderName, receiverAcc)
	//receiverKyc := Kyc{}
	// senderKyc := Kyc{}
	// err, receiverKyc := getKycOfBankAccount(stub, args[1])
	// if err == false {
	// 	return shim.Error(`{"status": 500 , "message": "Receiver Not Found"}`)

	// }
	// err, senderKyc = getKycOfBankAccount(stub, args[0])
	// if err == false {
	// 	return shim.Error(`{"status": 500 , "message": "Sender Account Not Found"}`)
	// }

	// fmt.Println(amount, senderKyc.KycStatus, receiverKyc.KycStatus, senderKyc.Accounts[args[0]].AccountType.Limit)
	// var txStatus TransactionStatus = INITIATED
	var transactionStatus TransactionStatus
	transaction := Transaction{}
	if transactionStatusFlag == "i" {
		transactionStatus = INITIATED

	} else {
		transactionStatus = ACCEPTEDSENDERBANK
	}
	var risk_rating Riskrating
	if riskrating == "low" || riskrating == "LOW" {
		risk_rating = LOW
	}
	if riskrating == "medium" || riskrating == "MEDIUM" {
		risk_rating = MEDIUM
	}
	if riskrating == "high" || riskrating == "HIGH" {
		risk_rating = HIGH
	}
	if riskrating == "" || riskrating == "" {
		risk_rating = UNKNOWN
	}
	transaction = Transaction{
		ObjectType:         "transaction",
		TransactionID:      txID,
		SenderName:         senderName,
		SenderAccount:      senderAcc,
		SenderBranchName:   senderBranch,
		ReceiverAccount:    receiverAcc,
		ReceiverName:       receiverName,
		ReceiverBranchName: receiverBranch,
		Amount:             amount,
		TransactionStatus:  transactionStatus,
		Created:            timecreated,
		Purpose:            purpose,
		Reference:          reference,
		Flagged:            true,
		DocHash:            dochash,
		Riskrating:         risk_rating,
		Crimerelated:       crimerelated,
		RecieverEDD:        recieveredd,
	}

	asBytes, _ := json.Marshal(transaction)

	err := stub.PutState(transaction.TransactionID, asBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status":200 , "message": "Data Updated" }`))

}

//===============================================
////// verifying Edd Document Hash from blockchain
//================================================
func (b *Bank) verifyEDD(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error(`{"status": 500 , "message": "Please provide 2 arguments Transaction Id and Hash"}`)
	}
	asBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	transaction := Transaction{}
	json.Unmarshal(asBytes, &transaction)
	if transaction.TransactionID == "" {
		return shim.Error(`{"status": 500 , "message": "Transaction not found"}`)
	}
	if transaction.DocHash == args[1] {
		return shim.Success([]byte(`{"status": 200 , "data": "true"}`))
	} else {
		return shim.Success([]byte(`{"status": 200 , "data": "false"}`))
	}
}

//===============================================
////// verifying Edd Document Hash from blockchain
//================================================
func (b *Bank) verifyReceiverEDD(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error(`{"status": 500 , "message": "Please provide 2 arguments Transaction Id and Hash"}`)
	}
	asBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	transaction := Transaction{}
	json.Unmarshal(asBytes, &transaction)
	if transaction.TransactionID == "" {
		return shim.Error(`{"status": 500 , "message": "Transaction not found"}`)
	}
	if transaction.RecieverEDD == args[1] {
		return shim.Success([]byte(`{"status": 200 , "data": "true"}`))
	} else {
		return shim.Success([]byte(`{"status": 200 , "data": "false"}`))
	}
}

//==========================================================================
// getPendingTransactionSenderBank::: of organization MSPID
//==========================================================================
func (b *Bank) getPendingTransactionSenderBank(stub shim.ChaincodeStubInterface) pb.Response {
	val, ok, err := cid.GetAttributeValue(stub, "branchCode")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 500 , "message": "Client has no Attribute BranchCode"}`)
	}
	_, ok, err = cid.GetAttributeValue(stub, "userType")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 403 , "message":"Client has no attribute UserType"}`)

	}
	mspId, _ := cid.GetMSPID(stub)
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	defer resultsIterator.Close()
	// mspId, _ := cid.GetMSPID(stub)
	//branchCode := cid.AssertAttributeValue(stub, "branchCode"+"_"+mspId, transaction.SenderBranchName)
	branchCode := mspId + "_" + val

	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		// Unmarshal kyc
		transaction := Transaction{}
		json.Unmarshal(queryResponse.Value, &transaction)

		// Select customers according to the search query
		//_, senderKyc := getKycOfBankAccount(stub, transaction.SenderAccount)
		senderAccountBranchCode := transaction.SenderBranchName
		fmt.Println(senderAccountBranchCode, val)
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		if senderAccountBranchCode == branchCode && transaction.TransactionStatus == "Initiate" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Value\":")
			buffer.WriteString(string(queryResponse.Value))

			buffer.WriteString("}")
		}

	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return shim.Success([]byte(`{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`))

	}
	fmt.Println(string(buffer.Bytes()))
	return shim.Success([]byte(`{"status": 200 , "data": "No Pending Transaction"}`))

}

//==========================================================================
// getPendingTransactionReceiverBank of organization MSPID
//==========================================================================
func (b *Bank) getPendingTransactionReceiverBank(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	val, ok, err := cid.GetAttributeValue(stub, "branchCode")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 500 , "message": "Client has no Attribute BranchCode"}`)
	}
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	defer resultsIterator.Close()
	mspId, _ := cid.GetMSPID(stub)
	//branchCode := cid.AssertAttributeValue(stub, "branchCode"+"_"+mspId, transaction.SenderBranchName)
	branchCode := mspId + "_" + val

	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(`{"status": 500 , "message" : "` + string(err.Error()) + `"}`)
		}

		// Unmarshal kyc
		transaction := Transaction{}
		json.Unmarshal(queryResponse.Value, &transaction)

		// Select customers according to the search query
		_, receiverKyc := getKycOfBankAccount(stub, transaction.ReceiverAccount)
		receiverAccountBranchCode := receiverKyc.Accounts[transaction.ReceiverAccount].BranchCode
		// fmt.Println(receiverAccountBranchCode, val)
		// receiverMspID := strings.Split(receiverAccountBranchCode, "_")
		// fmt.Println(receiverMspID[0])
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}
		if mspId == "HBLTR" {
			transaction.Amount = transaction.Amount / 28
		} else if mspId == "HBLPK" {
			transaction.Amount = transaction.Amount * 28
		}
		response, err := json.Marshal(transaction)

		if receiverAccountBranchCode == branchCode && transaction.TransactionStatus == "Sender Bank Approved" {
			if first == false {
				buffer.WriteString(",")
			}

			first = false
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Value\":")
			buffer.WriteString(string(response))

			buffer.WriteString("}")
		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return shim.Success([]byte(`{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`))

	}
	return shim.Success([]byte(`{"status": 200 , "data": "No Pending Transaction"}`))

}

//=================================================================================
// updateTransactionStatus: this func will update Transaction status at sender bank
//Args[0]: Transaction Id
//Args[1]: Status
//==================================================================================
func (b *Bank) updateTransactionStatusSender(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error(`{"status": 500 , "message":"Please Provide 3 arguments"}`)

	}
	val, ok, err := cid.GetAttributeValue(stub, "branchCode")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 500 , "message":"Client has no Attribute branchCode"}`)
	}

	transaction := Transaction{}
	transactionAsBytes, _ := stub.GetState(args[0])
	json.Unmarshal(transactionAsBytes, &transaction)
	// senderkyc := Kyc{}
	//_, _ := getKycOfBankAccount(stub, transaction.SenderAccount)
	// json.Unmarshal(senderKycAsBytes, &senderkyc)
	// json.Unmarshal(transactionAsBytes, &transaction)
	mspId, _ := cid.GetMSPID(stub)
	//branchCode := cid.AssertAttributeValue(stub, "branchCode"+"_"+mspId, transaction.SenderBranchName)
	branchCode := mspId + "_" + val
	if branchCode != transaction.SenderBranchName {
		return shim.Error(`{"status": 403 , "message":"You are not allowed to Change Transaction Status" }`)
	}
	_, ok, err = cid.GetAttributeValue(stub, "userType")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 500 , "message":"Client has no Attribute UserType"}`)
	}

	if args[1] == "rejected" || args[1] == "Rejected" {
		transaction.TransactionStatus = REJECTEDSENDERBANK
		transaction.Comment = args[2]
	} else if args[1] == "approved" || args[1] == "Approved" {
		transaction.TransactionStatus = ACCEPTEDSENDERBANK
		transaction.Comment = args[2]
	} else if args[1] == "pending" {
		transaction.TransactionStatus = PENDINGTRANSACTION
		transaction.Comment = args[2]
	} else {
		return shim.Error(`{"status": 500 , "message": "Please Provide Rejected , Approved  or pending Status"}`)
	}
	asBytes, _ := json.Marshal(transaction)
	err = stub.PutState(transaction.TransactionID, asBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	err = stub.SetEvent("evtsender", []byte(transaction.TransactionID))
	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
}

//=================================================================================
// CancelTransacation: this func will update Transaction status at sender bank
//Args[0]: Transaction Id
//Args[1]: Status
//==================================================================================
func (b *Bank) CancelTransacation(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error(`{"status": 500 , "message":"Please Provide 2 arguments"}`)

	}
	val, ok, err := cid.GetAttributeValue(stub, "branchCode")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 500 , "message":"Client has no Attribute branchCode"}`)
	}

	transaction := Transaction{}
	transactionAsBytes, _ := stub.GetState(args[0])
	json.Unmarshal(transactionAsBytes, &transaction)
	// senderkyc := Kyc{}
	//_, _ := getKycOfBankAccount(stub, transaction.SenderAccount)
	// json.Unmarshal(senderKycAsBytes, &senderkyc)
	// json.Unmarshal(transactionAsBytes, &transaction)
	if transaction.TransactionStatus == REJECTEDRECEIVERBANK || transaction.TransactionStatus == ACCEPTEDRECEIVERBANK {
		return shim.Error(`{"status": 403 , "message": "Can not change because transaction has been processed on receiver Bank"}`)
	}
	mspId, _ := cid.GetMSPID(stub)
	//branchCode := cid.AssertAttributeValue(stub, "branchCode"+"_"+mspId, transaction.SenderBranchName)
	branchCode := mspId + "_" + val
	if branchCode != transaction.SenderBranchName {
		return shim.Error(`{"status": 403 , "message":"You are not allowed to Change Transaction Status" }`)
	}
	_, ok, err = cid.GetAttributeValue(stub, "userType")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 500 , "message":"Client has no Attribute UserType"}`)
	}

	if strings.ToLower(args[1]) == "cancelled" {
		transaction.TransactionStatus = CANCELLED
	} else {
		return shim.Error(`{"status": 500 , "message": "Please Provide Status cancelled"}`)
	}
	asBytes, _ := json.Marshal(transaction)
	err = stub.PutState(transaction.TransactionID, asBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	err = stub.SetEvent("evtsender", []byte(transaction.TransactionID))
	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
}

//=======================================================================
//processPendingTransactionReceiver: this func will process pending transaction on receiver bank
// reject or accept it.
//=======================================================================
func (b *Bank) processPendingTransactionReceiver(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		return shim.Error(`{"status": 500 , "message":"Please Provide 4 arguments"}`)

	}
	val, ok, err := cid.GetAttributeValue(stub, "branchCode")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	if !ok {
		return shim.Error(`{"status": 500 , "message": "Client has no Attribute Branch Code"}`)
	}
	transaction := Transaction{}
	transactionAsBytes, _ := stub.GetState(args[0])
	json.Unmarshal(transactionAsBytes, &transaction)
	// account:= BranchAccount{}
	//_, _ = getKycOfBankAccount(stub, transaction.ReceiverAccount)
	// json.Unmarshal(receiverKyc, &branchAccount)
	mspId, _ := cid.GetMSPID(stub)

	branchCode := mspId + "_" + val
	if branchCode != transaction.ReceiverBranchName {
		return shim.Error(`{"status": 500 , "message": "You are not allowed To updated Status"}`)
	}
	if transaction.TransactionStatus == REJECTEDSENDERBANK {
		return shim.Error(`{"status": 504 , "message": "Sorry Sender Bank has Rejected this transaction"}`)
	}
	// if receiverKyc.Accounts[transaction.ReceiverAccount].MSPID
	if args[1] == "rejected" || args[1] == "Rejected" {
		transaction.TransactionStatus = REJECTEDRECEIVERBANK
		transaction.Comment = args[2]
	} else if args[1] == "Approved" || args[1] == "approved" {
		transaction.TransactionStatus = ACCEPTEDRECEIVERBANK
		transaction.Comment = args[2]
	} else if args[1] == "pending" {
		transaction.TransactionStatus = PENDINGTRANSACTION
		transaction.Comment = args[2]
	} else {
		return shim.Error(`{"status": 500 , "message" : "Please Provide approved , rejected or pending"}`)

	}
	transaction.RecieverEDD = args[3]
	transactionBytes, err := json.Marshal(transaction)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	err = stub.PutState(transaction.TransactionID, transactionBytes)
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
}

//============================================================================
//getRejectTransactionOfCustomer: this function will return reject transaction
//of customer which are rejected at customer bank branch
//
//============================================================================
func (b *Bank) getRejectTransactionOfSenderBank(stub shim.ChaincodeStubInterface) pb.Response {
	val, _, err := cid.GetAttributeValue(stub, "branchCode")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	mspId, _ := cid.GetMSPID(stub)
	branchCode := mspId + "_" + val
	fmt.Println(branchCode)
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		// Unmarshal kyc
		transaction := Transaction{}
		if err := json.Unmarshal(queryResponse.Value, &transaction); err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		// Select customers according to the search query
		//_, senderKyc := getKycOfBankAccount(stub, transaction.SenderAccount)
		senderAccountBranchCode := transaction.SenderBranchName
		fmt.Println("Sender Account::", senderAccountBranchCode, val)
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		if transaction.TransactionStatus == "Sender Bank Rejected" || transaction.TransactionStatus == "Receiver Bank Rejected" {

			if senderAccountBranchCode == branchCode {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString("{\"Key\":")
				buffer.WriteString("\"")
				buffer.WriteString(queryResponse.Key)
				buffer.WriteString("\"")

				buffer.WriteString(", \"Value\":")
				buffer.WriteString(string(queryResponse.Value))

				buffer.WriteString("}")
			}

		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return shim.Success([]byte(`{"status": 200 , "data":` + string(buffer.Bytes()) + `}`))
	}
	return shim.Success([]byte(`{"status" : 200 , "message": "No Rejected Transaction"}`))
}

//============================================================================
//getRejectTransactionOfCustomer: this function will return reject transaction
//of customer which are rejected at customer bank branch
//
//============================================================================
func (b *Bank) getRejectTransactionOfReceiverBank(stub shim.ChaincodeStubInterface) pb.Response {
	val, _, err := cid.GetAttributeValue(stub, "branchCode")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	mspId, _ := cid.GetMSPID(stub)
	branchCode := mspId + "_" + val
	fmt.Println(branchCode)
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		// Unmarshal kyc
		transaction := Transaction{}
		if err := json.Unmarshal(queryResponse.Value, &transaction); err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		// Select customers according to the search query
		//_, senderKyc := getKycOfBankAccount(stub, transaction.SenderAccount)
		receiverAccountBranchCode := transaction.ReceiverBranchName
		fmt.Println("Sender Account::", receiverAccountBranchCode, val)
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}
		if mspId == "HBLTR" {
			transaction.Amount = transaction.Amount / 28
		} else if mspId == "HBLPK" {
			transaction.Amount = transaction.Amount * 28
		}
		response, err := json.Marshal(transaction)

		if transaction.TransactionStatus == "Sender Bank Rejected" || transaction.TransactionStatus == "Receiver Bank Rejected" {

			if receiverAccountBranchCode == branchCode {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString("{\"Key\":")
				buffer.WriteString("\"")
				buffer.WriteString(queryResponse.Key)
				buffer.WriteString("\"")

				buffer.WriteString(", \"Value\":")
				buffer.WriteString(string(response))

				buffer.WriteString("}")
			}

		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return shim.Success([]byte(`{"status": 200 , "data":` + string(buffer.Bytes()) + `}`))

	}
	return shim.Success([]byte(`{"status" : 200 , "message": "No Rejected Transaction"}`))

}

// ==========================================================================
// getTransactionByID: this func will return transaction detail of transaction id
// ===========================================================================
func (b *Bank) getTransactionByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error(`{"status": 500 , "message":"Please Provide 1 argument"}`)
	}
	asBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	transaction := Transaction{}
	json.Unmarshal(asBytes, &transaction)
	if transaction.TransactionID == "" {
		return shim.Error(`{"status": 404 , "message": "transaction not found"}`)
	}
	return shim.Success([]byte(`{"status": 200 , "data": ` + string(asBytes) + `}`))

}

// =================================================================
// getNotificationFromReceiverBank = this func will check the send
// transaction of bank and return if any update made on receiver end.
//====================================================================
func (b *Bank) getNotificationFromReceiverBank(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	val, _, err := cid.GetAttributeValue(stub, "branchCode")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	mspId, _ := cid.GetMSPID(stub)
	branchCode := mspId + "_" + val
	fmt.Println(branchCode)
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		// Unmarshal kyc
		transaction := Transaction{}
		if err := json.Unmarshal(queryResponse.Value, &transaction); err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		// Select customers according to the search query
		//_, senderKyc := getKycOfBankAccount(stub, transaction.SenderAccount)
		senderAccountBranchCode := transaction.SenderBranchName
		fmt.Println("Sender Account::", senderAccountBranchCode, val)
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		if transaction.TransactionStatus == "Receiver Bank Accepted" || transaction.TransactionStatus == "Receiver Bank Rejected" {

			if senderAccountBranchCode == branchCode {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString("{\"Key\":")
				buffer.WriteString("\"")
				buffer.WriteString(queryResponse.Key)
				buffer.WriteString("\"")

				buffer.WriteString(", \"Value\":")
				buffer.WriteString(string(queryResponse.Value))

				buffer.WriteString("}")
			}

		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return shim.Success([]byte(`{"status": 200 , "data":` + string(buffer.Bytes()) + `}`))
	}
	return shim.Success([]byte(`{"status" : 200 , "message": "No updated Transaction"}`))
}

//getKycOfBankAccount: this function will return kyc of account provided
//=======================================================================
func getKycOfBankAccount(stub shim.ChaincodeStubInterface, accNo string) (bool, Kyc) {
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return false, Kyc{}
	}
	defer resultsIterator.Close()

	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {

		}
		kyc := Kyc{}
		// Unmarshal kyc
		json.Unmarshal(queryResponse.Value, &kyc)
		value := kyc.Accounts[accNo].AccountNumber
		// Select customers according to the search query
		if value != "" {
			if !first {
			}
			return true, kyc
		}
	}
	return false, Kyc{}

}

//========================================================================
// writeAccountTypeToLedger : this func will write account type to ledger
//========================================================================
func writeAccountTypeLedger(stub shim.ChaincodeStubInterface, accountTypes []AccountType) pb.Response {
	for i := 0; i < len(accountTypes); i++ {
		key := accountTypes[i].AccountTypeName
		chkBytes, _ := stub.GetState(key)
		if chkBytes == nil { //only add if it is not already present
			asBytes, _ := json.Marshal(accountTypes[i])
			err := stub.PutState(accountTypes[i].AccountTypeName, asBytes)
			if err != nil {
				return shim.Error(`{"status": 500 , "message" : "` + string(err.Error()) + `"}`)
			}
		} else {
			msg := " Account type with key:" + key + " already exists.. skipping ......."
			return shim.Error(`{"status": 500 , "message": "` + msg + `"}`)
		}
	}
	return shim.Success([]byte(`{"status":200 , "message":"Data Updated"}`))
}

//====================================================================
//getCertDetail: this function will return client(requester) details
//====================================================================
func (b *Bank) getCertDetail(stub shim.ChaincodeStubInterface) pb.Response {
	clientId, _ := cid.GetID(stub)
	clientMSPID, _ := cid.GetMSPID(stub)
	userType, _, _ := cid.GetAttributeValue(stub, "userType")
	branchCode, _, _ := cid.GetAttributeValue(stub, "branchCode")
	if clientId != "" {
		response := json.RawMessage([]byte(`{"status": 200 , "data": {"ClientID": "` + clientId + `" , "ClientMSPID": "` + clientMSPID + `" , "UserType":"` + userType + `" , "BranchCode": "` + branchCode + `"}}`))

		return shim.Success(response)
	} else {
		return shim.Error(`{"status": 404 , "message" : "Certificate not found" }`)
	}

}

// ================================================================================================
// getTransactionByAccountNumber : this function will return transaction of account number provided
// ================================================================================================
func (b *Bank) getTransactionByAccountNumber(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var accountNumber string
	var status TransactionStatus
	if len(args) == 2 {
		accountNumber = args[0]
		status = TransactionStatus(args[1])
	} else if len(args) == 1 {
		accountNumber = args[0]
		status = TransactionStatus("")
	}
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		// Unmarshal transaction
		transaction := Transaction{}
		if err := json.Unmarshal(queryResponse.Value, &transaction); err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}

		if err != nil {
			return shim.Error(`{"status": 500 , "message": "` + string(err.Error()) + `"}`)
		}
		if len(status) == 0 && transaction.ObjectType == "transaction" {
			if transaction.ReceiverAccount == accountNumber || transaction.TransactionStatus == status {
				fmt.Println(transaction)

				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString("{\"Key\":")
				buffer.WriteString("\"")
				buffer.WriteString(queryResponse.Key)
				buffer.WriteString("\"")

				buffer.WriteString(", \"Value\":")
				buffer.WriteString(string(queryResponse.Value))

				buffer.WriteString("}")

			}
		} else if transaction.ObjectType == "transaction" && len(status) > 0 {
			if transaction.ReceiverAccount == accountNumber && transaction.TransactionStatus == status {

				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString("{\"Key\":")
				buffer.WriteString("\"")
				buffer.WriteString(queryResponse.Key)
				buffer.WriteString("\"")

				buffer.WriteString(", \"Value\":")
				buffer.WriteString(string(queryResponse.Value))

				buffer.WriteString("}")
			}
		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return shim.Success([]byte(`{"status": 200 , "data":` + string(buffer.Bytes()) + `}`))
	}
	return shim.Success([]byte(`{"status" : 200 , "message": "No Transaction"}`))
}

//===================================
//Init : this function init Chaincode
//===================================
func (b *Bank) Init(stub shim.ChaincodeStubInterface) pb.Response {

	accountTypes := []AccountType{
		{AccountTypeName: "Individual", Limit: 20000},
		{AccountTypeName: "Business", Limit: 500000},
		{AccountTypeName: "Joint", Limit: 10000},
	}
	writeAccountTypeLedger(stub, accountTypes)
	return shim.Success(nil)
}

//=======================================
//Invoke : this function invoke Chaincode
//=======================================
func (b *Bank) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "addkyc" {
		return b.addKyc(stub, args)
	} else if function == "addBranch" {
		return b.addBranch(stub, args)
	} else if function == "updateBranchAddress" {
		return b.updateBranchAddress(stub, args)
	} else if function == "getBranchDetail" {
		return b.getBranchDetail(stub, args)
	} else if function == "checkStatusOfAccount" {
		return b.checkStatusOfAccount(stub, args)
	} else if function == "getCustomerKycData" {
		return b.getCustomerKycData(stub, args)
	} else if function == "updateKycDocumentHash" {
		return b.updateKycDocumentHash(stub, args)
	} else if function == "addToBlackList" {
		return b.addToBlackList(stub, args)
	} else if function == "removeFromBlackList" {
		return b.removeFromBlackList(stub, args)
	} else if function == "updateAccountLimit" {
		return b.updateAccountLimit(stub, args)
	} else if function == "checkPersonalKycStatus" {
		return b.checkPersonalKycStatus(stub, args)
	} else if function == "updateAccountKycStatus" {
		return b.updateAccountKycStatus(stub, args)
	} else if function == "queryCustomerKycHistory" {
		return b.queryCustomerKycHistory(stub, args)
	} else if function == "searchPendingCustomer" {
		return b.searchPendingCustomer(stub, args)
	} else if function == "checkAccount" {
		return b.checkAccount(stub, args)
	} else if function == "addUpdateBankAccount" {
		return b.addUpdateBankAccount(stub, args)
		// } else if function == "getAccountDetail" {
		// 	return b.getAccountList(stub, args)
		// } else if function == "getKycOfBankAccount" {
		// 	return b.getKycOfBankAccount(stub, args)
	} else if function == "transferInitiate" {
		return b.transferInitiate(stub, args)
	} else if function == "getPendingTransactionSenderBank" {
		return b.getPendingTransactionSenderBank(stub)
	} else if function == "updateTransactionStatusSender" {
		return b.updateTransactionStatusSender(stub, args)
	} else if function == "getPendingTransactionReceiverBank" {
		return b.getPendingTransactionReceiverBank(stub, args)
	} else if function == "processPendingTransactionReceiver" {
		return b.processPendingTransactionReceiver(stub, args)
	} else if function == "getRejectTransactionOfSenderBank" {
		return b.getRejectTransactionOfSenderBank(stub)
	} else if function == "getCertDetail" {
		return b.getCertDetail(stub)
	} else if function == "getRejectTransactionOfReceiverBank" {
		return b.getRejectTransactionOfReceiverBank(stub)
	} else if function == "updatedPersonalKycStatus" {
		return b.updatedPersonalKycStatus(stub, args)
	} else if function == "getTransactionByID" {
		return b.getTransactionByID(stub, args)
	} else if function == "getNotificationFromReceiverBank" {
		return b.getNotificationFromReceiverBank(stub, args)
	} else if function == "cancelTransacation" {
		return b.CancelTransacation(stub, args)
	} else if function == "getTransactionByAccountNumber" {
		return b.getTransactionByAccountNumber(stub, args)
	} else if function == "verifyEDD" {
		return b.verifyEDD(stub, args)
	} else if function == "verifyReceiverEDD" {
		return b.verifyReceiverEDD(stub, args)
	} else {
		return shim.Error(`{"status": 500 , "message": "Not smart contract function ......."}`)
	}
}

func main() {
	err := shim.Start(new(Bank))
	if err != nil {
		fmt.Println("Error occurred in starting chaincode: %s", err)
	} else {
		fmt.Printf("Chaincode started successfully")
	}
}
