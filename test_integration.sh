#!/bin/bash

echo "🧪 SAML SSO Database Integration Test"
echo "====================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if PostgreSQL is running
echo -e "\n${YELLOW}1. Checking PostgreSQL Database...${NC}"
if docker-compose ps | grep -q "saml-postgres.*Up.*healthy"; then
    echo -e "${GREEN}✅ PostgreSQL is running and healthy${NC}"
else
    echo -e "${RED}❌ PostgreSQL is not running. Starting it now...${NC}"
    docker-compose up -d
    sleep 5
fi

# Check database users
echo -e "\n${YELLOW}2. Checking Database Users...${NC}"
echo "Sample users in database:"
docker exec saml-postgres psql -U saml_user -d saml_sso -c "SELECT email, first_name, last_name, is_active FROM users;" 2>/dev/null

# Check if application is running
echo -e "\n${YELLOW}3. Checking SAML Application...${NC}"
if curl -s -I http://localhost:8080/debug >/dev/null 2>&1; then
    echo -e "${GREEN}✅ SAML application is running on port 8080${NC}"
else
    echo -e "${RED}❌ SAML application is not running${NC}"
    echo "Start it with: go run ."
    exit 1
fi

# Test debug endpoint
echo -e "\n${YELLOW}4. Testing Debug Endpoint...${NC}"
if curl -s http://localhost:8080/debug | grep -q "SAML Debug Information"; then
    echo -e "${GREEN}✅ Debug endpoint is working${NC}"
    echo "You can view detailed debug info at: http://localhost:8080/debug"
else
    echo -e "${RED}❌ Debug endpoint failed${NC}"
fi

# Test SAML redirect
echo -e "\n${YELLOW}5. Testing SAML SSO Flow...${NC}"
RESPONSE=$(curl -s -I http://localhost:8080)
if echo "$RESPONSE" | grep -q "Location.*home"; then
    echo -e "${GREEN}✅ Root redirect to /home is working${NC}"
    
    # Test home endpoint SAML redirect
    HOME_RESPONSE=$(curl -s -I http://localhost:8080/home)
    if echo "$HOME_RESPONSE" | grep -q "Location.*mocksaml.com"; then
        echo -e "${GREEN}✅ SAML redirect to IdP is working${NC}"
        echo "Redirect URL contains mocksaml.com"
    else
        echo -e "${RED}❌ SAML redirect failed${NC}"
        echo "Response: $HOME_RESPONSE"
    fi
else
    echo -e "${RED}❌ Root redirect failed${NC}"
    echo "Response: $RESPONSE"
fi

# Test metadata endpoint
echo -e "\n${YELLOW}6. Testing SAML Metadata Endpoint...${NC}"
if curl -s http://localhost:8080/saml/metadata | grep -q "EntityDescriptor"; then
    echo -e "${GREEN}✅ SAML metadata endpoint is working${NC}"
else
    echo -e "${RED}❌ SAML metadata endpoint failed${NC}"
fi

echo -e "\n${YELLOW}7. Manual Testing Instructions:${NC}"
echo "To test the complete flow:"
echo "1. Open browser and go to: http://localhost:8080"
echo "2. You'll be redirected to /home, then to mocksaml.com"
echo "3. Test with these users:"
echo -e "   ${GREEN}✅ jackson@example.com (should work - user exists and is active)${NC}"
echo -e "   ${GREEN}✅ test@example.com (should work - user exists and is active)${NC}"
echo -e "   ${GREEN}✅ admin@example.com (should work - user exists and is active)${NC}"
echo -e "   ${RED}❌ inactive@example.com (should fail - user exists but inactive)${NC}"
echo -e "   ${RED}❌ nonexistent@example.com (should fail - user doesn't exist)${NC}"

echo -e "\n${YELLOW}8. Debug and Troubleshooting:${NC}"
echo "Debug endpoint (no auth required): http://localhost:8080/debug"
echo "This shows:"
echo "  - SAML session status"
echo "  - Email extraction results"
echo "  - Database authorization status"
echo "  - All authorized users in database"

echo -e "\n${YELLOW}9. Database Management Commands:${NC}"
echo "Add new user:"
echo "  docker exec saml-postgres psql -U saml_user -d saml_sso -c \"INSERT INTO users (email, first_name, last_name) VALUES ('newuser@example.com', 'New', 'User');\""
echo ""
echo "Deactivate user:"
echo "  docker exec saml-postgres psql -U saml_user -d saml_sso -c \"UPDATE users SET is_active = false WHERE email = 'user@example.com';\""
echo ""
echo "View all users:"
echo "  docker exec saml-postgres psql -U saml_user -d saml_sso -c \"SELECT * FROM users;\""

echo -e "\n${GREEN}🎉 Integration test completed!${NC}"
echo "The SAML SSO application with database authentication is ready for testing."
echo -e "\n${YELLOW}Key URLs:${NC}"
echo "  🏠 Home: http://localhost:8080"
echo "  🔍 Debug: http://localhost:8080/debug"
echo "  📋 SAML Metadata: http://localhost:8080/saml/metadata" 