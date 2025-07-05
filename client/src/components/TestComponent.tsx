import { Box, useColorModeValue } from "@chakra-ui/react";

export default function TestComponent() {
  const bg = useColorModeValue("white", "gray.900");

  return (
    <Box p={4} bg={bg}>
      Hello Chakra
    </Box>
  );
}