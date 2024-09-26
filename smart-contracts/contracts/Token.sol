// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract LifePointToken is ERC20, Ownable {
    event BulkTransfer(address indexed token, uint256 totalAmount);

    constructor() ERC20("LifePoint", "LP") {
        _mint(msg.sender, 100000000 * 10 ** decimals()); // Mint 100,000,000 tokens to the owner
    }

    // Function for the owner to mint tokens
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }

    // Bulk transfer function
    function bulkTransfer(address[] calldata _to, uint256[] calldata _value) external {
        // Ensure the number of recipients does not exceed 255
        require(_to.length == _value.length, "Recipients and values length mismatch");
        require(_to.length <= 255, "Maximum 255 recipients allowed");

        // Start bulk transfer
        uint256 totalAmount = 0;
        for (uint8 i = 0; i < _to.length; i++) {
            _transfer(msg.sender, _to[i], _value[i]);
            totalAmount += _value[i];
        }

        // Emit event for the bulk transfer
        emit BulkTransfer(address(this), totalAmount);
    }
}