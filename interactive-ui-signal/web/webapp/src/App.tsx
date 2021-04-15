import React from 'react'
import './App.css'

import {ChakraProvider} from '@chakra-ui/react'
import {Header} from './Header'
import {Accounts} from './Accounts'
import {Provider as BusProvider} from 'react-bus'

interface AppProps {}

export const App: React.FC<AppProps> = ({}) => {
    return (
      <ChakraProvider>
        <BusProvider>
          <Header/>
          <Accounts/>
        </BusProvider>
      </ChakraProvider>
    )
}

export default App
