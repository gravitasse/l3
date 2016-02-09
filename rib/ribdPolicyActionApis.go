// ribdPolicyActionApis.go
package main

import (
	"ribd"
	"errors"
	"l3/rib/ribdCommonDefs"
	"utils/patriciaDB"
	"strconv"
)

var PolicyActionsDB = patriciaDB.NewTrie()
type RedistributeActionInfo struct {
	redistribute bool
	redistributeTargetProtocol int
}
type PolicyAction struct {
	name          string
	actionType int
	actionInfo interface {}
	policyList []string
	actionGetBulkInfo string
	localDBSliceIdx int
}
var localPolicyActionsDB []localDB
func updateLocalActionsDB(prefix patriciaDB.Prefix) {
    localDBRecord := localDB{prefix:prefix, isValid:true}
    if(localPolicyActionsDB == nil) {
		localPolicyActionsDB = make([]localDB, 0)
	} 
	localPolicyActionsDB = append(localPolicyActionsDB, localDBRecord)
}
func (m RouteServiceHandler) 	CreatePolicyDefinitionStmtRouteDispositionAction(cfg *ribd.PolicyDefinitionStmtRouteDispositionAction )(val bool, err error) {
	logger.Println("CreatePolicyDefinitionStmtRouteDispositionAction")
	policyAction := PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if(policyAction == nil) {
	   logger.Println("Defining a new policy action with name ", cfg.Name)
	   newPolicyAction := PolicyAction{name:cfg.Name,actionType:ribdCommonDefs.PolicyActionTypeRouteDisposition,actionInfo:cfg.RouteDisposition ,localDBSliceIdx:(len(localPolicyActionsDB))}
       newPolicyAction.actionGetBulkInfo =   cfg.RouteDisposition
		if ok := PolicyActionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyAction); ok != true {
			logger.Println(" return value not ok")
			return val, err
		}
	  updateLocalActionsDB(patriciaDB.Prefix(cfg.Name))
	} else {
		logger.Println("Duplicate action name")
		err = errors.New("Duplicate policy action definition")
		return val, err
	}
	return val, err
}

func (m RouteServiceHandler) CreatePolicyDefinitionStmtAdminDistanceAction(cfg *ribd.PolicyDefinitionStmtAdminDistanceAction) (val bool, err error) {
	logger.Println("CreatePolicyDefinitionStmtAdminDistanceAction")
	policyAction := PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if(policyAction == nil) {
	   logger.Println("Defining a new policy action with name ", cfg.Name)
	   newPolicyAction := PolicyAction{name:cfg.Name,actionType:ribdCommonDefs.PoilcyActionTypeSetAdminDistance,actionInfo:cfg.Value ,localDBSliceIdx:(len(localPolicyActionsDB))}
       newPolicyAction.actionGetBulkInfo =  "Set admin distance to value "+strconv.Itoa(int(cfg.Value))
		if ok := PolicyActionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyAction); ok != true {
			logger.Println(" return value not ok")
			return val, err
		}
	  updateLocalActionsDB(patriciaDB.Prefix(cfg.Name))
	} else {
		logger.Println("Duplicate action name")
		err = errors.New("Duplicate policy action definition")
		return val, err
	}
	return val, err
}

func (m RouteServiceHandler) CreatePolicyDefinitionStmtRedistributionAction(cfg *ribd.PolicyDefinitionStmtRedistributionAction) (val bool, err error) {
	logger.Println("CreatePolicyDefinitionStmtRedistributionAction")
	targetProtoType := -1

	policyAction := PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if(policyAction == nil) {
	   logger.Println("Defining a new policy action with name ", cfg.Name)
	   retProto,found := RouteProtocolTypeMapDB[cfg.RedistributeTargetProtocol]
	   if(found == false ) {
          logger.Println("Invalid target protocol type for redistribution %s ", cfg.RedistributeTargetProtocol)
		  return val,err
	   }
	   targetProtoType = retProto
	   logger.Printf("target protocol for RedistributeTargetProtocol %s is %d\n", cfg.RedistributeTargetProtocol, targetProtoType)
	   redistributeActionInfo := RedistributeActionInfo{redistribute:cfg.Redistribute, redistributeTargetProtocol:targetProtoType}
	   newPolicyAction := PolicyAction{name:cfg.Name,actionType:ribdCommonDefs.PolicyActionTypeRouteRedistribute,actionInfo:redistributeActionInfo ,localDBSliceIdx:(len(localPolicyActionsDB))}
       redistributeAction := " "
	   if redistributeActionInfo.redistribute == false {
          redistributeAction = "Don't"		
	   }
       newPolicyAction.actionGetBulkInfo = redistributeAction + " Redistribute to Target Protocol " + cfg.RedistributeTargetProtocol
		if ok := PolicyActionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyAction); ok != true {
			logger.Println(" return value not ok")
			return val, err
		}
	    updateLocalActionsDB(patriciaDB.Prefix(cfg.Name))
	} else {
		logger.Println("Duplicate action name")
		err = errors.New("Duplicate policy action definition")
		return val, err
	}
	return val, err
}
/*
func (m RouteServiceHandler) GetBulkPolicyDefinitionStmtRedistributionActions( fromIndex ribd.Int, rcount ribd.Int) (policyStmts *ribd.PolicyDefinitionStmtRedistributionActionsGetInfo, err error){
	logger.Println("getBulkPolicyDefinitionStmtRedistributionActions")
    var i, validCount, toIndex ribd.Int
	var tempNode []ribd.PolicyDefinitionStmtRedistributionAction = make ([]ribd.PolicyDefinitionStmtRedistributionAction, rcount)
	var nextNode *ribd.PolicyDefinitionStmtRedistributionAction
    var returnNodes []*ribd.PolicyDefinitionStmtRedistributionAction
	var returnGetInfo ribd.PolicyDefinitionStmtRedistributionActionsGetInfo
	i = 0
	policyActions := &returnGetInfo
	more := true
    if(localPolicyActionsDB == nil) {
		logger.Println("localPolicyActionsDB not initialized")
		return policyActions, err
	}
	for ;;i++ {
		logger.Printf("Fetching trie record for index %d\n", i+fromIndex)
		if(i+fromIndex >= ribd.Int(len(localPolicyActionsDB))) {
			logger.Println("All the policy actions fetched")
			more = false
			break
		}
		if(localPolicyActionsDB[i+fromIndex].isValid == false) {
			logger.Println("Invalid policy action statement")
			continue
		}
		if(validCount==rcount) {
			logger.Println("Enough policy actions fetched")
			break
		}
		logger.Printf("Fetching trie record for index %d and prefix %v\n", i+fromIndex, (localPolicyActionsDB[i+fromIndex].prefix))
		prefixNodeGet := PolicyActionsDB.Get(localPolicyActionsDB[i+fromIndex].prefix)
		if(prefixNodeGet != nil) {
			prefixNode := prefixNodeGet.(PolicyAction)
			if prefixNode.actionType != ribdCommonDefs.PolicyActionTypeRouteRedistribute {
				continue
			}
			nextNode = &tempNode[validCount]
		    nextNode.Name = prefixNode.name
			nextNode.RedistributeTargetProtocol = ReverseRouteProtoTypeMapDB[prefixNode.actionInfo.(int)]
			toIndex = ribd.Int(prefixNode.localDBSliceIdx)
			if(len(returnNodes) == 0){
				returnNodes = make([]*ribd.PolicyDefinitionStmtRedistributionAction, 0)
			}
			returnNodes = append(returnNodes, nextNode)
			validCount++
		}
	}
	logger.Printf("Returning %d list of policyActions", validCount)
	policyActions.PolicyDefinitionStmtRedistributionActionList = returnNodes
	policyActions.StartIdx = fromIndex
	policyActions.EndIdx = toIndex+1
	policyActions.More = more
	policyActions.Count = validCount
	return policyActions, err
}*/

func (m RouteServiceHandler) GetBulkPolicyDefinitionActionState( fromIndex ribd.Int, rcount ribd.Int) (policyActions *ribd.PolicyDefinitionActionStateGetInfo, err error){//(routes []*ribd.Routes, err error) {
	logger.Println("GetBulkPolicyDefinitionActionState")
    var i, validCount, toIndex ribd.Int
	var tempNode []ribd.PolicyDefinitionActionState = make ([]ribd.PolicyDefinitionActionState, rcount)
	var nextNode *ribd.PolicyDefinitionActionState
    var returnNodes []*ribd.PolicyDefinitionActionState
	var returnGetInfo ribd.PolicyDefinitionActionStateGetInfo
	i = 0
	policyActions = &returnGetInfo
	more := true
    if(localPolicyActionsDB == nil) {
		logger.Println("PolicyDefinitionStmtMatchProtocolActionGetInfo not initialized")
		return policyActions, err
	}
	for ;;i++ {
		logger.Printf("Fetching trie record for index %d\n", i+fromIndex)
		if(i+fromIndex >= ribd.Int(len(localPolicyActionsDB))) {
			logger.Println("All the policy Actions fetched")
			more = false
			break
		}
		if(localPolicyActionsDB[i+fromIndex].isValid == false) {
			logger.Println("Invalid policy Action statement")
			continue
		}
		if(validCount==rcount) {
			logger.Println("Enough policy Actions fetched")
			break
		}
		logger.Printf("Fetching trie record for index %d and prefix %v\n", i+fromIndex, (localPolicyActionsDB[i+fromIndex].prefix))
		prefixNodeGet := PolicyActionsDB.Get(localPolicyActionsDB[i+fromIndex].prefix)
		if(prefixNodeGet != nil) {
			prefixNode := prefixNodeGet.(PolicyAction)
			nextNode = &tempNode[validCount]
		    nextNode.Name = prefixNode.name
			nextNode.ActionInfo = prefixNode.actionGetBulkInfo
            if prefixNode.policyList != nil {
				nextNode.PolicyList = make([]string,0)
			}
			for idx := 0;idx < len(prefixNode.policyList);idx++ {
				nextNode.PolicyList = append(nextNode.PolicyList, prefixNode.policyList[idx])
			}
 			toIndex = ribd.Int(prefixNode.localDBSliceIdx)
			if(len(returnNodes) == 0){
				returnNodes = make([]*ribd.PolicyDefinitionActionState, 0)
			}
			returnNodes = append(returnNodes, nextNode)
			validCount++
		}
	}
	logger.Printf("Returning %d list of policyActions", validCount)
	policyActions.PolicyDefinitionActionStateList = returnNodes
	policyActions.StartIdx = fromIndex
	policyActions.EndIdx = toIndex+1
	policyActions.More = more
	policyActions.Count = validCount
	return policyActions, err
}
