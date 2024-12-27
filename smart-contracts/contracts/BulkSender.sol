// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract BulkSender {
    // Events for logging transfers
    event ERC20Transfer(address indexed token, address indexed from, address indexed to, uint256 amount);
    event NativeTransfer(address indexed from, address indexed to, uint256 amount);

    error ArraysLengthMismatch(uint256 recipientsLength, uint256 amountsLength);

    /**
     * @notice Transfers tokens or Ether to multiple recipients in a single transaction.
     * @param recipients An array of recipient addresses.
     * @param amounts An array of amounts to transfer to each recipient.
     * @param tokenAddress The address of the ERC-20 token to transfer, or address(0) for Ether transfers.
     */
    function bulkTransfer(
        address[] calldata recipients,
        uint256[] calldata amounts,
        address tokenAddress
    ) external payable {
        // Check that the recipients and amounts arrays have the same length
        if (recipients.length != amounts.length) {
            revert ArraysLengthMismatch(recipients.length, amounts.length);
        }

        // If tokenAddress is address(0), perform Ether transfers
        if (tokenAddress == address(0)) {
            _bulkTransferNative(recipients, amounts);
        } else {
            // Otherwise, perform ERC-20 token transfers
            _bulkTransferERC20(recipients, amounts, tokenAddress);
        }

        // Return any leftover Ether back to the sender
        if (address(this).balance > 0) {
            payable(msg.sender).transfer(address(this).balance);
        }
    }

    /**
     * @notice Internal function to transfer Ether to multiple recipients.
     * @param recipients An array of recipient addresses.
     * @param amounts An array of amounts to transfer to each recipient.
     */
    function _bulkTransferNative(address[] calldata recipients, uint256[] calldata amounts) private {
        for (uint256 i = 0; i < recipients.length; i++) {
            address recipient = recipients[i];
            uint256 amount = amounts[i];
            payable(recipient).transfer(amount);
            emit NativeTransfer(msg.sender, recipient, amount);
        }
    }

    /**
     * @notice Internal function to transfer ERC-20 tokens to multiple recipients.
     * @param recipients An array of recipient addresses.
     * @param amounts An array of amounts to transfer to each recipient.
     * @param tokenAddress The address of the ERC-20 token to transfer.
     */
    function _bulkTransferERC20(
        address[] calldata recipients,
        uint256[] calldata amounts,
        address tokenAddress
    ) private {
        IERC20 token = IERC20(tokenAddress);
        for (uint256 i = 0; i < recipients.length; i++) {
            address recipient = recipients[i];
            uint256 amount = amounts[i];
            token.transferFrom(msg.sender, recipient, amount);
            emit ERC20Transfer(tokenAddress, msg.sender, recipient, amount);
        }
    }
}
