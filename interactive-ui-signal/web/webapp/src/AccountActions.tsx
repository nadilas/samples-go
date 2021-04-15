import React, {useState} from 'react'
import {Button, ButtonGroup, useToast} from '@chakra-ui/react'
import {Account} from './Accounts'
import axios from 'axios'
import {useBus} from 'react-bus'

interface AccountActionsProps {
  account: Account
  plan: string
}

export const AccountActions: React.FC<AccountActionsProps> = ({account, plan}) => {
  const planTarget = plan === 'trial' ? 'premium' : 'trial'
  const toast = useToast()
  const bus = useBus()
  return (
    <ButtonGroup size="sm" isAttached variant="outline">
      <Button
        mr="-px"
        onClick={() => {
          axios.post(
            'http://localhost:8080/api/account/upgrade',
            {account: account.execution.workflow_id, actor: account.execution.workflow_id, to: planTarget }
          ).then(r => {
            if (r.status === 200) {
              if (r.data.error) {
                toast({
                  title: "Account upgrade failed.",
                  description: r.data.error,
                  status: "error",
                  duration: 5000,
                  isClosable: true,
                })
              } else {
                // reload UI
                setTimeout(function() {
                  bus.emit('fetchAccounts', undefined)
                }, 1000)
                // inform user
                toast({
                  title: "Account changed.",
                  description: "Account successfully upgraded to premium.",
                  status: "success",
                  duration: 3000,
                  isClosable: true,
                })
              }
            }
          }).catch(err => {
            toast({
              title: "Account upgrade failed.",
              description: err.message,
              status: "error",
              duration: 5000,
              isClosable: true,
            })
          })
        }}
      >Move to {planTarget}</Button>
      <Button
        mr="-px"
        colorScheme="red"
        onClick={() =>{
          axios.post(
            'http://localhost:8080/api/account/delete',
            {account: account.execution.workflow_id, actor: account.execution.workflow_id }
          ).then(r => {
            if (r.status === 200) {
              if (r.data.error) {
                toast({
                  title: "Account upgrade failed.",
                  description: r.data.error,
                  status: "error",
                  duration: 5000,
                  isClosable: true,
                })
              } else {
                // reload UI
                setTimeout(function() {
                  bus.emit('fetchAccounts', undefined)
                }, 1000)
                // inform user
                toast({
                  title: "Account deleted.",
                  description: "Account successfully deleted.",
                  status: "success",
                  duration: 3000,
                  isClosable: true,
                })
              }
            }
          }).catch(err => {
            toast({
              title: "Account deletion failed.",
              description: err.message,
              status: "error",
              duration: 5000,
              isClosable: true,
            })
          })
        }}
      >Delete</Button>
    </ButtonGroup>
  )
}
