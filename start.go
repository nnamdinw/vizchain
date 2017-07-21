/*
Copyright 2016 IBM

Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Licensed Materials - Property of IBM
© Copyright IBM Corp. 2016
*/




/*

Hierarchy :
	AccountHolder
		AssetID[](footage::vID)

	Footage
		vID
		Owner
		Frames[] video_frame

	video_frame
		hash
		timecode 






*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}


type video_frame struct {
	Hash  string    `json:"Hash"`
	Timecode string      `json:"Timecode"`
}

type footage struct {
	vID     string  `json:"vID"`
	Owner    Account  `json:"owner"`
	Frames    []video_frame `json:"frames"`
}

type Account struct {
	ID          string  `json:"id"`
	Name      string  `json:"name"`
	AssetsIds   []string `json:"assetIds"`
}



func (t *SimpleChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Creating account")

	// Obtain the username to associate with the account
	if len(args) != 2 {
		fmt.Println("Error obtaining username")
		return nil, errors.New("createAccount accepts a single username argument")
	}
	username := args[0]
	fullname := args[1]

	// Build an account object for the user
	var assetIds []string
	var account = Account{ID: username, Name: fullname, AssetsIds: assetIds}
	accountBytes, err := json.Marshal(&account)
	if err != nil {
		fmt.Println("error creating account" + account.ID)
		return nil, errors.New("Error creating account " + account.ID)
	}

	fmt.Println("Attempting to get state of any existing account for " + account.ID)
	existingBytes, err := stub.GetState(account.ID)
	if err == nil {

		var company Account
		err = json.Unmarshal(existingBytes, &company)
		if err != nil {
			fmt.Println("Error unmarshalling account " + account.ID + "\n--->: " + err.Error())

			if strings.Contains(err.Error(), "unexpected end") {
				fmt.Println("No data means existing account found for " + account.ID + ", initializing account.")
				err = stub.PutState(account.ID, accountBytes)

				if err == nil {
					fmt.Println("created account" + account.ID)
					return nil, nil
				} else {
					fmt.Println("failed to create initialize account for " + account.ID)
					return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
				}
			} else {
				return nil, errors.New("Error unmarshalling existing account " + account.ID)
			}
		} else {
			fmt.Println("Account already exists for " + account.ID + " " + company.ID)
			return nil, errors.New("Can't reinitialize existing user " + account.ID)
		}
	} else {

		fmt.Println("No existing account found for " + account.ID + ", initializing account.")
		err = stub.PutState( account.ID, accountBytes)

		if err == nil {
			fmt.Println("created account" + account.ID)
			return nil, nil
		} else {
			fmt.Println("failed to create initialize account for " + account.ID)
			return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
		}

	}

}


func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Init firing. Function will be ignored: " + function)

	// Initialize the collection of commercial paper keys
	fmt.Println("Initializing paper keys collection")
	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	err := stub.PutState("FootageKeys", blankBytes)
	if err != nil {
		fmt.Println("Failed to initialize paper key collection")
	}

	fmt.Println("Initialization complete")
	return "GoAway", nil
}

func (t *SimpleChaincode) createNewFootage(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Creating commercial paper")

	/*		0
		json
	  	{
			"ticker":  "string",
			"par": 0.00,
			"qty": 10,
			"discount": 7.5,
			"maturity": 30,
			"owners": [ // This one is not required
				{
					"company": "company1",
					"quantity": 5
				},
				{
					"company": "company3",
					"quantity": 3
				},
				{
					"company": "company4",
					"quantity": 2
				}
			],				
			"issuer":"company2",
			"issueDate":"1456161763790"  (current time in milliseconds as a string)

		}
	*/
	//need minimum two arg , two arg means creating empty footage, 4 arg means appending frame to existing footage.. 
	if len(args) != 2 {
		fmt.Println("error invalid arguments")
		return nil, errors.New("Incorrect number of arguments. Expecting accout ID of footage creator and videoID") //if adding new frames to existing footage, pass me existing ID - else - ill handle a brand new vID
	}



	var newfootage footage
	var err error
	var account Account
	var argCheck = false
	fmt.Println("Unmarshalling Footage")
	err = json.Unmarshal([]byte(args[0]), &account)
	if err != nil {
		fmt.Println("error invalid footage issue")
		return nil, errors.New("Invalid footage issue")
	}
	newfootage.Owner = account
	newfootage.vID = args[1]
	account.AssetsIds = append(account.AssetsIds, newfootage.vID)

	var videoframe video_frame

	if(len(args) == 4){
		videoframe.Hash = args[2]
		videoframe.Timecode = args[3]
	} else {
		videoframe.Hash = ""
		videoframe.Timecode = ""
		argCheck = true
	}
	newfootage.Frames = append(newfootage.Frames, videoframe)

	fmt.Println("Getting State on footage " + newfootage.vID)
	cpRxBytes, err := stub.GetState(newfootage.vID)
	if cpRxBytes == nil {
		fmt.Println("vID does not exist, creating it") //new footage
		cpBytes, err := json.Marshal(&newfootage)
		if err != nil {
			fmt.Println("Error marshalling foortage")
			return nil, errors.New("Error issuing footage")
		}
		err = stub.PutState(newfootage.vID, cpBytes)
		if err != nil {
			fmt.Println("Error issuing footage")
			return nil, errors.New("Error issuing footage")
		}

		fmt.Println("Marshalling account bytes to write")
		accountBytesToWrite, err := json.Marshal(&account)
		if err != nil {
			fmt.Println("Error marshalling account")
			return nil, errors.New("Error issuing footage")
		}
		err = stub.PutState( newfootage.Owner.ID, accountBytesToWrite)
		if err != nil {
			fmt.Println("Error putting state on accountBytesToWrite")
			return nil, errors.New("Error issuing commercial paper")
		}


		// Update the paper keys by adding the new key
		fmt.Println("Getting Footage Keys")
		keysBytes, err := stub.GetState("footageKeys") //get existing account's footage
		if err != nil {
			fmt.Println("Error retrieving footage keys")
			return nil, errors.New("Error retrieving footage keys")
		}
		var keys []string
		err = json.Unmarshal(keysBytes, &keys)
		if err != nil {
			fmt.Println("Error unmarshel keys")
			return nil, errors.New("Error unmarshalling Footage keys ")
		}

		fmt.Println("Appending the new key to Paper Keys")
		foundKey := false
		for _, key := range keys {
			if key == newfootage.vID {
				foundKey = true
			}
		}
		if foundKey == false { 
			//new footage case
			keys = append(keys, newfootage.vID)
			keysBytesToWrite, err := json.Marshal(&keys)
			if err != nil {
				fmt.Println("Error marshalling footage")
				return nil, errors.New("Error marshalling the keys")
			}
			fmt.Println("Put state on Footage Keys")
			err = stub.PutState("footageKeys", keysBytesToWrite)
			if err != nil {
				fmt.Println("Error writting keys back")
				return nil, errors.New("Error writing the keys back")
			}
		}

		fmt.Println("Create footage paper %+v\n", newfootage)
		return nil, nil
	} else {
		fmt.Println("Footage exists, appending frames!")

		var cprx footage
		fmt.Println("Unmarshalling footage " + newfootage.vID)
		err = json.Unmarshal(cpRxBytes, &cprx)
		if err != nil {
			fmt.Println("Error unmarshalling footage " + newfootage.vID)
			return nil, errors.New("Error unmarshalling footage " + newfootage.vID)
		}
/*
		for key, val := range cprx.Frames {
			if val.Company == cp.Issuer {
				cprx.Owners[key].Quantity += cp.Qty
				break
			}	
		}
*/
		if(argCheck == false){
			return nil, errors.New("Error, Footage ID exists but no new frame data provided")
		}

		cprx.Frames = append(cprx.Frames,videoframe)
		cpWriteBytes, err := json.Marshal(&cprx)
		if err != nil {
			fmt.Println("Error marshalling footage")
			return nil, errors.New("Error issuing footage")
		}
		err = stub.PutState(newfootage.vID, cpWriteBytes)
		if err != nil {
			fmt.Println("Error issuing footage")
			return nil, errors.New("Error issuing footage")
		}

		fmt.Println("Updated footage %+v\n", cprx)
		return nil, nil
	}
}

func getAllFootage(stub shim.ChaincodeStubInterface) ([]footage, error) {

	var allFootage []footage

	// Get list of all the keys
	keysBytes, err := stub.GetState("footageKeys") //get keys of this account's footages
	if err != nil {
		fmt.Println("Error retrieving footage keys")
		return nil, errors.New("Error retrieving footages")
	}
	var keys []string
	err = json.Unmarshal(keysBytes, &keys)
	if err != nil {
		fmt.Println("Error unmarshalling footage keys")
		return nil, errors.New("Error unmarshalling footage keys")
	}

	// Get all the cps
	for _, value := range keys {
		cpBytes, err := stub.GetState(value)

		var feet footage
		err = json.Unmarshal(cpBytes, &feet) //gross
		if err != nil {
			fmt.Println("Error retrieving footage " + value)
			return nil, errors.New("Error retrieving cp " + value)
		}

		fmt.Println("Appending CP" + value)
		allFootage = append(allFootage, feet)
	}

	return allFootage, nil
}

func getFootage(cpid string, stub shim.ChaincodeStubInterface) (footage, error) {
	var feet footage //here feet is a single footage object

	cpBytes, err := stub.GetState(cpid) //here 'cpid' refers to vID field of footage
	if err != nil {
		fmt.Println("Error retrieving footage " + cpid)
		return feet, errors.New("Error retrieving footage " + cpid)
	}

	err = json.Unmarshal(cpBytes, &feet)
	if err != nil {
		fmt.Println("Error unmarshalling footage " + cpid)
		return feet, errors.New("Error unmarshalling cp " + cpid)
	}

	return feet, nil
}

func getAccount(companyID string, stub shim.ChaincodeStubInterface) (Account, error) {
	var shooter Account  //shooter of footgage / account holder
	companyBytes, err := stub.GetState(companyID)
	if err != nil {
		fmt.Println("Account not found " + companyID)
		return shooter, errors.New("Account not found " + companyID)
	}

	err = json.Unmarshal(companyBytes, &shooter)
	if err != nil {
		fmt.Println("Error unmarshalling account " + companyID + "\n err:" + err.Error())
		return shooter, errors.New("Error unmarshalling account " + companyID)
	}

	return shooter, nil
}



func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Query running. Function: " + function)

	if function == "GetAllFootage" {
		fmt.Println("Getting all Footages")
		allCPs, err := getAllFootage(stub)
		if err != nil {
			fmt.Println("Error from getAllFootage")
			return nil, err
		} else {
			allCPsBytes, err1 := json.Marshal(&allCPs)
			if err1 != nil {
				fmt.Println("Error marshalling allcps")
				return nil, err1
			}
			fmt.Println("All success, returning allcps")
			return allCPsBytes, nil
		}
	} else if function == "GetFootage" {
		fmt.Println("Getting particular footage")
		cp, err := getFootage(args[0], stub)
		if err != nil {
			fmt.Println("Error Getting particular footage")
			return nil, err
		} else {
			cpBytes, err1 := json.Marshal(&cp)
			if err1 != nil {
				fmt.Println("Error marshalling the footage")
				return nil, err1
			}
			fmt.Println("All success, returning the footage")
			return cpBytes, nil
		}
	} else if function == "GetAccount" {
		fmt.Println("Getting the account")
		company, err := getAccount(args[0], stub)
		if err != nil {
			fmt.Println("Error from getAccount")
			return nil, err
		} else {
			companyBytes, err1 := json.Marshal(&company)
			if err1 != nil {
				fmt.Println("Error marshalling the account")
				return nil, err1
			}
			fmt.Println("All success, returning the account")
			return companyBytes, nil
		}
	} else {
		fmt.Println("Generic Query call")
		bytes, err := stub.GetState(args[0])

		if err != nil {
			fmt.Println("Some error happenend: " + err.Error())
			return nil, err
		}

		fmt.Println("All success, returning from generic")
		return bytes, nil
	}
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Invoke running. Function: " + function)

	if function == "craeteNewFootage" {
		return t.createNewFootage(stub, args) //we use this to create brand new footage and add to existing footage
	} else if function == "createAccount" {
		return t.createAccount(stub, args)
	}

	return nil, errors.New("Received unknown function invocation: " + function)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode: %s", err)
	}
}

