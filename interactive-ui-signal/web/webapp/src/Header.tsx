import React from 'react'
import {
  Box,
  Button,
  Flex,
  Heading,
  Modal, ModalBody,
  ModalCloseButton,
  ModalContent, ModalFooter,
  ModalHeader,
  ModalOverlay,
  Spacer, useDisclosure
} from '@chakra-ui/react'
import {Formik, Form} from 'formik'
import {InputField} from './InputField'
import {useBus} from 'react-bus'
import axios from 'axios'

interface HeaderProps {
}

export const Header: React.FC<HeaderProps> = ({}) => {
  const {isOpen, onOpen, onClose} = useDisclosure()
  const bus = useBus()
  return (
    <>
      <Modal
        isOpen={isOpen}
        onClose={onClose}
      >
        <ModalOverlay/>
        <ModalContent>
          <ModalHeader>Create your account</ModalHeader>
          <ModalCloseButton/>
          <Formik
            initialValues={{account: ''}}
            onSubmit={async (values, {setSubmitting, setErrors}) => {
              if (values.account === '') {
                setErrors({
                  account: 'Account is required'
                })
                return
              }
              const account = { name: values.account }
              try {
                const response = await axios.post('http://localhost:8080/api/accounts', account)
                if (response.status === 200) {
                  setTimeout(function() {
                    bus.emit('fetchAccounts', undefined)
                  }, 500)
                } else {
                  setErrors({
                    account: response.statusText
                  })
                }
              } catch (e) {
                setErrors({
                  account: e.message
                })
              }
              return onClose()
            }}
          >
            {({isSubmitting}) => (
              <Form>
                <ModalBody pb={6}>
                  <InputField name="account" placeholder="account" label="Account name"/>
                </ModalBody>

                <ModalFooter>
                  <Button
                    mr={2}
                    colorScheme="teal"
                    isLoading={isSubmitting}
                    type="submit"
                  >
                    Save
                  </Button>
                  <Button onClick={onClose}>Cancel</Button>
                </ModalFooter>
              </Form>
            )}
          </Formik>
        </ModalContent>
      </Modal>

      <Flex p="2" color="white" bgColor="teal">
        <Box p="2">
          <Heading size="md">Interactive UI signal - web client</Heading>
        </Box>
        <Spacer/>
        <Box>
          <Button colorScheme="teal" onClick={onOpen}>New account</Button>
        </Box>
      </Flex>
    </>
  )
}
