// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract MockUSDT is ERC20, Ownable {
    /**
     * @dev Constructor that mints an initial supply of tokens to a specified address.
     * @param initialRecipient The address to receive the initial supply.
     */
    constructor(address initialRecipient) ERC20("Mock USDT", "USDT") {
        require(initialRecipient != address(0), "Invalid address"); // Ensure valid address
        _mint(initialRecipient, 1_000_000 * 10**decimals()); // Mint initial supply to the specified address
    }

    /**
     * @dev Allows the contract owner to mint new tokens.
     * @param to The address to receive the minted tokens.
     * @param amount The amount of tokens to mint.
     */
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }
}
