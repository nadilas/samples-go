import React, {useEffect, useState} from 'react'
import {
    Alert,
    AlertDescription,
    AlertIcon,
    AlertTitle,
    Box,
    Center,
    Table,
    TableCaption,
    Tbody,
    Td,
    Th,
    Thead,
    Tr
} from '@chakra-ui/react'
import {useListener} from 'react-bus'
import {AccountActions} from './AccountActions'
import axios from 'axios'

export type Execution = {
    workflow_id: string
    run_id: string
}

export type Account = {
    execution: Execution
    type: { name: string }
    execution_time: string
    start_time: string
    status: number
    memo: { [k: string]: any }
    plan: string
}

interface AccountsProps {
}

export const Accounts: React.FC<AccountsProps> = ({}) => {
    const [accounts, setAccounts] = useState<Account[]>([])
    const [loadError, setLoadError] = useState<string | undefined>(undefined)
    const updateAccountPlan = (accountId: string, plan: string) => {
        setAccounts(prevState => {
            const f = prevState.filter(a => a.execution.workflow_id === accountId && a.plan !== plan)
            if (f.length > 0) {
                f[0].plan = plan
                return [...prevState]
            }
            return prevState
        })
    }

    const fetchAccounts = () => {
        axios.get('http://localhost:8080/api/accounts')
            .then(resp => {
                if (!resp.data) {
                    return
                }
                const accounts = []
                if (resp.data.executions) {
                    resp.data.executions.forEach((e: Account) => {
                        e.plan = 'loading'
                        axios.get('http://localhost:8080/api/plan?account=' + e.execution.workflow_id)
                            .then(r => {
                                if (r.status === 200) {
                                    updateAccountPlan(e.execution.workflow_id, r.data.plan)
                                }
                            })
                    })
                    accounts.push(...resp.data.executions)
                }
                setAccounts(accounts)
                setLoadError(undefined)
            })
            .catch(err => {
                setLoadError(err.message)
            })
    }
    useEffect(fetchAccounts, [])

    useListener('fetchAccounts', fetchAccounts)

    return typeof loadError === 'undefined' ? (
        <Table variant="simple">
            <TableCaption>System accounts</TableCaption>
            <Thead>
                <Tr>
                    <Th>Name</Th>
                    <Th>Type</Th>
                    <Th>Actions</Th>
                </Tr>
            </Thead>
            <Tbody>
                {
                    accounts.map(a =>
                        (
                            <Tr key={a.execution.run_id}>
                                <Td>{a.execution.workflow_id}</Td>
                                <Td>
                                    {a.plan}
                                </Td>
                                <Td><AccountActions account={a} plan={a.plan}/></Td>
                            </Tr>
                        ))
                }
            </Tbody>
        </Table>
    ) : (
        <Center>
            <Box mt={16}>
                <Alert status="error">
                    <AlertIcon/>
                    <AlertTitle mr={2}>API error</AlertTitle>
                    <AlertDescription>{loadError}</AlertDescription>
                </Alert>
            </Box>
        </Center>
    )
}
