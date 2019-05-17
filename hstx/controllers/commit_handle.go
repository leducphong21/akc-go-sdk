package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/Akachain/akc-go-sdk/common"
	"github.com/Akachain/akc-go-sdk/hstx/models"
	. "github.com/Akachain/akc-go-sdk/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/rs/xid"
)

type Commit models.Commit

//High secure transaction Commit handle
// ------------------- //
//Create Commit
func (commit *Commit) CreateCommit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		//Invalid arguments
		resErr := common.ResponseError{common.ERR2, common.ResCodeDict[common.ERR2]}
		return common.RespondError(resErr)
	}
	ProposalID := args[0]
	var admin *Admin
	adminbytes, err := Getalldata(stub, models.ADMINTABLE)

	admin = new(Admin)
	Adminlist := []*Admin{}

	for row_json_bytes := range adminbytes {
		admin = new(Admin)
		err = json.Unmarshal(row_json_bytes, admin)
		if err != nil {

			resErr := common.ResponseError{common.ERR6, common.ResCodeDict[common.ERR6]}
			return common.RespondError(resErr)
		}
		Adminlist = append(Adminlist, admin)
	}

	var quorumList = []Quorum{}
	quorumResutl := new(Quorum)
	commitResutl := new(Commit)

	//check ProposalID exist in Quorum
	queryStringQuorum := fmt.Sprintf("{\"selector\": {\"_id\": {\"$regex\": \"^Quorum_\"},\"ProposalID\": \"%s\"}}", ProposalID)
	resultsIterator, err := stub.GetQueryResult(queryStringQuorum)
	if err != nil {
		resErr := common.ResponseError{common.ERR4, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			resErr := common.ResponseError{common.ERR4, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())}
			return common.RespondError(resErr)
		}
		errQuo := json.Unmarshal(queryResponse.Value, quorumResutl)
		if errQuo != nil {
			//convert JSON eror
			resErr := common.ResponseError{common.ERR6, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR6], err.Error(), common.GetLine())}
			return common.RespondError(resErr)
		}
		quorumList = append(quorumList, *quorumResutl)
	}
	if quorumResutl.ProposalID == "" {
		resErr := common.ResponseError{common.ERR12, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR12], "", common.GetLine())}
		return common.RespondError(resErr)
	}

	//check Only Commit once
	queryStringCommit := fmt.Sprintf("{\"selector\": {\"_id\": {\"$regex\": \"^Commit_\"},\"ProposalID\": \"%s\"}}", ProposalID)
	resultsIterator, err = stub.GetQueryResult(queryStringCommit)
	if err != nil {
		resErr := common.ResponseError{common.ERR4, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			resErr := common.ResponseError{common.ERR4, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())}
			return common.RespondError(resErr)
		}
		errCommit := json.Unmarshal(queryResponse.Value, commitResutl)
		if errCommit != nil {
			//convert JSON eror
			resErr := common.ResponseError{common.ERR6, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR6], err.Error(), common.GetLine())}
			return common.RespondError(resErr)
		}
		quorumList = append(quorumList, *quorumResutl)
	}

	if commitResutl.CommitID != "" && commitResutl.Status == "Verify" {
		resErr := common.ResponseError{common.ERR11, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR11], "", common.GetLine())}
		return common.RespondError(resErr)
	}

	count := 0
	var quorumIDList []string
	for _, quorum := range quorumList {
		if quorum.Status == "Verify" {
			for _, admin := range Adminlist {
				if quorum.AdminID == admin.AdminID {
					count = count + 1
					quorumIDList = append(quorumIDList, quorum.QuorumID)
				}
			}
		}
	}
	fmt.Printf("count:%v \n", count)
	fmt.Printf("quorumIDList:%v \n", quorumIDList)
	fmt.Printf("adminList:%v \n", Adminlist)

	if count < 3 {
		fmt.Printf("Not Enough quorum \n")
		resErr := common.ResponseError{common.ERR10, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR10], "[]", common.GetLine())}
		return common.RespondError(resErr)
	}
	CommitID := xid.New().String()
	fmt.Printf("CommitID %v\n", CommitID)

	err1 := Createdata(stub, models.COMMITTABLE, []string{CommitID}, &Commit{CommitID: string(CommitID), ProposalID: ProposalID, QuorumID: quorumIDList, Status: "Verify"})
	if err1 != nil {
		resErr := common.ResponseError{common.ERR6, fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR6], err1.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}
	resSuc := common.ResponseSuccess{common.SUCCESS, common.ResCodeDict[common.SUCCESS], CommitID}
	return common.RespondSuccess(resSuc)
}

// GetCommitByID
func (commit *Commit) GetCommitByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		//Invalid arguments
		resErr := common.ResponseError{common.ERR2, common.ResCodeDict[common.ERR2]}
		return common.RespondError(resErr)
	}
	DataID := args[0]
	res := GetDataByID(stub, DataID, commit, models.COMMITTABLE)
	return res
}

// GetAllCommit
func (commit *Commit) GetAllCommit(stub shim.ChaincodeStubInterface) pb.Response {
	res := GetAllData(stub, commit, models.COMMITTABLE)
	return res
}

// ------------------- //
