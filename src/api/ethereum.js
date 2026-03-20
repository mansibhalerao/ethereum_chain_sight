import { ethers } from 'ethers';

const RPC_URL = "https://eth.llamarpc.com";

const provider = new ethers.JsonRpcProvider(RPC_URL);

export default provider;