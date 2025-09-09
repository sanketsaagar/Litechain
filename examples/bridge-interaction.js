// LightChain L2 Bridge Interaction Example
// How to move assets between L1 Ethereum and L2 LightChain

const { ethers } = require('ethers');

// Configuration from your blockchain
const L1_RPC = 'https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY';
const L2_RPC = 'http://localhost:8545'; // Your LightChain L2
const AGGLAYER_BRIDGE = '0x...'; // AggLayer bridge contract address

class LightChainBridge {
    constructor() {
        this.l1Provider = new ethers.providers.JsonRpcProvider(L1_RPC);
        this.l2Provider = new ethers.providers.JsonRpcProvider(L2_RPC);
    }

    /**
     * Deposit ETH from L1 to L2
     * Uses AggLayer unified bridge (no wrapped tokens!)
     */
    async depositETHToL2(amount, l2RecipientAddress) {
        console.log(`🌉 Depositing ${amount} ETH from L1 to L2...`);
        
        // 1. Call L1 bridge contract
        const bridgeContract = new ethers.Contract(
            AGGLAYER_BRIDGE,
            ['function deposit(address recipient) payable'],
            this.l1Provider
        );
        
        const tx = await bridgeContract.deposit(l2RecipientAddress, {
            value: ethers.utils.parseEther(amount)
        });
        
        console.log(`📤 L1 transaction hash: ${tx.hash}`);
        
        // 2. Wait for L1 confirmation
        await tx.wait();
        console.log('✅ L1 deposit confirmed');
        
        // 3. Wait for AggLayer certificate (5-10 minutes)
        console.log('⏳ Waiting for AggLayer certificate...');
        await this.waitForL2Balance(l2RecipientAddress, amount);
        
        console.log('🎉 ETH successfully bridged to L2!');
    }

    /**
     * Withdraw ETH from L2 to L1
     * Uses pessimistic proofs for security
     */
    async withdrawETHToL1(amount, l1RecipientAddress) {
        console.log(`🏦 Withdrawing ${amount} ETH from L2 to L1...`);
        
        // 1. Initiate withdrawal on L2
        const l2Bridge = new ethers.Contract(
            '0x0000000000000000000000000000000000001001', // L2 bridge precompile
            ['function withdraw(address recipient, uint256 amount)'],
            this.l2Provider
        );
        
        const tx = await l2Bridge.withdraw(
            l1RecipientAddress,
            ethers.utils.parseEther(amount)
        );
        
        console.log(`📤 L2 withdrawal initiated: ${tx.hash}`);
        
        // 2. Wait for L2 confirmation
        await tx.wait();
        console.log('✅ L2 withdrawal confirmed');
        
        // 3. Wait for pessimistic proof verification (15-30 minutes)
        console.log('🔐 Waiting for pessimistic proof verification...');
        await this.waitForL1Release(l1RecipientAddress, amount);
        
        console.log('🎉 ETH successfully withdrawn to L1!');
    }

    /**
     * Send transaction on L2 (fast and cheap)
     */
    async sendL2Transaction(to, amount) {
        console.log(`⚡ Sending ${amount} ETH on L2 to ${to}...`);
        
        const tx = {
            to: to,
            value: ethers.utils.parseEther(amount),
            gasLimit: 21000,
            gasPrice: ethers.utils.parseUnits('1', 'gwei') // Very cheap on L2!
        };
        
        const response = await this.l2Provider.sendTransaction(tx);
        console.log(`📤 L2 transaction hash: ${response.hash}`);
        
        // Wait for L2 confirmation (~2 seconds)
        await response.wait();
        console.log('✅ L2 transaction confirmed in ~2 seconds!');
        
        return response.hash;
    }

    /**
     * Check account balance on L2
     */
    async getL2Balance(address) {
        const balance = await this.l2Provider.getBalance(address);
        return ethers.utils.formatEther(balance);
    }

    /**
     * Monitor bridge activity
     */
    async monitorBridgeActivity() {
        console.log('👀 Monitoring bridge activity...');
        
        // Listen for deposit events
        this.l1Provider.on('block', async (blockNumber) => {
            console.log(`📦 New L1 block: ${blockNumber}`);
            // Check for bridge deposits
        });
        
        // Listen for L2 certificate events
        this.l2Provider.on('block', async (blockNumber) => {
            console.log(`🧱 New L2 block: ${blockNumber}`);
            // Check for AggLayer certificates
        });
    }

    // Helper methods
    async waitForL2Balance(address, expectedAmount) {
        // Poll L2 balance until deposit appears
        for (let i = 0; i < 60; i++) { // 10 minute timeout
            const balance = await this.getL2Balance(address);
            if (parseFloat(balance) >= parseFloat(expectedAmount)) {
                return true;
            }
            await new Promise(resolve => setTimeout(resolve, 10000)); // Wait 10s
        }
        throw new Error('Timeout waiting for L2 deposit');
    }

    async waitForL1Release(address, expectedAmount) {
        // Poll L1 balance until withdrawal completes
        for (let i = 0; i < 180; i++) { // 30 minute timeout
            const balance = await this.l1Provider.getBalance(address);
            // Check if withdrawal completed
            await new Promise(resolve => setTimeout(resolve, 10000)); // Wait 10s
        }
    }
}

// Usage examples
async function main() {
    const bridge = new LightChainBridge();
    
    // Example 1: Move 1 ETH from L1 to L2
    // await bridge.depositETHToL2('1.0', '0x742A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F4');
    
    // Example 2: Fast L2 transaction
    await bridge.sendL2Transaction(
        '0x8B3A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F5',
        '0.1'
    );
    
    // Example 3: Withdraw back to L1
    // await bridge.withdrawETHToL1('0.5', '0x742A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F4');
    
    // Example 4: Monitor activity
    // await bridge.monitorBridgeActivity();
}

// Run example
if (require.main === module) {
    main().catch(console.error);
}

module.exports = LightChainBridge;
