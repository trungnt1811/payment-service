// SPDX-License-Identifier: MIT
pragma solidity ^0.8.6;

interface ERC20 {
    function transferFrom(address sender, address recipient, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
}

contract Membership {
    address public owner;
    ERC20 public lifePointToken;
    
    uint256 public oneYearFee = 18000 * 10**18;
    uint256 public threeYearFee = 45000 * 10**18;

    enum MembershipDuration { ONE_YEAR, THREE_YEARS }

    struct Member {
        bool isMember;
        uint256 expiration;
    }

    mapping(address => Member) public members;
    mapping(uint64 => bool) public orderProcessed;

    event MembershipPurchased(address indexed user, uint256 amount, uint64 indexed orderId, uint8 duration);
    
    constructor(address _lifePointToken) {
        owner = msg.sender;
        lifePointToken = ERC20(_lifePointToken);
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can execute this");
        _;
    }

    function purchaseMembership(uint64 orderId, MembershipDuration duration) external {  // Changed to uint64
        require(!members[msg.sender].isMember, "Already a member");
        require(!orderProcessed[orderId], "Order already processed");

        uint256 membershipFee = getMembershipFee(duration);
        
        uint256 userBalance = lifePointToken.balanceOf(msg.sender);
        require(userBalance >= membershipFee, "Not enough Life Point tokens");
        
        bool success = lifePointToken.transferFrom(msg.sender, address(this), membershipFee);
        require(success, "Token transfer failed");

        members[msg.sender] = Member({
            isMember: true,
            expiration: block.timestamp + getMembershipDuration(duration)
        });

        orderProcessed[orderId] = true;

        emit MembershipPurchased(msg.sender, membershipFee, orderId, uint8(duration));
    }

    function withdrawTokens(uint256 amount) external onlyOwner {
        bool success = lifePointToken.transferFrom(address(this), owner, amount);
        require(success, "Token withdrawal failed");
    }

    function getMembershipFee(MembershipDuration duration) internal view returns (uint256) {
        if (duration == MembershipDuration.ONE_YEAR) {
            return oneYearFee;
        } else {
            return threeYearFee;
        }
    }

    function getMembershipDuration(MembershipDuration duration) internal pure returns (uint256) {
        if (duration == MembershipDuration.ONE_YEAR) {
            return 365 days;
        } else {
            return 1095 days;
        }
    }

    function isMembershipActive(address member) external view returns (bool) {
        return members[member].isMember && block.timestamp < members[member].expiration;
    }
}
